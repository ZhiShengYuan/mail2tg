package fetcher

import (
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/kexi/mail-to-tg/internal/fetcher/gmail"
	"github.com/kexi/mail-to-tg/internal/fetcher/imap"
	"github.com/kexi/mail-to-tg/internal/parser"
	"github.com/kexi/mail-to-tg/internal/queue"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/kexi/mail-to-tg/pkg/models"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	db              *storage.MariaDB
	publisher       *queue.Publisher
	parser          *parser.Parser
	cfg             *config.Config
	pollers         map[string]*imap.Poller
	gmailClients    map[string]*gmail.Client
	mu              sync.RWMutex
	stopped         bool
}

func NewManager(
	db *storage.MariaDB,
	publisher *queue.Publisher,
	cfg *config.Config,
) (*Manager, error) {
	// Decode encryption key
	encryptionKey, err := base64.StdEncoding.DecodeString(cfg.Security.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	emailParser := parser.NewParser(encryptionKey, cfg.Storage.AttachmentsPath)

	return &Manager{
		db:           db,
		publisher:    publisher,
		parser:       emailParser,
		cfg:          cfg,
		pollers:      make(map[string]*imap.Poller),
		gmailClients: make(map[string]*gmail.Client),
	}, nil
}

func (m *Manager) Start() error {
	log.Info().Msg("Starting fetch manager")

	// Load active accounts
	accounts, err := m.db.GetActiveEmailAccounts()
	if err != nil {
		return fmt.Errorf("failed to load active accounts: %w", err)
	}

	log.Info().Int("count", len(accounts)).Msg("Loaded active email accounts")

	// Start fetchers for each account
	for _, account := range accounts {
		if err := m.startFetcherForAccount(account.ID); err != nil {
			log.Error().
				Err(err).
				Str("account_id", account.ID).
				Str("email", account.EmailAddress).
				Msg("Failed to start fetcher for account")
		}
	}

	// Periodically check for new or updated accounts
	go m.watchAccounts()

	return nil
}

func (m *Manager) Stop() {
	log.Info().Msg("Stopping fetch manager")
	m.stopped = true

	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop all pollers
	for _, poller := range m.pollers {
		poller.Stop()
	}

	// Stop all Gmail clients
	for _, client := range m.gmailClients {
		client.Stop()
	}
}

func (m *Manager) watchAccounts() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for !m.stopped {
		<-ticker.C

		accounts, err := m.db.GetActiveEmailAccounts()
		if err != nil {
			log.Error().Err(err).Msg("Failed to load accounts")
			continue
		}

		m.mu.RLock()
		existingPollers := make(map[string]bool)
		for id := range m.pollers {
			existingPollers[id] = true
		}
		existingGmail := make(map[string]bool)
		for id := range m.gmailClients {
			existingGmail[id] = true
		}
		m.mu.RUnlock()

		// Start fetchers for new accounts
		for _, account := range accounts {
			if account.Provider == "imap" && !existingPollers[account.ID] {
				log.Info().
					Str("account_id", account.ID).
					Str("email", account.EmailAddress).
					Msg("Starting fetcher for new IMAP account")
				m.startFetcherForAccount(account.ID)
			} else if account.Provider == "gmail" && !existingGmail[account.ID] {
				log.Info().
					Str("account_id", account.ID).
					Str("email", account.EmailAddress).
					Msg("Starting fetcher for new Gmail account")
				m.startFetcherForAccount(account.ID)
			}
		}
	}
}

func (m *Manager) startFetcherForAccount(accountID string) error {
	account, err := m.db.GetEmailAccountByID(accountID)
	if err != nil {
		return fmt.Errorf("failed to load account: %w", err)
	}
	if account == nil {
		return fmt.Errorf("account not found")
	}

	if !account.IsActive {
		return fmt.Errorf("account is not active")
	}

	switch account.Provider {
	case "imap":
		return m.startIMAPPoller(account)
	case "gmail":
		return m.startGmailClient(account)
	default:
		return fmt.Errorf("unsupported provider: %s", account.Provider)
	}
}

func (m *Manager) startIMAPPoller(account *models.EmailAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.pollers[account.ID]; exists {
		return nil
	}

	interval := time.Duration(m.cfg.MailFetcher.IMAPPollInterval) * time.Second
	poller := imap.NewPoller(account, m.db, m.publisher, m.parser, interval)

	m.pollers[account.ID] = poller

	// Start in goroutine
	go func() {
		if err := poller.Start(); err != nil {
			log.Error().
				Err(err).
				Str("account_id", account.ID).
				Msg("IMAP poller stopped with error")
		}
	}()

	log.Info().
		Str("account_id", account.ID).
		Str("email", account.EmailAddress).
		Msg("Started IMAP poller")

	return nil
}

func (m *Manager) startGmailClient(account *models.EmailAccount) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running
	if _, exists := m.gmailClients[account.ID]; exists {
		return nil
	}

	client, err := gmail.NewClient(account, m.db, m.publisher, m.parser, &m.cfg.MailFetcher.Gmail)
	if err != nil {
		return fmt.Errorf("failed to create Gmail client: %w", err)
	}

	m.gmailClients[account.ID] = client

	// Start in goroutine
	go func() {
		if err := client.Start(); err != nil {
			log.Error().
				Err(err).
				Str("account_id", account.ID).
				Msg("Gmail client stopped with error")
		}
	}()

	log.Info().
		Str("account_id", account.ID).
		Str("email", account.EmailAddress).
		Msg("Started Gmail client")

	return nil
}

func (m *Manager) StopFetcherForAccount(accountID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if poller, exists := m.pollers[accountID]; exists {
		poller.Stop()
		delete(m.pollers, accountID)
		log.Info().Str("account_id", accountID).Msg("Stopped IMAP poller")
	}

	if client, exists := m.gmailClients[accountID]; exists {
		client.Stop()
		delete(m.gmailClients, accountID)
		log.Info().Str("account_id", accountID).Msg("Stopped Gmail client")
	}
}

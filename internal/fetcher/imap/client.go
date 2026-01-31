package imap

import (
	"fmt"
	"io"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/rs/zerolog/log"
)

type Client struct {
	server   string
	port     int
	username string
	password string
}

func NewClient(server string, port int, username, password string) *Client {
	return &Client{
		server:   server,
		port:     port,
		username: username,
		password: password,
	}
}

type Message struct {
	UID         uint32
	MessageID   string
	Subject     string
	From        []*imap.Address
	To          []*imap.Address
	Date        string
	Body        string
	HTMLBody    string
	Flags       []string
	InReplyTo   string
	References  string
	RawMessage  []byte
}

func (c *Client) FetchUnread() ([]*Message, error) {
	// Connect with TLS
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", c.server, c.port), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP server: %w", err)
	}
	defer imapClient.Logout()

	log.Debug().
		Str("server", c.server).
		Int("port", c.port).
		Str("username", c.username).
		Msg("Connected to IMAP server")

	// Login
	if err := imapClient.Login(c.username, c.password); err != nil {
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	// Select INBOX
	mbox, err := imapClient.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("failed to select INBOX: %w", err)
	}

	log.Debug().Uint32("messages", mbox.Messages).Msg("Selected INBOX")

	if mbox.Messages == 0 {
		return []*Message{}, nil
	}

	// Search for unseen messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	uids, err := imapClient.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	if len(uids) == 0 {
		log.Debug().Msg("No unread messages found")
		return []*Message{}, nil
	}

	log.Debug().Int("count", len(uids)).Msg("Found unread messages")

	// Fetch messages
	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{
		imap.FetchEnvelope,
		imap.FetchFlags,
		imap.FetchUid,
		section.FetchItem(),
	}

	go func() {
		done <- imapClient.Fetch(seqset, items, messages)
	}()

	var result []*Message
	for msg := range messages {
		if msg == nil {
			continue
		}

		r := msg.GetBody(section)
		if r == nil {
			log.Warn().Uint32("uid", msg.Uid).Msg("Message body is nil")
			continue
		}

		rawBody, err := io.ReadAll(r)
		if err != nil {
			log.Error().Err(err).Uint32("uid", msg.Uid).Msg("Failed to read message body")
			continue
		}

		message := &Message{
			UID:        msg.Uid,
			RawMessage: rawBody,
			Flags:      msg.Flags,
		}

		if msg.Envelope != nil {
			message.Subject = msg.Envelope.Subject
			message.From = msg.Envelope.From
			message.To = msg.Envelope.To
			message.Date = msg.Envelope.Date.String()
			message.MessageID = msg.Envelope.MessageId
			message.InReplyTo = msg.Envelope.InReplyTo
		}

		result = append(result, message)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	return result, nil
}

func (c *Client) MarkAsSeen(uid uint32) error {
	imapClient, err := client.DialTLS(fmt.Sprintf("%s:%d", c.server, c.port), nil)
	if err != nil {
		return err
	}
	defer imapClient.Logout()

	if err := imapClient.Login(c.username, c.password); err != nil {
		return err
	}

	if _, err := imapClient.Select("INBOX", false); err != nil {
		return err
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}

	return imapClient.Store(seqset, item, flags, nil)
}

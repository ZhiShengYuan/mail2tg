package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kexi/mail-to-tg/internal/storage"
	"github.com/kexi/mail-to-tg/pkg/config"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router *gin.Engine
	cfg    *config.WebConfig
	db     *storage.MariaDB
}

func NewServer(cfg *config.WebConfig, db *storage.MariaDB) *Server {
	if cfg.TLSEnabled {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("internal/web/templates/*.html")

	s := &Server{
		router: router,
		cfg:    cfg,
		db:     db,
	}

	s.setupRoutes()

	return s
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.handleHealth)
	s.router.GET("/email/:token", s.handleViewEmail)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	log.Info().
		Str("addr", addr).
		Bool("tls", s.cfg.TLSEnabled).
		Msg("Starting web server")

	if s.cfg.TLSEnabled {
		return s.router.RunTLS(addr, s.cfg.TLSCert, s.cfg.TLSKey)
	}

	return s.router.Run(addr)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

package web

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (s *Server) handleViewEmail(c *gin.Context) {
	token := c.Param("token")

	// Get and validate token
	viewToken, err := s.db.GetEmailViewToken(token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get view token")
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}

	if viewToken == nil {
		c.String(http.StatusNotFound, "Email not found or link has expired")
		return
	}

	// Get email
	email, err := s.db.GetEmailMessageByID(viewToken.EmailID)
	if err != nil || email == nil {
		log.Error().Err(err).Str("email_id", viewToken.EmailID).Msg("Failed to get email")
		c.String(http.StatusNotFound, "Email not found")
		return
	}

	// Increment view count
	s.db.IncrementTokenViewCount(token)

	// Render email
	htmlContent := ""
	if email.SanitizedHTML != nil {
		htmlContent = *email.SanitizedHTML
	} else if email.TextBody != nil {
		// Convert plain text to HTML
		htmlContent = "<pre>" + template.HTMLEscapeString(*email.TextBody) + "</pre>"
	} else {
		htmlContent = "<p>No content available</p>"
	}

	subject := "No subject"
	if email.Subject != nil {
		subject = *email.Subject
	}

	fromName := email.FromAddress
	if email.FromName != nil {
		fromName = *email.FromName + " <" + email.FromAddress + ">"
	}

	data := gin.H{
		"Subject":     subject,
		"From":        fromName,
		"To":          email.ToAddresses,
		"Date":        email.Date.Format("2006-01-02 15:04:05"),
		"HTMLContent": template.HTML(htmlContent),
	}

	c.HTML(http.StatusOK, "email.html", data)
}

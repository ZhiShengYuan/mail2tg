package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

type OAuthManager struct {
	config *oauth2.Config
}

func NewOAuthManager(credentialsPath string) (*OAuthManager, error) {
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailModifyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	return &OAuthManager{config: config}, nil
}

func (o *OAuthManager) GetAuthURL(state string) string {
	return o.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (o *OAuthManager) ExchangeCode(code string) (*oauth2.Token, error) {
	ctx := context.Background()
	token, err := o.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("unable to exchange code: %w", err)
	}
	return token, nil
}

func (o *OAuthManager) TokenFromJSON(jsonToken string) (*oauth2.Token, error) {
	var token oauth2.Token
	if err := json.Unmarshal([]byte(jsonToken), &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (o *OAuthManager) TokenToJSON(token *oauth2.Token) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (o *OAuthManager) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now(),
	}

	ctx := context.Background()
	tokenSource := o.config.TokenSource(ctx, token)

	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token: %w", err)
	}

	return newToken, nil
}

func (o *OAuthManager) GetClient(token *oauth2.Token) *oauth2.TokenSource {
	ctx := context.Background()
	tokenSource := o.config.TokenSource(ctx, token)
	return &tokenSource
}

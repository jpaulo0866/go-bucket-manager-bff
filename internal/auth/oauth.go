package auth

import (
	"context"
	"go-bucket-manager-bff/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthManager struct {
	config *oauth2.Config
}

func NewOAuthManager(cfg *config.Config) *OAuthManager {
	return &OAuthManager{
		config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  "http://localhost:" + cfg.ServerPort + "/login/oauth2/code/google",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (o *OAuthManager) GetAuthURL(state string) string {
	return o.config.AuthCodeURL(state)
}

func (o *OAuthManager) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return o.config.Exchange(ctx, code)
}

func (o *OAuthManager) Config() *oauth2.Config {
    return o.config
}
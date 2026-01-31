package discord

import (
	"golang.org/x/oauth2"

	"github.com/zitadel/saml/pkg/provider"
	"github.com/zitadel/zitadel/internal/idp"
	"github.com/zitadel/zitadel/internal/idp/providers/oauth"
)

const (
	authURL  string = "https://discord.com/oauth2/authorize"
	tokenURL string = "https://discord.com/api/oauth2/token"
	userURL  string = "https://discord.com/api/users/@me"
)

var _ idp.Provider = (*Provider)(nil)

// Provider is the [idp.Provider] implementation for Discord
type Provider struct {
	*oauth.Provider
}

type ProviderOpts func(provider *Provider)

// New creates a Discord provider using the [oauth.Provider] (OAuth 2.0 generic provider)
func New(clientID, clientSecret, redirectURI string, scopes []string, opts ...ProviderOpts) (*Provider, error) {
	for _, opt := range opts {
		opt(provider)
	}
	config := newConfig(clientID, clientSecret, redirectURI, scopes)
	rp, err := oauth.New(
		config,
		name,
		userURL,
		func() idp.User {
			return &User{isEmailVerified: provider.emailVerified}
		},
		provider.options...,
	)
	if err != nil {
		return nil, err
	}
	provider.Provider = rp
	return provider, nil
}

func newConfig(clientID, secret, callbackURL string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  callbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURLTemplate,
			TokenURL: tokenURLTemplate,
		},
		Scopes: ensureMinimalScope(scopes),
	}
}

// ensureMinimalScope ensures that at least identify is set
// if none is provided it will request `identify email`
func ensureMinimalScope(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{"identify", "email"}
	}
	var identifySet bool
	for _, scope := range scopes {
		if scope == "identify" {
			identifySet = true
			continue
		}
	}
	if !identifySet {
		scopes = append(scopes, "identify")
	}
	return scopes
}

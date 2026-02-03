package discord

import (
	"fmt"
	"strconv"

	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/idp"
	"github.com/zitadel/zitadel/internal/idp/providers/oauth"
)

const (
	name     string = "Discord"
	authURL  string = "https://discord.com/oauth2/authorize"
	tokenURL string = "https://discord.com/api/oauth2/token"
	userURL  string = "https://discord.com/api/users/@me"
)

var _ idp.Provider = (*Provider)(nil)

// Provider is the [idp.Provider] implementation for Discord
type Provider struct {
	*oauth.Provider
}

// New creates a Discord provider using the [oauth.Provider] (OAuth 2.0 generic provider)
func New(clientID, clientSecret, redirectURI string, scopes []string, prompt string, options ...oauth.ProviderOpts) (*Provider, error) {
	config := newConfig(clientID, clientSecret, redirectURI, scopes)
	// Prepend prompt option so user-provided options can override it
	options = append([]oauth.ProviderOpts{withPrompt(prompt)}, options...)
	rp, err := oauth.New(
		config,
		name,
		userURL,
		func() idp.User {
			return new(User)
		},
		options...,
	)
	if err != nil {
		return nil, err
	}
	return &Provider{
		Provider: rp,
	}, nil
}

// withPrompt returns an OAuth provider option that adds the prompt parameter to the auth URL.
// Discord supports "consent" and "none". If empty, defaults to "consent".
func withPrompt(prompt string) oauth.ProviderOpts {
	if prompt == "" {
		prompt = "consent"
	}
	return oauth.WithAuthURLOpt(func(_ bool) rp.AuthURLOpt {
		return promptOpt(prompt)
	})
}

// promptOpt creates an rp.AuthURLOpt that sets the prompt parameter.
func promptOpt(prompt string) rp.AuthURLOpt {
	return func() []oauth2.AuthCodeOption {
		return []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("prompt", prompt)}
	}
}

func newConfig(clientID, secret, callbackURL string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: secret,
		RedirectURL:  callbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: ensureMinimalScope(scopes),
	}
}

// User is a representation of the authenticated Discord user and implements the [idp.User] interface
// https://discord.com/developers/docs/resources/user#user-object
type User struct {
	ID                   string              `json:"id"`
	Username             string              `json:"username"`
	Discriminator        string              `json:"discriminator"`
	GlobalName           string              `json:"global_name,omitempty"`
	Avatar               string              `json:"avatar,omitempty"`
	Bot                  bool                `json:"bot,omitempty"`
	System               bool                `json:"system,omitempty"`
	MFAEnabled           bool                `json:"mfa_enabled,omitempty"`
	Banner               string              `json:"banner,omitempty"`
	AccentColor          int                 `json:"accent_color,omitempty"`
	Locale               string              `json:"locale,omitempty"`
	Verified             bool                `json:"verified,omitempty"`
	Email                domain.EmailAddress `json:"email,omitempty"`
	Flags                int                 `json:"flags,omitempty"`
	PremiumType          int                 `json:"premium_type,omitempty"`
	PublicFlags          int                 `json:"public_flags,omitempty"`
	AvatarDecorationData struct {
		Asset string `json:"asset"`
		SkuID string `json:"sku_id"`
	} `json:"avatar_decoration_data,omitempty"`
	Collectibles struct {
		Nameplate struct {
			SkuID   string `json:"sku_id"`
			Asset   string `json:"asset"`
			Label   string `json:"label"`
			Palette string `json:"palette"`
		} `json:"nameplate,omitempty"`
	} `json:"collectibles,omitempty"`
	PrimaryGuild struct {
		IdentityGuildID string `json:"identity_guild_id,omitempty"`
		IdentityEnabled bool   `json:"identity_enabled,omitempty"`
		Tag             string `json:"tag,omitempty"`
		Badge           string `json:"badge,omitempty"`
	} `json:"primary_guild,omitempty"`
}

// GetID is an implementation of the [idp.User] interface
func (u *User) GetID() string {
	return u.ID
}

// GetFirstName is an implementation of the [idp.User] interface
// It returns an empty string because Discord does not provide a first name.
func (u *User) GetFirstName() string {
	return ""
}

// GetLastName is an implementation of the [idp.User] interface
// It returns an empty string because Discord does not provide a last name.
func (u *User) GetLastName() string {
	return ""
}

// GetDisplayName is an implementation of the [idp.User] interface
// It returns the GlobalName if set, otherwise the Username.
func (u *User) GetDisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

// GetNickname is an implementation of the [idp.User] interface
// It returns the Username.
func (u *User) GetNickname() string {
	if u.Discriminator != "" && u.Discriminator != "0" {
		return u.Username + "#" + u.Discriminator
	}
	return u.Username
}

// GetPreferredUsername is an implementation of the [idp.User] interface
// It returns the Username.
func (u *User) GetPreferredUsername() string {
	return u.Username
}

// GetEmail is an implementation of the [idp.User] interface
func (u *User) GetEmail() domain.EmailAddress {
	return u.Email
}

// IsEmailVerified is an implementation of the [idp.User] interface
func (u *User) IsEmailVerified() bool {
	return u.Verified
}

// GetPhone is an implementation of the [idp.User] interface
// It returns an empty string because Discord does not provide the user's phone.
func (u *User) GetPhone() domain.PhoneNumber {
	return ""
}

// IsPhoneVerified is an implementation of the [idp.User] interface
// It returns false because Discord does not provide the user's phone.
func (u *User) IsPhoneVerified() bool {
	return false
}

// GetPreferredLanguage is an implementation of the [idp.User] interface
// It returns the locale as a BCP 47 language tag.
func (u *User) GetPreferredLanguage() language.Tag {
	return language.Make(u.Locale)
}

// GetProfile is an implementation of the [idp.User] interface
// It returns the URL to the user's Discord profile.
func (u *User) GetProfile() string {
	return "https://discord.com/users/" + u.ID
}

// GetAvatarURL is an implementation of the [idp.User] interface
// It returns the URL to the user's avatar.
func (u *User) GetAvatarURL() string {
	if u.Avatar == "" {
		// Parse the Discord snowflake ID (string) to compute the default avatar index.
		// If parsing fails or ID is empty, fall back to index 0.
		index := 0
		if u.ID != "" {
			if v, err := strconv.ParseUint(u.ID, 10, 64); err == nil {
				index = int((v >> 22) % 6)
			}
		}
		return fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", index)
	}
	return "https://cdn.discordapp.com/avatars/" + u.ID + "/" + u.Avatar + ".png"
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

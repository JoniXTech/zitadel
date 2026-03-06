package discord

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/idp/providers/oauth"
)

func TestSession_FetchUser(t *testing.T) {
	type fields struct {
		name         string
		clientID     string
		clientSecret string
		redirectURI  string
		scopes       []string
		prompt       string
		httpMock     func()
		options      []oauth.ProviderOpts
		authURL      string
		code         string
		tokens       *oidc.Tokens[*oidc.IDTokenClaims]
	}
	type want struct {
		err               func(error) bool
		id                string
		firstName         string
		lastName          string
		displayName       string
		nickName          string
		preferredUsername string
		email             string
		isEmailVerified   bool
		phone             string
		isPhoneVerified   bool
		preferredLanguage language.Tag
		avatarURL         string
		profile           string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "unauthenticated session, error",
			fields: fields{
				clientID:     "clientID",
				clientSecret: "clientSecret",
				redirectURI:  "redirectURI",
				httpMock: func() {
					gock.New("https://discord.com").
						Get("/api/users/@me").
						Reply(http.StatusOK).
						JSON(userinfo())
				},
				authURL: "https://discord.com/oauth2/authorize?client_id=clientID&redirect_uri=redirectURI&response_type=code&scope=identify+email&state=testState",
				tokens:  nil,
			},
			want: want{
				err: func(err error) bool {
					return errors.Is(err, oauth.ErrCodeMissing)
				},
			},
		},
		{
			name: "user error",
			fields: fields{
				clientID:     "clientID",
				clientSecret: "clientSecret",
				redirectURI:  "redirectURI",
				httpMock: func() {
					gock.New("https://discord.com").
						Get("/api/users/@me").
						Reply(http.StatusInternalServerError)
				},
				authURL: "https://discord.com/oauth2/authorize?client_id=clientID&redirect_uri=redirectURI&response_type=code&scope=identify+email&state=testState",
				tokens: &oidc.Tokens[*oidc.IDTokenClaims]{
					Token: &oauth2.Token{
						AccessToken: "accessToken",
						TokenType:   oidc.BearerToken,
					},
					IDTokenClaims: oidc.NewIDTokenClaims(
						"https://discord.com/api/oauth2/token",
						"sub2",
						[]string{"clientID"},
						time.Now().Add(1*time.Hour),
						time.Now().Add(-1*time.Second),
						"nonce",
						"",
						nil,
						"clientID",
						0,
					),
				},
			},
			want: want{
				err: func(err error) bool {
					return err.Error() == "http status not ok: 500 Internal Server Error "
				},
			},
		},
		{
			name: "successful fetch",
			fields: fields{
				clientID:     "clientID",
				clientSecret: "clientSecret",
				redirectURI:  "redirectURI",
				httpMock: func() {
					gock.New("https://discord.com").
						Get("/api/users/@me").
						Reply(http.StatusOK).
						JSON(userinfo())
				},
				authURL: "https://discord.com/oauth2/authorize?client_id=clientID&redirect_uri=redirectURI&response_type=code&scope=identify+email&state=testState",
				tokens: &oidc.Tokens[*oidc.IDTokenClaims]{
					Token: &oauth2.Token{
						AccessToken: "accessToken",
						TokenType:   oidc.BearerToken,
					},
					IDTokenClaims: oidc.NewIDTokenClaims(
						"https://discord.com/api/oauth2/token",
						"sub",
						[]string{"clientID"},
						time.Now().Add(1*time.Hour),
						time.Now().Add(-1*time.Second),
						"nonce",
						"",
						nil,
						"clientID",
						0,
					),
				},
			},
			want: want{
				id:                "id",
				firstName:         "username",
				lastName:          "0",
				displayName:       "firstname lastname",
				nickName:          "username",
				preferredUsername: "username",
				email:             "email",
				isEmailVerified:   true,
				phone:             "",
				isPhoneVerified:   false,
				preferredLanguage: language.English,
				avatarURL:         "https://cdn.discordapp.com/avatars/id/avatarhash.png",
				profile:           "https://discord.com/users/id",
			},
		},
		{
			name: "successful fetch with email verified",
			fields: fields{
				clientID:     "clientID",
				clientSecret: "clientSecret",
				redirectURI:  "redirectURI",
				options:      []oauth.ProviderOpts{},
				httpMock: func() {
					gock.New("https://discord.com").
						Get("/api/users/@me").
						Reply(http.StatusOK).
						JSON(userinfo())
				},
				authURL: "https://discord.com/oauth2/authorize?client_id=clientID&redirect_uri=redirectURI&response_type=code&scope=identify+email&state=testState",
				tokens: &oidc.Tokens[*oidc.IDTokenClaims]{
					Token: &oauth2.Token{
						AccessToken: "accessToken",
						TokenType:   oidc.BearerToken,
					},
					IDTokenClaims: oidc.NewIDTokenClaims(
						"https://discord.com/api/oauth2/token",
						"sub",
						[]string{"clientID"},
						time.Now().Add(1*time.Hour),
						time.Now().Add(-1*time.Second),
						"nonce",
						"",
						nil,
						"clientID",
						0,
					),
				},
			},
			want: want{
				id:                "id",
				firstName:         "username",
				lastName:          "0",
				displayName:       "firstname lastname",
				nickName:          "username",
				preferredUsername: "username",
				email:             "email",
				isEmailVerified:   true,
				phone:             "",
				isPhoneVerified:   false,
				preferredLanguage: language.English,
				avatarURL:         "https://cdn.discordapp.com/avatars/id/avatarhash.png",
				profile:           "https://discord.com/users/id",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gock.Off()
			if tt.fields.httpMock != nil {
				tt.fields.httpMock()
			}
			a := assert.New(t)

			provider, err := New(tt.fields.clientID, tt.fields.clientSecret, tt.fields.redirectURI, tt.fields.scopes, tt.fields.prompt, tt.fields.options...)
			require.NoError(t, err)

			session := &oauth.Session{
				AuthURL:  tt.fields.authURL,
				Code:     tt.fields.code,
				Tokens:   tt.fields.tokens,
				Provider: provider.Provider,
			}

			user, err := session.FetchUser(context.Background())
			if tt.want.err != nil {
				if !tt.want.err(err) {
					a.Fail("invalid error", err)
				}
				return
			}
			a.NoError(err)
			a.Equal(tt.want.id, user.GetID())
			a.Equal(tt.want.firstName, user.GetFirstName())
			a.Equal(tt.want.lastName, user.GetLastName())
			a.Equal(tt.want.displayName, user.GetDisplayName())
			a.Equal(tt.want.nickName, user.GetNickname())
			a.Equal(tt.want.preferredUsername, user.GetPreferredUsername())
			a.Equal(domain.EmailAddress(tt.want.email), user.GetEmail())
			a.Equal(tt.want.isEmailVerified, user.IsEmailVerified())
			a.Equal(domain.PhoneNumber(tt.want.phone), user.GetPhone())
			a.Equal(tt.want.isPhoneVerified, user.IsPhoneVerified())
			a.Equal(tt.want.preferredLanguage, user.GetPreferredLanguage())
			a.Equal(tt.want.avatarURL, user.GetAvatarURL())
			a.Equal(tt.want.profile, user.GetProfile())

			// Verify RawInfo is populated with Discord-specific fields
			if du, ok := user.(*User); ok {
				a.NotNil(du.RawInfo)
				a.Equal(tt.want.id, du.RawInfo["id"])
			}
		})
	}
}

func userinfo() *User {
	return &User{
		ID:            "id",
		Username:      "username",
		Discriminator: "0",
		GlobalName:    "firstname lastname",
		Avatar:        "avatarhash",
		Bot:           false,
		System:        false,
		MFAEnabled:    false,
		Banner:        "bannerhash",
		AccentColor:   16711680,
		Locale:        "en",
		Verified:      true,
		Email:         "email",
		Flags:         0,
		PremiumType:   0,
		PublicFlags:   0,
	}
}

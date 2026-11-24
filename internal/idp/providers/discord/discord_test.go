package discord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zitadel/zitadel/internal/idp"
	"github.com/zitadel/zitadel/internal/idp/providers/oidc"
)

func TestProvider_BeginAuth(t *testing.T) {
	type fields struct {
		clientID     string
		clientSecret string
		redirectURI  string
		scopes       []string
		prompt       string
	}
	tests := []struct {
		name   string
		fields fields
		want   idp.Session
	}{
		{
			name: "successful auth",
			fields: fields{
				clientID:     "clientID",
				clientSecret: "clientSecret",
				redirectURI:  "redirectURI",
				scopes:       []string{"identify"},
				prompt:       "consent",
			},
			want: &oidc.Session{
				AuthURL: "https://discord.com/oauth2/authorize?client_id=clientID&prompt=consent&redirect_uri=redirectURI&response_type=code&scope=identify&state=testState",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			provider, err := New(tt.fields.clientID, tt.fields.clientSecret, tt.fields.redirectURI, tt.fields.scopes, tt.fields.prompt)
			r.NoError(err)

			ctx := context.Background()
			session, err := provider.BeginAuth(ctx, "testState")
			r.NoError(err)
			auth, err := session.GetAuth(ctx)
			authExpected, errExpected := tt.want.GetAuth(ctx)
			a.ErrorIs(err, errExpected)
			a.Equal(authExpected, auth)
		})
	}
}

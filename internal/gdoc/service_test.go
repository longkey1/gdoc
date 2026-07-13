package gdoc

import (
	"testing"

	"github.com/longkey1/gdoc/internal/google"
)

func TestNewAuthenticator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		authType    AuthType
		wantOAuth   bool
		wantSvcAcct bool
	}{
		{name: "oauth", authType: AuthTypeOAuth, wantOAuth: true},
		{name: "service account", authType: AuthTypeServiceAccount, wantSvcAcct: true},
		{name: "empty defaults to oauth", authType: "", wantOAuth: true},
		{name: "unknown defaults to oauth", authType: "something-else", wantOAuth: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newAuthenticator(&Config{
				AuthType:                     tt.authType,
				GoogleApplicationCredentials: "/creds.json",
				GoogleUserCredentials:        "/token.json",
			})

			if _, ok := got.(*google.OAuthAuthenticator); ok != tt.wantOAuth {
				t.Errorf("newAuthenticator() OAuthAuthenticator = %v, want %v", ok, tt.wantOAuth)
			}
			if _, ok := got.(*google.ServiceAccountAuthenticator); ok != tt.wantSvcAcct {
				t.Errorf("newAuthenticator() ServiceAccountAuthenticator = %v, want %v", ok, tt.wantSvcAcct)
			}
		})
	}
}

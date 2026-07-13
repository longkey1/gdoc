package gdoc

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// writeFile creates a file with the given content, creating parent
// directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// LoadConfig reads from the global viper instance, so these subtests
// reset it and must not run in parallel.
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Config
	}{
		{
			name: "all fields",
			content: `
auth_type = "service_account"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
`,
			want: Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
			},
		},
		{
			name: "oauth",
			content: `
auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
`,
			want: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
				GoogleUserCredentials:        "/path/to/token.json",
			},
		},
		{
			name:    "missing auth_type defaults to oauth",
			content: "application_credentials = \"/path/to/credentials.json\"\n",
			want: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/path/to/credentials.json",
			},
		},
		{
			name:    "empty file defaults to oauth",
			content: "",
			want:    Config{AuthType: AuthTypeOAuth},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			t.Cleanup(viper.Reset)

			path := filepath.Join(t.TempDir(), "config.toml")
			writeFile(t, path, tt.content)
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("ReadInConfig() error = %v", err)
			}

			got, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("LoadConfig() = %+v, want %+v", *got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "oauth with all credentials",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/creds.json",
				GoogleUserCredentials:        "/token.json",
			},
		},
		{
			name: "service account without user credentials",
			config: Config{
				AuthType:                     AuthTypeServiceAccount,
				GoogleApplicationCredentials: "/creds.json",
			},
		},
		{
			name: "missing application credentials",
			config: Config{
				AuthType:              AuthTypeOAuth,
				GoogleUserCredentials: "/token.json",
			},
			wantErr: true,
		},
		{
			name: "oauth without user credentials",
			config: Config{
				AuthType:                     AuthTypeOAuth,
				GoogleApplicationCredentials: "/creds.json",
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

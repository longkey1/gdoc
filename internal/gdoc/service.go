package gdoc

import (
	"context"

	"github.com/longkey1/gdoc/internal/google"
)

// Service represents the gdoc application service
type Service struct {
	Docs  *google.DocsService
	Drive *google.DriveService
}

// NewService creates a new gdoc service based on the configuration
func NewService(ctx context.Context, config *Config) (*Service, error) {
	auth := newAuthenticator(config)

	docsSvc, err := google.NewDocsService(ctx, auth)
	if err != nil {
		return nil, err
	}

	driveSvc, err := google.NewDriveService(ctx, auth)
	if err != nil {
		return nil, err
	}

	return &Service{
		Docs:  docsSvc,
		Drive: driveSvc,
	}, nil
}

func newAuthenticator(config *Config) google.Authenticator {
	switch config.AuthType {
	case AuthTypeServiceAccount:
		return google.NewServiceAccountAuthenticator(config.GoogleApplicationCredentials)
	case AuthTypeOAuth:
		fallthrough
	default:
		return google.NewOAuthAuthenticator(
			config.GoogleApplicationCredentials,
			config.GoogleUserCredentials,
		)
	}
}

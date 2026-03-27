package google

import (
	"context"
	"fmt"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// DocsService wraps the Google Docs API service
type DocsService struct {
	*docs.Service
}

// NewDocsService creates a new Docs service with the given authenticator
func NewDocsService(ctx context.Context, auth Authenticator) (*DocsService, error) {
	client, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %v", err)
	}

	var srv *docs.Service
	if client != nil {
		srv, err = docs.NewService(ctx, option.WithHTTPClient(client))
	} else {
		srv, err = docs.NewService(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create docs service: %v", err)
	}

	return &DocsService{srv}, nil
}

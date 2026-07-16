# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, etc.) when working with code in this repository.

## Project Overview

gdoc is a CLI client for Google Docs. It uses the Google Docs API and Drive API to list documents and retrieve their content in plain text, JSON, or Markdown format. Built with Go 1.26, cobra, and viper.

## Build & Development Commands

```bash
make build       # Build binary to bin/gdoc
make test        # Run tests (go test ./...)
make fmt         # Format code (go fmt ./...)
make vet         # Vet code (go vet ./...)
make lint        # Run golangci-lint (version managed by go.mod tool directive)
make tidy        # Tidy dependencies (go mod tidy)
```

Release: `make release type=patch|minor|major dryrun=false` (GitHub Actions + goreleaser)

## Architecture

```
main.go                     # Entry point. Version info injected via ldflags
cmd/                        # cobra command definitions
  root.go                   # Config loading ($HOME/.config/gdoc/config.toml via viper)
  auth.go                   # OAuth flow (only for auth_type=oauth)
  list.go                   # List documents via Drive API
  get.go                    # Get document content via Docs API (--format text|json|markdown, --tab)
internal/
  gdoc/
    config.go               # Config struct (auth_type, application_credentials, user_credentials)
    service.go              # Service: creates Authenticator based on auth_type, initializes Docs/Drive services
    docs.go                 # ListDocuments, GetDocumentRaw, FindTabBody, extractText
    markdown.go             # Google Docs API structs → Markdown conversion
    format.go               # Output formatters (text/json/markdown, tablewriter)
  google/
    auth.go                 # Authenticator interface, OAuthAuthenticator, ServiceAccountAuthenticator
    docs.go                 # DocsService wrapper
    drive.go                # DriveService wrapper
  version/                  # Version info set at build time via ldflags
```

### Key Design Patterns

- **Auth abstraction**: `google.Authenticator` interface switches between OAuth and Service Account. `gdoc.Service` selects the concrete type based on `auth_type` in config.
- **Tab support**: Supports Google Docs tab feature. Fetches with `IncludeTabsContent(true)` and recursively searches tabs via `FindTabBody`.
- **Output format**: `OutputFormat` type uniformly handles text/json/markdown output across commands.

## Configuration

Config file: `$HOME/.config/gdoc/config.toml`

```toml
auth_type = "oauth"                    # "oauth" or "service_account"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
```

## Environment

- direnv (`.envrc`) sets GOROOT/GOPATH/GOBIN/GOCACHE
- Go version managed via `got` command (`got path 1.26`)
- Product name stored in `.product_name` file

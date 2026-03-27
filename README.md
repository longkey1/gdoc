# gdoc

Google Docs CLI client - A command-line tool for interacting with Google Documents.

## Installation

### Using Go

```bash
go install github.com/longkey1/gdoc@latest
```

### Using Homebrew (coming soon)

```bash
brew install longkey1/tap/gdoc
```

## Setup

### 1. Create Google Cloud Project and OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Docs API and Google Drive API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Download the credentials JSON file

### 2. Create Configuration File

Create `~/.config/gdoc/config.toml`:

```toml
auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
```

### 3. Authenticate

```bash
gdoc auth
```

This will open your browser for Google OAuth authentication.

## Usage

### List Documents

```bash
# List your documents
gdoc list

# Search documents by name
gdoc list -q "meeting notes"

# Show only documents owned by me
gdoc list --mine

# Output as JSON
gdoc list --format json
```

### Get Document Content

```bash
# Get document as plain text
gdoc get <document-id>

# Get document as Markdown (headings, bold, italic, lists, tables, links)
gdoc get <document-id> --format markdown

# Get raw Google Docs API response as JSON
gdoc get <document-id> --format json

# Get content from a specific tab
gdoc get <document-id> --tab <tab-id>

# Get a specific tab as Markdown
gdoc get <document-id> --tab <tab-id> --format markdown
```

### Tabs

Google Docs supports multiple tabs within a single document. Use the `--tab` flag with the `get` command to access specific tabs.

- Without `--tab`: returns the first tab's content
- With `--tab <tab-id>`: returns the specified tab's content
- With `--format json`: returns the full API response including all tabs

To find tab IDs, use `--format json` and look at the `tabs[].tabProperties.tabId` field.

### Output Formats

| Format | Description |
|--------|-------------|
| `text` | Plain text output (default) |
| `markdown` | Markdown conversion with headings, text styles, lists, tables, and links |
| `json` | Raw Google Docs API response (includes all tabs) |

#### Markdown Conversion Details

The following elements are converted:

- **Headings**: `TITLE` / `HEADING_1` → `#`, `HEADING_2` → `##`, ... `HEADING_6` → `######`
- **Bold**: `**text**`
- **Italic**: `*text*`
- **Strikethrough**: `~~text~~`
- **Links**: `[text](url)`
- **Unordered lists**: `- item`
- **Ordered lists**: `1. item`
- **Nested lists**: indented with spaces
- **Tables**: Markdown table format
- **Superscript/Subscript**: `<sup>` / `<sub>` tags

### Typical Workflow

```bash
# 1. Find the document
gdoc list -q "project proposal"
# → ID: abc123, NAME: "Project Proposal 2026"

# 2. Get the content as Markdown
gdoc get abc123 --format markdown

# 3. Export to a file
gdoc get abc123 --format markdown > proposal.md

# 4. Check tabs (if the document has multiple tabs)
gdoc get abc123 --format json | jq '.tabs[].tabProperties'

# 5. Get a specific tab
gdoc get abc123 --tab t.123456 --format markdown
```

### Version

```bash
gdoc version
```

## Configuration Options

| Option | Description |
|--------|-------------|
| `auth_type` | Authentication type: `oauth` or `service_account` |
| `application_credentials` | Path to OAuth client credentials JSON file |
| `user_credentials` | Path to store OAuth user token (for OAuth auth type) |

## License

Apache License 2.0

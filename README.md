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

gdoc supports two authentication methods: OAuth (for personal use) and Service Account (for automation/CI).

### OAuth

#### 1. Create Google Cloud Project and OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Docs API and Google Drive API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Download the credentials JSON file

#### 2. Create Configuration File

Create `~/.config/gdoc/config.toml`:

```toml
auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
```

#### 3. Authenticate

```bash
gdoc auth
```

This will open your browser for Google OAuth authentication.

### Service Account

#### 1. Create a Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Docs API and Google Drive API
4. Create a Service Account and download its credentials JSON file
5. Share the target documents/folders with the service account's email address

#### 2. Create Configuration File

Create `~/.config/gdoc/config.toml`:

```toml
auth_type = "service_account"
application_credentials = "/path/to/service-account.json"
```

`user_credentials` and the `gdoc auth` command are not needed for Service Account authentication.

## Usage

### List Documents

```bash
# List your documents
gdoc list

# Search documents by name
gdoc list -q "meeting notes"

# Show only documents owned by me
gdoc list --mine

# Limit the number of results (default: 20)
gdoc list --max-results 50

# Output as JSON
gdoc list --format json
```

### Get Document Content

The `get` command accepts either a document ID or a Google Docs URL.

```bash
# Get document as plain text
gdoc get <document-id>

# Pass a Google Docs URL instead of an ID
gdoc get https://docs.google.com/document/d/<document-id>/edit

# Get document as Markdown (headings, bold, italic, lists, tables, links)
gdoc get <document-id> --format markdown

# Get raw Google Docs API response as JSON
gdoc get <document-id> --format json

# Get content from a specific tab
gdoc get <document-id> --tab <tab-id>

# Get a specific tab as Markdown
gdoc get <document-id> --tab <tab-id> --format markdown

# Pass a URL that already includes a tab; --tab is inferred from the URL
gdoc get "https://docs.google.com/document/d/<document-id>/edit?tab=<tab-id>"
```

When the URL contains a `tab` query parameter, the tab is taken from the URL. Specifying `--tab` together with such a URL is rejected as ambiguous.

### Create a Document

```bash
# Create an empty document
gdoc create --title "My New Document"

# Create with content from stdin
echo "Hello, World!" | gdoc create --title "My Document"

# Create with content from a file
gdoc create --title "My Document" -f content.txt

# Create with Markdown formatting
gdoc create --title "My Document" -f content.md --format markdown
```

### Update Document Content

The `update` command accepts either a document ID or a Google Docs URL, with the
same URL handling rules as `get` (a `tab` query parameter is honored; combining
it with `--tab` is rejected as ambiguous).

```bash
# Replace entire content (stdin)
echo "New content" | gdoc update <document-id>

# Pass a Google Docs URL instead of an ID
echo "New content" | gdoc update https://docs.google.com/document/d/<document-id>/edit

# Replace entire content from a file
gdoc update <document-id> -f content.txt

# Append to the end
echo "Appended text" | gdoc update <document-id> --append

# Prepend to the beginning
echo "Prepended text" | gdoc update <document-id> --append beginning

# Update a specific tab
echo "Tab content" | gdoc update <document-id> --tab <tab-id>

# Update a specific tab via URL
echo "Tab content" | gdoc update "https://docs.google.com/document/d/<document-id>/edit?tab=<tab-id>"

# Append to a specific tab
echo "More content" | gdoc update <document-id> --tab <tab-id> --append

# Update with Markdown formatting
echo "# Heading\n**bold** text" | gdoc update <document-id> --format markdown
```

### Tabs

Google Docs supports multiple tabs within a single document. Use the `--tab` flag with `get` and `update` commands to access specific tabs.

- Without `--tab`: targets the first tab
- With `--tab <tab-id>`: targets the specified tab
- With `--format json` (get only): returns the full API response including all tabs

To find tab IDs, use `--format json` and look at the `tabs[].tabProperties.tabId` field.

> **Note**: Tab creation is not supported by the Google Docs API. Create tabs in the Google Docs UI, then use `gdoc update --tab <tab-id>` to write content.

### Input Formats

| Format | Description |
|--------|-------------|
| `text` | Plain text input (default) |
| `markdown` | Markdown converted to Google Docs formatting (headings, bold, italic, strikethrough, links, lists) |

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

# 6. Create a new document from a Markdown file
gdoc create --title "Weekly Report" -f report.md --format markdown

# 7. Append to an existing document
echo "Updated at $(date)" | gdoc update abc123 --append
```

> **Note**: If you previously authenticated with read-only scopes, run `gdoc auth` again after upgrading to a version with write support.

### Version

```bash
gdoc version

# Show only the version number
gdoc version --short
```

## Configuration Options

| Option | Description |
|--------|-------------|
| `auth_type` | Authentication type: `oauth` or `service_account` |
| `application_credentials` | Path to OAuth client credentials JSON file |
| `user_credentials` | Path to store OAuth user token (for OAuth auth type) |
| `read_only` | When `true`, disables write commands (`create`, `update`). Also settable via the `GDOC_READ_ONLY` env var, and overridable per-invocation with `--read-only`/`--read-only=false` |

## License

Apache License 2.0

package gdoc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/api/docs/v1"
)

func TestFormatDocumentList(t *testing.T) {
	t.Parallel()

	documents := []DocumentInfo{
		{
			ID:           "doc1",
			Name:         "First Doc",
			CreatedTime:  "2026-01-01T00:00:00Z",
			ModifiedTime: "2026-01-02T00:00:00Z",
			WebViewLink:  "https://docs.google.com/document/d/doc1/edit",
		},
		{
			ID:           "doc2",
			Name:         "Second Doc",
			CreatedTime:  "2026-02-01T00:00:00Z",
			ModifiedTime: "2026-02-02T00:00:00Z",
			WebViewLink:  "https://docs.google.com/document/d/doc2/edit",
		},
	}

	t.Run("json round-trips", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatDocumentList(&buf, documents, OutputFormatJSON); err != nil {
			t.Fatalf("FormatDocumentList() error = %v", err)
		}

		var got []DocumentInfo
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if len(got) != len(documents) {
			t.Fatalf("decoded %d documents, want %d", len(got), len(documents))
		}
		for i := range documents {
			if got[i] != documents[i] {
				t.Errorf("document[%d] = %+v, want %+v", i, got[i], documents[i])
			}
		}
	})

	t.Run("text renders a table", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatDocumentList(&buf, documents, OutputFormatText); err != nil {
			t.Fatalf("FormatDocumentList() error = %v", err)
		}

		out := buf.String()
		for _, want := range []string{"ID", "NAME", "MODIFIED", "doc1", "First Doc", "doc2", "Second Doc", "2026-01-02T00:00:00Z"} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "https://docs.google.com") {
			t.Errorf("table output must not include web view links:\n%s", out)
		}
	})
}

func TestFormatDocumentContent(t *testing.T) {
	t.Parallel()

	content := &DocumentContent{
		ID:    "doc1",
		Title: "My Doc",
		Body:  "line one\nline two\n",
	}

	t.Run("text writes the body verbatim", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatDocumentContent(&buf, content, OutputFormatText); err != nil {
			t.Fatalf("FormatDocumentContent() error = %v", err)
		}
		if got := buf.String(); got != content.Body {
			t.Errorf("FormatDocumentContent() = %q, want %q", got, content.Body)
		}
	})

	t.Run("json round-trips", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatDocumentContent(&buf, content, OutputFormatJSON); err != nil {
			t.Fatalf("FormatDocumentContent() error = %v", err)
		}

		var got DocumentContent
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if got != *content {
			t.Errorf("decoded content = %+v, want %+v", got, *content)
		}
	})
}

func TestFormatDocumentRaw(t *testing.T) {
	t.Parallel()

	doc := &docs.Document{DocumentId: "doc1", Title: "My Doc"}

	var buf bytes.Buffer
	if err := FormatDocumentRaw(&buf, doc); err != nil {
		t.Fatalf("FormatDocumentRaw() error = %v", err)
	}

	var got docs.Document
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if got.DocumentId != "doc1" || got.Title != "My Doc" {
		t.Errorf("decoded document = %+v, want id %q and title %q", got, "doc1", "My Doc")
	}
}

func TestFormatDocumentTab(t *testing.T) {
	t.Parallel()

	tab := tabWithBody(
		para("HEADING_1", textElem("Head\n", nil)),
		para("", textElem("body text\n", nil)),
	)

	tests := []struct {
		name   string
		format OutputFormat
		want   string
	}{
		{
			name:   "text extracts plain text",
			format: OutputFormatText,
			want:   "Head\nbody text\n",
		},
		{
			name:   "markdown converts headings",
			format: OutputFormatMarkdown,
			want:   "# Head\nbody text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			if err := FormatDocumentTab(&buf, tab, tt.format); err != nil {
				t.Fatalf("FormatDocumentTab() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("FormatDocumentTab() = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("json round-trips", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		if err := FormatDocumentTab(&buf, tab, OutputFormatJSON); err != nil {
			t.Fatalf("FormatDocumentTab() error = %v", err)
		}

		var got docs.DocumentTab
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("output is not valid JSON: %v", err)
		}
		if got.Body == nil || len(got.Body.Content) != 2 {
			t.Errorf("decoded tab body = %+v, want 2 content elements", got.Body)
		}
	})
}

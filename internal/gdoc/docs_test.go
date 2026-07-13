package gdoc

import (
	"testing"

	"google.golang.org/api/docs/v1"
)

func TestParseDocumentInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantDocID string
		wantTabID string
		wantErr   bool
	}{
		{
			name:      "bare document id",
			input:     "1AbC-dEf_123",
			wantDocID: "1AbC-dEf_123",
		},
		{
			name:      "https url",
			input:     "https://docs.google.com/document/d/1AbC/edit",
			wantDocID: "1AbC",
		},
		{
			name:      "http url",
			input:     "http://docs.google.com/document/d/1AbC/edit",
			wantDocID: "1AbC",
		},
		{
			name:      "url with tab query",
			input:     "https://docs.google.com/document/d/1AbC/edit?tab=t.abc123",
			wantDocID: "1AbC",
			wantTabID: "t.abc123",
		},
		{
			name:      "url without trailing path segment",
			input:     "https://docs.google.com/document/d/1AbC",
			wantDocID: "1AbC",
		},
		{
			name:    "non-docs google url",
			input:   "https://docs.google.com/spreadsheets/d/1AbC/edit",
			wantErr: true,
		},
		{
			name:    "url with short path",
			input:   "https://docs.google.com/document/",
			wantErr: true,
		},
		{
			name:    "unparseable url",
			input:   "https://[::1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			docID, tabID, err := ParseDocumentInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDocumentInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if docID != tt.wantDocID {
				t.Errorf("ParseDocumentInput() docID = %q, want %q", docID, tt.wantDocID)
			}
			if tabID != tt.wantTabID {
				t.Errorf("ParseDocumentInput() tabID = %q, want %q", tabID, tt.wantTabID)
			}
		})
	}
}

func TestFindTabBody(t *testing.T) {
	t.Parallel()

	first := tabWithBody(para("", textElem("first\n", nil)))
	second := tabWithBody(para("", textElem("second\n", nil)))
	child := tabWithBody(para("", textElem("child\n", nil)))

	doc := &docs.Document{Tabs: []*docs.Tab{
		{TabProperties: &docs.TabProperties{TabId: "t1"}, DocumentTab: first},
		{
			TabProperties: &docs.TabProperties{TabId: "t2"},
			DocumentTab:   second,
			ChildTabs: []*docs.Tab{
				{TabProperties: &docs.TabProperties{TabId: "t3"}, DocumentTab: child},
			},
		},
	}}

	tests := []struct {
		name    string
		doc     *docs.Document
		tabID   string
		want    *docs.DocumentTab
		wantErr bool
	}{
		{
			name:  "empty tab id returns first tab",
			doc:   doc,
			tabID: "",
			want:  first,
		},
		{
			name:  "top-level tab by id",
			doc:   doc,
			tabID: "t2",
			want:  second,
		},
		{
			name:  "nested child tab by id",
			doc:   doc,
			tabID: "t3",
			want:  child,
		},
		{
			name:    "tab not found",
			doc:     doc,
			tabID:   "nosuch",
			wantErr: true,
		},
		{
			name:    "document without tabs",
			doc:     &docs.Document{},
			tabID:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FindTabBody(tt.doc, tt.tabID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FindTabBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("FindTabBody() = %p, want %p", got, tt.want)
			}
		})
	}
}

func TestExtractText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body *docs.Body
		want string
	}{
		{
			name: "nil body",
			body: nil,
			want: "",
		},
		{
			name: "empty body",
			body: &docs.Body{},
			want: "",
		},
		{
			name: "paragraph runs are concatenated",
			body: &docs.Body{Content: []*docs.StructuralElement{
				para("", textElem("Hello, ", nil), textElem("world\n", nil)),
				para("", textElem("next\n", nil)),
			}},
			want: "Hello, world\nnext\n",
		},
		{
			name: "styles are ignored in plain text",
			body: &docs.Body{Content: []*docs.StructuralElement{
				para("HEADING_1", textElem("Head\n", &docs.TextStyle{Bold: true})),
			}},
			want: "Head\n",
		},
		{
			name: "table cells are tab separated",
			body: &docs.Body{Content: []*docs.StructuralElement{
				{Table: &docs.Table{TableRows: []*docs.TableRow{
					{TableCells: []*docs.TableCell{cell("a"), cell("b")}},
					{TableCells: []*docs.TableCell{cell("c"), cell("d")}},
				}}},
			}},
			want: "a\tb\nc\td\n",
		},
		{
			name: "elements without text runs are skipped",
			body: &docs.Body{Content: []*docs.StructuralElement{
				para("", &docs.ParagraphElement{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "obj1"}}),
			}},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractText(tt.body); got != tt.want {
				t.Errorf("extractText() = %q, want %q", got, tt.want)
			}
		})
	}
}

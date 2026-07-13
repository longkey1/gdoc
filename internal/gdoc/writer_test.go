package gdoc

import (
	"testing"

	"google.golang.org/api/docs/v1"
)

// docWithTab builds a document containing a single tab with the given
// ID and body content.
func docWithTab(tabID string, content ...*docs.StructuralElement) *docs.Document {
	return &docs.Document{Tabs: []*docs.Tab{
		{
			TabProperties: &docs.TabProperties{TabId: tabID},
			DocumentTab:   &docs.DocumentTab{Body: &docs.Body{Content: content}},
		},
	}}
}

func TestGetTabEndIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		doc     *docs.Document
		tabID   string
		want    int64
		wantErr bool
	}{
		{
			name: "last element end index",
			doc: docWithTab("t1",
				&docs.StructuralElement{StartIndex: 0, EndIndex: 10},
				&docs.StructuralElement{StartIndex: 10, EndIndex: 42},
			),
			tabID: "t1",
			want:  42,
		},
		{
			name:  "empty tab id uses first tab",
			doc:   docWithTab("t1", &docs.StructuralElement{EndIndex: 7}),
			tabID: "",
			want:  7,
		},
		{
			name:  "empty body content",
			doc:   docWithTab("t1"),
			tabID: "t1",
			want:  1,
		},
		{
			name: "nil body",
			doc: &docs.Document{Tabs: []*docs.Tab{
				{
					TabProperties: &docs.TabProperties{TabId: "t1"},
					DocumentTab:   &docs.DocumentTab{},
				},
			}},
			tabID: "t1",
			want:  1,
		},
		{
			name:    "tab not found",
			doc:     docWithTab("t1", &docs.StructuralElement{EndIndex: 7}),
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

			got, err := GetTabEndIndex(tt.doc, tt.tabID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetTabEndIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("GetTabEndIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

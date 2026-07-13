package gdoc

import (
	"testing"

	"google.golang.org/api/docs/v1"
)

// textElem builds a paragraph element containing a text run.
func textElem(content string, style *docs.TextStyle) *docs.ParagraphElement {
	return &docs.ParagraphElement{TextRun: &docs.TextRun{Content: content, TextStyle: style}}
}

// para builds a structural element containing a paragraph with an
// optional named style.
func para(namedStyle string, elems ...*docs.ParagraphElement) *docs.StructuralElement {
	p := &docs.Paragraph{Elements: elems}
	if namedStyle != "" {
		p.ParagraphStyle = &docs.ParagraphStyle{NamedStyleType: namedStyle}
	}
	return &docs.StructuralElement{Paragraph: p}
}

// bulletPara builds a structural element for a list item.
func bulletPara(listID string, nesting int64, elems ...*docs.ParagraphElement) *docs.StructuralElement {
	return &docs.StructuralElement{Paragraph: &docs.Paragraph{
		Bullet:   &docs.Bullet{ListId: listID, NestingLevel: nesting},
		Elements: elems,
	}}
}

// cell builds a table cell containing a single text paragraph.
func cell(text string) *docs.TableCell {
	return &docs.TableCell{Content: []*docs.StructuralElement{para("", textElem(text+"\n", nil))}}
}

// tabWithBody builds a document tab from structural elements.
func tabWithBody(elems ...*docs.StructuralElement) *docs.DocumentTab {
	return &docs.DocumentTab{Body: &docs.Body{Content: elems}}
}

func TestDocumentTabToMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tab  *docs.DocumentTab
		want string
	}{
		{
			name: "nil tab",
			tab:  nil,
			want: "",
		},
		{
			name: "nil body",
			tab:  &docs.DocumentTab{},
			want: "",
		},
		{
			name: "plain paragraph",
			tab:  tabWithBody(para("", textElem("Hello\n", nil))),
			want: "Hello\n",
		},
		{
			name: "title",
			tab:  tabWithBody(para("TITLE", textElem("Doc Title\n", nil))),
			want: "# Doc Title\n",
		},
		{
			name: "heading levels",
			tab: tabWithBody(
				para("HEADING_1", textElem("One\n", nil)),
				para("HEADING_2", textElem("Two\n", nil)),
				para("HEADING_3", textElem("Three\n", nil)),
				para("HEADING_4", textElem("Four\n", nil)),
				para("HEADING_5", textElem("Five\n", nil)),
				para("HEADING_6", textElem("Six\n", nil)),
			),
			want: "# One\n## Two\n### Three\n#### Four\n##### Five\n###### Six\n",
		},
		{
			name: "bold",
			tab:  tabWithBody(para("", textElem("Hello\n", &docs.TextStyle{Bold: true}))),
			want: "**Hello**\n",
		},
		{
			name: "italic",
			tab:  tabWithBody(para("", textElem("Hello\n", &docs.TextStyle{Italic: true}))),
			want: "*Hello*\n",
		},
		{
			name: "strikethrough",
			tab:  tabWithBody(para("", textElem("Hello\n", &docs.TextStyle{Strikethrough: true}))),
			want: "~~Hello~~\n",
		},
		{
			name: "bold italic",
			tab:  tabWithBody(para("", textElem("Hi\n", &docs.TextStyle{Bold: true, Italic: true}))),
			want: "***Hi***\n",
		},
		{
			name: "link",
			tab:  tabWithBody(para("", textElem("text\n", &docs.TextStyle{Link: &docs.Link{Url: "https://example.com"}}))),
			want: "[text](https://example.com)\n",
		},
		{
			name: "superscript",
			tab:  tabWithBody(para("", textElem("2\n", &docs.TextStyle{BaselineOffset: "SUPERSCRIPT"}))),
			want: "<sup>2</sup>\n",
		},
		{
			name: "subscript",
			tab:  tabWithBody(para("", textElem("2\n", &docs.TextStyle{BaselineOffset: "SUBSCRIPT"}))),
			want: "<sub>2</sub>\n",
		},
		{
			name: "whitespace-only run is not styled",
			tab:  tabWithBody(para("", textElem("\n", &docs.TextStyle{Bold: true}))),
			want: "\n",
		},
		{
			name: "mixed runs in one paragraph",
			tab: tabWithBody(para("",
				textElem("plain ", nil),
				textElem("bold", &docs.TextStyle{Bold: true}),
				textElem("\n", nil),
			)),
			want: "plain **bold**\n",
		},
		{
			name: "nested unordered list",
			tab: tabWithBody(
				bulletPara("list1", 0, textElem("a\n", nil)),
				bulletPara("list1", 1, textElem("b\n", nil)),
			),
			want: "- a\n  - b\n",
		},
		{
			name: "ordered list",
			tab: &docs.DocumentTab{
				Body: &docs.Body{Content: []*docs.StructuralElement{
					bulletPara("list1", 0, textElem("a\n", nil)),
				}},
				Lists: map[string]docs.List{
					"list1": {ListProperties: &docs.ListProperties{
						NestingLevels: []*docs.NestingLevel{{GlyphType: "DECIMAL"}},
					}},
				},
			},
			want: "1. a\n",
		},
		{
			name: "bullet with unknown list id",
			tab:  tabWithBody(bulletPara("nosuch", 0, textElem("a\n", nil))),
			want: "- a\n",
		},
		{
			name: "table",
			tab: tabWithBody(&docs.StructuralElement{Table: &docs.Table{TableRows: []*docs.TableRow{
				{TableCells: []*docs.TableCell{cell("H1"), cell("H2")}},
				{TableCells: []*docs.TableCell{cell("a"), cell("b")}},
			}}}),
			want: "| H1 | H2 |\n| --- | --- |\n| a | b |\n\n",
		},
		{
			name: "empty table",
			tab:  tabWithBody(&docs.StructuralElement{Table: &docs.Table{}}),
			want: "",
		},
		{
			name: "section break",
			tab:  tabWithBody(&docs.StructuralElement{SectionBreak: &docs.SectionBreak{}}),
			want: "\n---\n\n",
		},
		{
			name: "horizontal rule",
			tab: tabWithBody(para("",
				&docs.ParagraphElement{HorizontalRule: &docs.HorizontalRule{}},
				textElem("\n", nil),
			)),
			want: "\n---\n\n",
		},
		{
			name: "inline image",
			tab: tabWithBody(para("",
				&docs.ParagraphElement{InlineObjectElement: &docs.InlineObjectElement{InlineObjectId: "obj1"}},
				textElem("\n", nil),
			)),
			want: "![image](obj1)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := DocumentTabToMarkdown(tt.tab); got != tt.want {
				t.Errorf("DocumentTabToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDocumentToMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		doc  *docs.Document
		want string
	}{
		{
			name: "nil document",
			doc:  nil,
			want: "",
		},
		{
			name: "first tab is used",
			doc: &docs.Document{
				Tabs: []*docs.Tab{{DocumentTab: tabWithBody(para("", textElem("from tab\n", nil)))}},
				Body: &docs.Body{Content: []*docs.StructuralElement{para("", textElem("from body\n", nil))}},
			},
			want: "from tab\n",
		},
		{
			name: "body fallback without tabs",
			doc: &docs.Document{
				Body: &docs.Body{Content: []*docs.StructuralElement{para("HEADING_1", textElem("Head\n", nil))}},
			},
			want: "# Head\n",
		},
		{
			name: "body fallback when first tab has no document tab",
			doc: &docs.Document{
				Tabs: []*docs.Tab{{}},
				Body: &docs.Body{Content: []*docs.StructuralElement{para("", textElem("from body\n", nil))}},
			},
			want: "from body\n",
		},
		{
			name: "no tabs and no body",
			doc:  &docs.Document{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := DocumentToMarkdown(tt.doc); got != tt.want {
				t.Errorf("DocumentToMarkdown() = %q, want %q", got, tt.want)
			}
		})
	}
}

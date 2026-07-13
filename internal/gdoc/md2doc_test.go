package gdoc

import (
	"reflect"
	"testing"

	"google.golang.org/api/docs/v1"
)

func TestParseLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		want parsedLine
	}{
		{
			name: "empty line",
			line: "",
			want: parsedLine{segments: []textSegment{{text: ""}}},
		},
		{
			name: "plain text",
			line: "hello world",
			want: parsedLine{segments: []textSegment{{text: "hello world"}}},
		},
		{
			name: "heading level 1",
			line: "# Title",
			want: parsedLine{headingLevel: 1, segments: []textSegment{{text: "Title"}}},
		},
		{
			name: "heading level 3",
			line: "### Section",
			want: parsedLine{headingLevel: 3, segments: []textSegment{{text: "Section"}}},
		},
		{
			name: "heading level 6",
			line: "###### Deep",
			want: parsedLine{headingLevel: 6, segments: []textSegment{{text: "Deep"}}},
		},
		{
			name: "seven hashes is not a heading",
			line: "####### x",
			want: parsedLine{segments: []textSegment{{text: "####### x"}}},
		},
		{
			name: "hash without space is not a heading",
			line: "#nospace",
			want: parsedLine{segments: []textSegment{{text: "#nospace"}}},
		},
		{
			name: "unordered list with dash",
			line: "- item",
			want: parsedLine{isBullet: true, segments: []textSegment{{text: "item"}}},
		},
		{
			name: "unordered list with asterisk",
			line: "* item",
			want: parsedLine{isBullet: true, segments: []textSegment{{text: "item"}}},
		},
		{
			name: "ordered list",
			line: "1. item",
			want: parsedLine{isOrdered: true, segments: []textSegment{{text: "item"}}},
		},
		{
			name: "ordered list with multi-digit number",
			line: "10. item",
			want: parsedLine{isOrdered: true, segments: []textSegment{{text: "item"}}},
		},
		{
			name: "heading with inline formatting",
			line: "# **b**",
			want: parsedLine{headingLevel: 1, segments: []textSegment{{text: "b", bold: true}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := parseLine(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLine(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseInlineFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want []textSegment
	}{
		{
			name: "empty string",
			text: "",
			want: []textSegment{{text: ""}},
		},
		{
			name: "plain text",
			text: "plain",
			want: []textSegment{{text: "plain"}},
		},
		{
			name: "bold only",
			text: "**bold**",
			want: []textSegment{{text: "bold", bold: true}},
		},
		{
			name: "bold surrounded by plain text",
			text: "pre **b** post",
			want: []textSegment{
				{text: "pre "},
				{text: "b", bold: true},
				{text: " post"},
			},
		},
		{
			name: "italic",
			text: "*i*",
			want: []textSegment{{text: "i", italic: true}},
		},
		{
			name: "strikethrough",
			text: "~~s~~",
			want: []textSegment{{text: "s", strikethrough: true}},
		},
		{
			name: "link",
			text: "[t](https://example.com)",
			want: []textSegment{{text: "t", linkURL: "https://example.com"}},
		},
		{
			name: "italic and bold in one line",
			text: "a *b* **c**",
			want: []textSegment{
				{text: "a "},
				{text: "b", italic: true},
				{text: " "},
				{text: "c", bold: true},
			},
		},
		{
			name: "bold and link in one line",
			text: "**a** [b](c)",
			want: []textSegment{
				{text: "a", bold: true},
				{text: " "},
				{text: "b", linkURL: "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := parseInlineFormatting(tt.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseInlineFormatting(%q) = %+v, want %+v", tt.text, got, tt.want)
			}
		})
	}
}

func TestToRequests(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		baseIndex int64
		tabID     string
		markdown  string
		want      []*docs.Request
	}{
		{
			name:      "empty markdown",
			baseIndex: 1,
			markdown:  "",
			want:      nil,
		},
		{
			name:      "plain text",
			baseIndex: 1,
			markdown:  "hello",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "hello",
					Location: &docs.Location{Index: 1},
				}},
			},
		},
		{
			name:      "tab id is propagated",
			baseIndex: 1,
			tabID:     "t.1",
			markdown:  "x",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "x",
					Location: &docs.Location{Index: 1, TabId: "t.1"},
				}},
			},
		},
		{
			name:      "heading",
			baseIndex: 1,
			markdown:  "# Title",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "Title",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range:          &docs.Range{StartIndex: 1, EndIndex: 6},
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "HEADING_1"},
					Fields:         "namedStyleType",
				}},
			},
		},
		{
			name:      "multibyte heading offsets count runes",
			baseIndex: 1,
			markdown:  "# 見出し",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "見出し",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range:          &docs.Range{StartIndex: 1, EndIndex: 4},
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "HEADING_1"},
					Fields:         "namedStyleType",
				}},
			},
		},
		{
			name:      "bold",
			baseIndex: 1,
			markdown:  "a **b**",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "a b",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{StartIndex: 3, EndIndex: 4},
					TextStyle: &docs.TextStyle{
						Bold:            true,
						ForceSendFields: []string{"Bold"},
					},
					Fields: "bold",
				}},
			},
		},
		{
			name:      "italic",
			baseIndex: 1,
			markdown:  "*i*",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "i",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{StartIndex: 1, EndIndex: 2},
					TextStyle: &docs.TextStyle{
						Italic:          true,
						ForceSendFields: []string{"Italic"},
					},
					Fields: "italic",
				}},
			},
		},
		{
			name:      "strikethrough",
			baseIndex: 1,
			markdown:  "~~s~~",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "s",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{StartIndex: 1, EndIndex: 2},
					TextStyle: &docs.TextStyle{
						Strikethrough:   true,
						ForceSendFields: []string{"Strikethrough"},
					},
					Fields: "strikethrough",
				}},
			},
		},
		{
			name:      "link",
			baseIndex: 1,
			markdown:  "[t](https://u)",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "t",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{StartIndex: 1, EndIndex: 2},
					TextStyle: &docs.TextStyle{
						Link: &docs.Link{Url: "https://u"},
					},
					Fields: "link",
				}},
			},
		},
		{
			name:      "unordered list",
			baseIndex: 1,
			markdown:  "- x",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "x",
					Location: &docs.Location{Index: 1},
				}},
				{CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range:        &docs.Range{StartIndex: 1, EndIndex: 2},
					BulletPreset: "BULLET_DISC_CIRCLE_SQUARE",
				}},
			},
		},
		{
			name:      "ordered list",
			baseIndex: 1,
			markdown:  "1. x",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "x",
					Location: &docs.Location{Index: 1},
				}},
				{CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range:        &docs.Range{StartIndex: 1, EndIndex: 2},
					BulletPreset: "NUMBERED_DECIMAL_ALPHA_ROMAN",
				}},
			},
		},
		{
			name:      "non-default base index",
			baseIndex: 5,
			markdown:  "- x",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "x",
					Location: &docs.Location{Index: 5},
				}},
				{CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range:        &docs.Range{StartIndex: 5, EndIndex: 6},
					BulletPreset: "BULLET_DISC_CIRCLE_SQUARE",
				}},
			},
		},
		{
			name:      "multiline document",
			baseIndex: 1,
			markdown:  "# Title\nHello **world**\n- item",
			want: []*docs.Request{
				{InsertText: &docs.InsertTextRequest{
					Text:     "Title\nHello world\nitem",
					Location: &docs.Location{Index: 1},
				}},
				{UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range:          &docs.Range{StartIndex: 1, EndIndex: 7},
					ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "HEADING_1"},
					Fields:         "namedStyleType",
				}},
				{UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{StartIndex: 13, EndIndex: 18},
					TextStyle: &docs.TextStyle{
						Bold:            true,
						ForceSendFields: []string{"Bold"},
					},
					Fields: "bold",
				}},
				{CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range:        &docs.Range{StartIndex: 19, EndIndex: 23},
					BulletPreset: "BULLET_DISC_CIRCLE_SQUARE",
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := &BasicMarkdownConverter{BaseIndex: tt.baseIndex, TabID: tt.tabID}
			got, err := c.ToRequests(tt.markdown)
			if err != nil {
				t.Fatalf("ToRequests() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToRequests() = %s, want %s", requestsString(got), requestsString(tt.want))
			}
		})
	}
}

// requestsString renders requests for readable test failure output.
func requestsString(reqs []*docs.Request) string {
	var sb []byte
	for _, r := range reqs {
		b, _ := r.MarshalJSON()
		sb = append(sb, b...)
		sb = append(sb, '\n')
	}
	return string(sb)
}

func TestHeadingLevelToNamedStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level int
		want  string
	}{
		{name: "level 1", level: 1, want: "HEADING_1"},
		{name: "level 2", level: 2, want: "HEADING_2"},
		{name: "level 3", level: 3, want: "HEADING_3"},
		{name: "level 4", level: 4, want: "HEADING_4"},
		{name: "level 5", level: 5, want: "HEADING_5"},
		{name: "level 6", level: 6, want: "HEADING_6"},
		{name: "level 0 falls back to normal text", level: 0, want: "NORMAL_TEXT"},
		{name: "level 7 falls back to normal text", level: 7, want: "NORMAL_TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := headingLevelToNamedStyle(tt.level); got != tt.want {
				t.Errorf("headingLevelToNamedStyle(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

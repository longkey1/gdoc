package gdoc

import (
	"regexp"
	"strings"

	"google.golang.org/api/docs/v1"
)

// BasicMarkdownConverter converts Markdown to Google Docs API batch update requests.
// It implements the MarkdownConverter interface defined in markdown.go.
type BasicMarkdownConverter struct {
	BaseIndex int64
	TabID     string
}

// textSegment represents a piece of text with its formatting.
type textSegment struct {
	text          string
	bold          bool
	italic        bool
	strikethrough bool
	linkURL       string
}

// parsedLine represents a parsed Markdown line.
type parsedLine struct {
	headingLevel int // 0 = normal paragraph, 1-6 = heading level
	isBullet     bool
	isOrdered    bool
	segments     []textSegment
}

var (
	boldRe          = regexp.MustCompile(`\*\*(.+?)\*\*`)
	strikethroughRe = regexp.MustCompile(`~~(.+?)~~`)
	linkRe          = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// ToRequests converts Markdown content to Google Docs API batch update requests.
func (c *BasicMarkdownConverter) ToRequests(markdown string) ([]*docs.Request, error) {
	lines := strings.Split(markdown, "\n")
	parsed := make([]parsedLine, 0, len(lines))

	for _, line := range lines {
		parsed = append(parsed, parseLine(line))
	}

	// Build plain text and track positions
	var plainText strings.Builder
	type linePosition struct {
		startOffset int64
		endOffset   int64
		line        parsedLine
	}
	type segmentPosition struct {
		startOffset int64
		endOffset   int64
		segment     textSegment
	}

	var linePositions []linePosition
	var segmentPositions []segmentPosition
	currentOffset := int64(0)

	for i, pl := range parsed {
		lineStart := currentOffset
		for _, seg := range pl.segments {
			segStart := currentOffset
			plainText.WriteString(seg.text)
			currentOffset += int64(len([]rune(seg.text)))
			segmentPositions = append(segmentPositions, segmentPosition{
				startOffset: segStart,
				endOffset:   currentOffset,
				segment:     seg,
			})
		}
		// Add newline after each line except the last
		if i < len(parsed)-1 {
			plainText.WriteString("\n")
			currentOffset++
		}
		linePositions = append(linePositions, linePosition{
			startOffset: lineStart,
			endOffset:   currentOffset,
			line:        pl,
		})
	}

	var requests []*docs.Request

	// Insert all text at once
	text := plainText.String()
	if text == "" {
		return nil, nil
	}

	requests = append(requests, &docs.Request{
		InsertText: &docs.InsertTextRequest{
			Text: text,
			Location: &docs.Location{
				Index: c.BaseIndex,
				TabId: c.TabID,
			},
		},
	})

	// Apply paragraph styles (headings)
	for _, lp := range linePositions {
		if lp.line.headingLevel > 0 {
			namedStyle := headingLevelToNamedStyle(lp.line.headingLevel)
			requests = append(requests, &docs.Request{
				UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + lp.startOffset,
						EndIndex:   c.BaseIndex + lp.endOffset,
						TabId:      c.TabID,
					},
					ParagraphStyle: &docs.ParagraphStyle{
						NamedStyleType: namedStyle,
					},
					Fields: "namedStyleType",
				},
			})
		}
	}

	// Apply text styles (bold, italic, strikethrough, links)
	for _, sp := range segmentPositions {
		if sp.segment.bold {
			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + sp.startOffset,
						EndIndex:   c.BaseIndex + sp.endOffset,
						TabId:      c.TabID,
					},
					TextStyle: &docs.TextStyle{
						Bold:            true,
						ForceSendFields: []string{"Bold"},
					},
					Fields: "bold",
				},
			})
		}
		if sp.segment.italic {
			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + sp.startOffset,
						EndIndex:   c.BaseIndex + sp.endOffset,
						TabId:      c.TabID,
					},
					TextStyle: &docs.TextStyle{
						Italic:          true,
						ForceSendFields: []string{"Italic"},
					},
					Fields: "italic",
				},
			})
		}
		if sp.segment.strikethrough {
			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + sp.startOffset,
						EndIndex:   c.BaseIndex + sp.endOffset,
						TabId:      c.TabID,
					},
					TextStyle: &docs.TextStyle{
						Strikethrough:   true,
						ForceSendFields: []string{"Strikethrough"},
					},
					Fields: "strikethrough",
				},
			})
		}
		if sp.segment.linkURL != "" {
			requests = append(requests, &docs.Request{
				UpdateTextStyle: &docs.UpdateTextStyleRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + sp.startOffset,
						EndIndex:   c.BaseIndex + sp.endOffset,
						TabId:      c.TabID,
					},
					TextStyle: &docs.TextStyle{
						Link: &docs.Link{
							Url: sp.segment.linkURL,
						},
					},
					Fields: "link",
				},
			})
		}
	}

	// Apply bullet styles
	for _, lp := range linePositions {
		if lp.line.isBullet {
			requests = append(requests, &docs.Request{
				CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + lp.startOffset,
						EndIndex:   c.BaseIndex + lp.endOffset,
						TabId:      c.TabID,
					},
					BulletPreset: "BULLET_DISC_CIRCLE_SQUARE",
				},
			})
		}
		if lp.line.isOrdered {
			requests = append(requests, &docs.Request{
				CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
					Range: &docs.Range{
						StartIndex: c.BaseIndex + lp.startOffset,
						EndIndex:   c.BaseIndex + lp.endOffset,
						TabId:      c.TabID,
					},
					BulletPreset: "NUMBERED_DECIMAL_ALPHA_ROMAN",
				},
			})
		}
	}

	return requests, nil
}

// parseLine parses a single Markdown line into a parsedLine.
func parseLine(line string) parsedLine {
	pl := parsedLine{}

	// Check for heading
	if strings.HasPrefix(line, "#") {
		level := 0
		for _, ch := range line {
			if ch == '#' {
				level++
			} else {
				break
			}
		}
		if level >= 1 && level <= 6 && len(line) > level && line[level] == ' ' {
			pl.headingLevel = level
			line = line[level+1:]
		}
	}

	// Check for unordered list
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		pl.isBullet = true
		line = line[2:]
	}

	// Check for ordered list
	if matched := regexp.MustCompile(`^\d+\.\s`).FindString(line); matched != "" {
		pl.isOrdered = true
		line = line[len(matched):]
	}

	// Parse inline formatting
	pl.segments = parseInlineFormatting(line)

	return pl
}

// markedRange represents a range of formatted text in the original Markdown string.
type markedRange struct {
	start         int
	end           int
	originalStart int
	originalEnd   int
	bold          bool
	italic        bool
	strikethrough bool
	linkURL       string
	innerText     string
}

// parseInlineFormatting parses inline Markdown formatting into text segments.
func parseInlineFormatting(text string) []textSegment {
	if text == "" {
		return []textSegment{{text: ""}}
	}

	// Find all formatting ranges in the original text
	var ranges []markedRange

	// Find links first (they contain text we need to preserve)
	for _, match := range linkRe.FindAllStringSubmatchIndex(text, -1) {
		linkText := text[match[2]:match[3]]
		linkURL := text[match[4]:match[5]]
		ranges = append(ranges, markedRange{
			originalStart: match[0],
			originalEnd:   match[1],
			linkURL:       linkURL,
			innerText:     linkText,
		})
	}

	// Find bold
	for _, match := range boldRe.FindAllStringSubmatchIndex(text, -1) {
		if !overlapsAny(ranges, match[0], match[1]) {
			ranges = append(ranges, markedRange{
				originalStart: match[0],
				originalEnd:   match[1],
				bold:          true,
				innerText:     text[match[2]:match[3]],
			})
		}
	}

	// Find strikethrough
	for _, match := range strikethroughRe.FindAllStringSubmatchIndex(text, -1) {
		if !overlapsAny(ranges, match[0], match[1]) {
			ranges = append(ranges, markedRange{
				originalStart: match[0],
				originalEnd:   match[1],
				strikethrough: true,
				innerText:     text[match[2]:match[3]],
			})
		}
	}

	// Find italic (single asterisk, not double)
	for _, match := range findItalicRanges(text, ranges) {
		ranges = append(ranges, match)
	}

	// Sort ranges by original start position
	sortRanges(ranges)

	// Build segments by walking through the text
	var segments []textSegment
	pos := 0
	for _, r := range ranges {
		// Add plain text before this range
		if r.originalStart > pos {
			segments = append(segments, textSegment{text: text[pos:r.originalStart]})
		}
		// Add formatted segment
		seg := textSegment{
			text:          r.innerText,
			bold:          r.bold,
			italic:        r.italic,
			strikethrough: r.strikethrough,
			linkURL:       r.linkURL,
		}
		segments = append(segments, seg)
		pos = r.originalEnd
	}

	// Add remaining plain text
	if pos < len(text) {
		segments = append(segments, textSegment{text: text[pos:]})
	}

	if len(segments) == 0 {
		return []textSegment{{text: text}}
	}

	return segments
}

// findItalicRanges finds italic markers that don't overlap with existing ranges.
func findItalicRanges(text string, existing []markedRange) []markedRange {
	var result []markedRange
	// Simple approach: find *text* patterns that aren't part of **text**
	i := 0
	for i < len(text) {
		if text[i] == '*' && (i+1 >= len(text) || text[i+1] != '*') {
			// Check it's not preceded by * (which would make it **)
			if i > 0 && text[i-1] == '*' {
				i++
				continue
			}
			// Find closing *
			end := strings.Index(text[i+1:], "*")
			if end > 0 {
				closeIdx := i + 1 + end
				// Make sure closing * is not part of **
				if closeIdx+1 < len(text) && text[closeIdx+1] == '*' {
					i++
					continue
				}
				if !overlapsAny(existing, i, closeIdx+1) {
					result = append(result, markedRange{
						originalStart: i,
						originalEnd:   closeIdx + 1,
						italic:        true,
						innerText:     text[i+1 : closeIdx],
					})
					i = closeIdx + 1
					continue
				}
			}
		}
		i++
	}
	return result
}

func overlapsAny(ranges []markedRange, start, end int) bool {
	for _, r := range ranges {
		if start < r.originalEnd && end > r.originalStart {
			return true
		}
	}
	return false
}

func sortRanges(ranges []markedRange) {
	for i := 1; i < len(ranges); i++ {
		for j := i; j > 0 && ranges[j].originalStart < ranges[j-1].originalStart; j-- {
			ranges[j], ranges[j-1] = ranges[j-1], ranges[j]
		}
	}
}

func headingLevelToNamedStyle(level int) string {
	switch level {
	case 1:
		return "HEADING_1"
	case 2:
		return "HEADING_2"
	case 3:
		return "HEADING_3"
	case 4:
		return "HEADING_4"
	case 5:
		return "HEADING_5"
	case 6:
		return "HEADING_6"
	default:
		return "NORMAL_TEXT"
	}
}

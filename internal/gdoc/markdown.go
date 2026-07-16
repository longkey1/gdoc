package gdoc

import (
	"fmt"
	"strings"

	"google.golang.org/api/docs/v1"
)

// DocumentTabToMarkdown converts a Google Docs DocumentTab to Markdown
func DocumentTabToMarkdown(tab *docs.DocumentTab) string {
	if tab == nil || tab.Body == nil {
		return ""
	}
	return bodyToMarkdown(tab.Body, tab.DocumentStyle, tab.Lists)
}

// DocumentToMarkdown converts a Google Docs document to Markdown (first tab fallback)
func DocumentToMarkdown(doc *docs.Document) string {
	if doc == nil {
		return ""
	}
	// Use tabs if available
	if len(doc.Tabs) > 0 && doc.Tabs[0].DocumentTab != nil {
		return DocumentTabToMarkdown(doc.Tabs[0].DocumentTab)
	}
	// Fallback to document body
	if doc.Body == nil {
		return ""
	}
	return bodyToMarkdown(doc.Body, nil, doc.Lists)
}

func bodyToMarkdown(body *docs.Body, _ *docs.DocumentStyle, lists map[string]docs.List) string {
	var sb strings.Builder

	for _, elem := range body.Content {
		if elem.Paragraph != nil {
			sb.WriteString(paragraphToMarkdown(elem.Paragraph, lists))
		}
		if elem.Table != nil {
			sb.WriteString(tableToMarkdown(elem.Table))
		}
		if elem.SectionBreak != nil {
			// Section breaks are rendered as horizontal rules
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String()
}

func paragraphToMarkdown(p *docs.Paragraph, lists map[string]docs.List) string {
	var sb strings.Builder

	// Check if this is a list item
	if p.Bullet != nil {
		sb.WriteString(bulletPrefix(p.Bullet, lists))
	} else {
		// Add heading prefix based on named style
		sb.WriteString(headingPrefix(p.ParagraphStyle))
	}

	// Process paragraph elements
	for _, elem := range p.Elements {
		if elem.TextRun != nil {
			text := elem.TextRun.Content
			// Don't style whitespace-only or newline-only content
			if strings.TrimSpace(text) == "" {
				sb.WriteString(text)
				continue
			}

			// Split trailing newline to avoid it being wrapped in style markers
			trimmed := strings.TrimRight(text, "\n")
			trailing := text[len(trimmed):]

			styled := applyTextStyle(trimmed, elem.TextRun.TextStyle)
			sb.WriteString(styled)
			sb.WriteString(trailing)
		}
		if elem.HorizontalRule != nil {
			sb.WriteString("\n---\n")
		}
		if elem.InlineObjectElement != nil {
			_, _ = fmt.Fprintf(&sb, "![image](%s)", elem.InlineObjectElement.InlineObjectId)
		}
	}

	return sb.String()
}

func headingPrefix(style *docs.ParagraphStyle) string {
	if style == nil {
		return ""
	}

	switch style.NamedStyleType {
	case "TITLE":
		return "# "
	case "HEADING_1":
		return "# "
	case "HEADING_2":
		return "## "
	case "HEADING_3":
		return "### "
	case "HEADING_4":
		return "#### "
	case "HEADING_5":
		return "##### "
	case "HEADING_6":
		return "###### "
	default:
		return ""
	}
}

func bulletPrefix(bullet *docs.Bullet, lists map[string]docs.List) string {
	nestingLevel := int(bullet.NestingLevel)
	indent := strings.Repeat("  ", nestingLevel)

	// Check if this is an ordered list
	if lists != nil && bullet.ListId != "" {
		if list, ok := lists[bullet.ListId]; ok {
			if list.ListProperties != nil && len(list.ListProperties.NestingLevels) > nestingLevel {
				nl := list.ListProperties.NestingLevels[nestingLevel]
				if nl.GlyphType == "DECIMAL" || nl.GlyphType == "ALPHA" || nl.GlyphType == "ROMAN" {
					return indent + "1. "
				}
			}
		}
	}

	return indent + "- "
}

func applyTextStyle(text string, style *docs.TextStyle) string {
	if style == nil || text == "" {
		return text
	}

	result := text

	if style.Bold {
		result = "**" + result + "**"
	}
	if style.Italic {
		result = "*" + result + "*"
	}
	if style.Strikethrough {
		result = "~~" + result + "~~"
	}
	if style.BaselineOffset == "SUPERSCRIPT" {
		result = "<sup>" + result + "</sup>"
	}
	if style.BaselineOffset == "SUBSCRIPT" {
		result = "<sub>" + result + "</sub>"
	}
	if style.Link != nil && style.Link.Url != "" {
		result = fmt.Sprintf("[%s](%s)", result, style.Link.Url)
	}

	return result
}

func tableToMarkdown(table *docs.Table) string {
	if table == nil || len(table.TableRows) == 0 {
		return ""
	}

	var sb strings.Builder

	for rowIdx, row := range table.TableRows {
		sb.WriteString("|")
		for _, cell := range row.TableCells {
			cellText := extractCellText(cell)
			sb.WriteString(" ")
			sb.WriteString(strings.ReplaceAll(cellText, "\n", " "))
			sb.WriteString(" |")
		}
		sb.WriteString("\n")

		// Add separator after header row
		if rowIdx == 0 {
			sb.WriteString("|")
			for range row.TableCells {
				sb.WriteString(" --- |")
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	return sb.String()
}

func extractCellText(cell *docs.TableCell) string {
	var sb strings.Builder
	for _, content := range cell.Content {
		if content.Paragraph != nil {
			for _, elem := range content.Paragraph.Elements {
				if elem.TextRun != nil {
					sb.WriteString(strings.TrimRight(elem.TextRun.Content, "\n"))
				}
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// MarkdownToDocRequests converts Markdown to Google Docs API batch update requests.
// This will be implemented when the update command is added.
// For now, this defines the interface for future use.
type MarkdownConverter interface {
	// ToRequests converts markdown content to Docs API batch update requests
	ToRequests(markdown string) ([]*docs.Request, error)
}

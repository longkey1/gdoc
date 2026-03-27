package gdoc

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	"google.golang.org/api/docs/v1"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatText     OutputFormat = "text"
	OutputFormatJSON     OutputFormat = "json"
	OutputFormatMarkdown OutputFormat = "markdown"
)

// FormatDocumentList outputs document list in the specified format
func FormatDocumentList(w io.Writer, documents []DocumentInfo, format OutputFormat) error {
	if format == OutputFormatJSON {
		data, err := json.MarshalIndent(documents, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	table := tablewriter.NewWriter(w)
	table.Header("ID", "NAME", "MODIFIED")

	for _, d := range documents {
		table.Append(d.ID, d.Name, d.ModifiedTime)
	}

	table.Render()
	return nil
}

// FormatDocumentContent outputs document content in the specified format (text only)
func FormatDocumentContent(w io.Writer, content *DocumentContent, format OutputFormat) error {
	if format == OutputFormatJSON {
		data, err := json.MarshalIndent(content, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	}

	fmt.Fprint(w, content.Body)
	return nil
}

// FormatDocumentRaw outputs the full raw API document as JSON
func FormatDocumentRaw(w io.Writer, doc *docs.Document) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal JSON: %w", err)
	}
	fmt.Fprintln(w, string(data))
	return nil
}

// FormatDocumentTab outputs a document tab in the specified format
func FormatDocumentTab(w io.Writer, tab *docs.DocumentTab, format OutputFormat) error {
	switch format {
	case OutputFormatJSON:
		data, err := json.MarshalIndent(tab, "", "  ")
		if err != nil {
			return fmt.Errorf("unable to marshal JSON: %w", err)
		}
		fmt.Fprintln(w, string(data))
		return nil
	case OutputFormatMarkdown:
		fmt.Fprint(w, DocumentTabToMarkdown(tab))
		return nil
	default:
		// text format: extract plain text
		fmt.Fprint(w, extractText(tab.Body))
		return nil
	}
}

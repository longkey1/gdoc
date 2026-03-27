package gdoc

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
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

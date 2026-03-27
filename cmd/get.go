/*
Copyright © 2025 longkey1

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"os"

	"github.com/longkey1/gdoc/internal/gdoc"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <document-id>",
	Short: "Get document content",
	Args:  cobra.ExactArgs(1),
	RunE:  runGet,
}

func runGet(cmd *cobra.Command, args []string) error {
	documentID := args[0]
	format, _ := cmd.Flags().GetString("format")
	tabID, _ := cmd.Flags().GetString("tab")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gdoc.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	doc, err := gdoc.GetDocumentRaw(svc.Docs.Service, documentID)
	if err != nil {
		return err
	}

	// JSON format outputs the full raw document
	if gdoc.OutputFormat(format) == gdoc.OutputFormatJSON {
		return gdoc.FormatDocumentRaw(os.Stdout, doc)
	}

	// For text/markdown, find the target tab
	tab, err := gdoc.FindTabBody(doc, tabID)
	if err != nil {
		return err
	}

	return gdoc.FormatDocumentTab(os.Stdout, tab, gdoc.OutputFormat(format))
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().String("format", "text", "Output format (text, json, or markdown)")
	getCmd.Flags().String("tab", "", "Tab ID to get content from (default: first tab)")
}

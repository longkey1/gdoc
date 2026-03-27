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
	"fmt"
	"os"

	"github.com/longkey1/gdoc/internal/gdoc"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Google Doc",
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	title, _ := cmd.Flags().GetString("title")
	filePath, _ := cmd.Flags().GetString("file")
	format, _ := cmd.Flags().GetString("format")

	if title == "" {
		return fmt.Errorf("--title is required")
	}

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gdoc.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	doc, err := gdoc.CreateDocument(svc.Docs.Service, title)
	if err != nil {
		return err
	}

	// Read input content if available
	content, err := gdoc.ReadInput(filePath)
	if err != nil {
		return err
	}

	if content != "" {
		if gdoc.OutputFormat(format) == gdoc.OutputFormatMarkdown {
			converter := &gdoc.BasicMarkdownConverter{
				BaseIndex: 1,
			}
			requests, err := converter.ToRequests(content)
			if err != nil {
				return fmt.Errorf("unable to convert markdown: %v", err)
			}
			if err := gdoc.BatchUpdate(svc.Docs.Service, doc.DocumentId, requests); err != nil {
				return err
			}
		} else {
			if err := gdoc.InsertText(svc.Docs.Service, doc.DocumentId, content, ""); err != nil {
				return err
			}
		}
	}

	fmt.Fprintf(os.Stdout, "Document created: %s\n", doc.DocumentId)
	fmt.Fprintf(os.Stdout, "URL: https://docs.google.com/document/d/%s/edit\n", doc.DocumentId)
	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("title", "", "Document title (required)")
	createCmd.Flags().StringP("file", "f", "", "Input file path (default: stdin)")
	createCmd.Flags().String("format", "text", "Input format (text or markdown)")
}

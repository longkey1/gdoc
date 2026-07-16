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
	"google.golang.org/api/docs/v1"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:         "update <document-id-or-url>",
	Short:       "Update document content",
	Args:        cobra.ExactArgs(1),
	RunE:        runUpdate,
	Annotations: map[string]string{writeAnnotation: "true"},
}

func runUpdate(cmd *cobra.Command, args []string) error {
	documentID, urlTabID, err := gdoc.ParseDocumentInput(args[0])
	if err != nil {
		return err
	}
	tabID, _ := cmd.Flags().GetString("tab")

	if urlTabID != "" {
		if cmd.Flags().Changed("tab") {
			return fmt.Errorf("cannot specify --tab when URL already contains tab")
		}
		tabID = urlTabID
	}
	appendMode, _ := cmd.Flags().GetString("append")
	filePath, _ := cmd.Flags().GetString("file")
	format, _ := cmd.Flags().GetString("format")

	appendFlagChanged := cmd.Flags().Changed("append")

	cfg := GetConfig()
	if err := cfg.Validate(); err != nil {
		return err
	}

	svc, err := gdoc.NewService(context.Background(), cfg)
	if err != nil {
		return err
	}

	content, err := gdoc.ReadInput(filePath)
	if err != nil {
		return err
	}

	if content == "" {
		return fmt.Errorf("no input content provided")
	}

	docsSvc := svc.Docs.Service

	if gdoc.OutputFormat(format) == gdoc.OutputFormatMarkdown {
		return updateWithMarkdown(docsSvc, documentID, tabID, appendMode, appendFlagChanged, content)
	}

	return updateWithText(docsSvc, documentID, tabID, appendMode, appendFlagChanged, content)
}

func updateWithText(svc *docs.Service, docID, tabID, appendMode string, appendFlagChanged bool, content string) error {
	if appendFlagChanged {
		if appendMode == "beginning" {
			if err := gdoc.PrependText(svc, docID, content, tabID); err != nil {
				return err
			}
		} else {
			if err := gdoc.AppendText(svc, docID, content, tabID); err != nil {
				return err
			}
		}
	} else {
		if err := gdoc.ReplaceContent(svc, docID, content, tabID); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "Document updated: %s\n", docID)
	_, _ = fmt.Fprintf(os.Stdout, "URL: https://docs.google.com/document/d/%s/edit\n", docID)
	return nil
}

func updateWithMarkdown(svc *docs.Service, docID, tabID, appendMode string, appendFlagChanged bool, content string) error {
	if appendFlagChanged {
		if appendMode == "beginning" {
			converter := &gdoc.BasicMarkdownConverter{
				BaseIndex: 1,
				TabID:     tabID,
			}
			requests, err := converter.ToRequests(content)
			if err != nil {
				return fmt.Errorf("unable to convert markdown: %v", err)
			}
			if err := gdoc.BatchUpdate(svc, docID, requests); err != nil {
				return err
			}
		} else {
			doc, err := gdoc.GetDocumentRaw(svc, docID)
			if err != nil {
				return err
			}
			endIndex, err := gdoc.GetTabEndIndex(doc, tabID)
			if err != nil {
				return err
			}
			converter := &gdoc.BasicMarkdownConverter{
				BaseIndex: endIndex - 1,
				TabID:     tabID,
			}
			requests, err := converter.ToRequests(content)
			if err != nil {
				return fmt.Errorf("unable to convert markdown: %v", err)
			}
			if err := gdoc.BatchUpdate(svc, docID, requests); err != nil {
				return err
			}
		}
	} else {
		doc, err := gdoc.GetDocumentRaw(svc, docID)
		if err != nil {
			return err
		}
		endIndex, err := gdoc.GetTabEndIndex(doc, tabID)
		if err != nil {
			return err
		}

		var requests []*docs.Request
		if endIndex > 2 {
			requests = append(requests, &docs.Request{
				DeleteContentRange: &docs.DeleteContentRangeRequest{
					Range: &docs.Range{
						StartIndex: 1,
						EndIndex:   endIndex - 1,
						TabId:      tabID,
					},
				},
			})
		}

		converter := &gdoc.BasicMarkdownConverter{
			BaseIndex: 1,
			TabID:     tabID,
		}
		mdRequests, err := converter.ToRequests(content)
		if err != nil {
			return fmt.Errorf("unable to convert markdown: %v", err)
		}
		requests = append(requests, mdRequests...)

		if err := gdoc.BatchUpdate(svc, docID, requests); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "Document updated: %s\n", docID)
	_, _ = fmt.Fprintf(os.Stdout, "URL: https://docs.google.com/document/d/%s/edit\n", docID)
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("tab", "", "Tab ID to update (default: first tab)")
	updateCmd.Flags().String("append", "end", "Append mode: 'beginning' or 'end' (default: replace entire content)")
	updateCmd.Flags().StringP("file", "f", "", "Input file path (default: stdin)")
	updateCmd.Flags().String("format", "text", "Input format (text or markdown)")
}

package gdoc

import (
	"fmt"
	"net/url"
	"strings"

	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// ParseDocumentInput accepts either a Google Docs document ID or a Google Docs URL,
// and returns the document ID along with an optional tab ID parsed from the URL's
// `tab` query parameter. Bare IDs return an empty tab ID.
func ParseDocumentInput(input string) (docID string, tabID string, err error) {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return input, "", nil
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", "", fmt.Errorf("invalid URL: %v", err)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "document" || parts[1] != "d" {
		return "", "", fmt.Errorf("not a Google Docs URL: %s", input)
	}

	return parts[2], u.Query().Get("tab"), nil
}

// DocumentInfo represents a document in Google Drive
type DocumentInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CreatedTime  string `json:"createdTime"`
	ModifiedTime string `json:"modifiedTime"`
	WebViewLink  string `json:"webViewLink"`
}

// ListDocuments returns the list of documents owned by or shared with the user
func ListDocuments(svc *drive.Service, query string, mine bool, maxResults int64) ([]DocumentInfo, error) {
	q := "mimeType='application/vnd.google-apps.document' and trashed=false"
	if mine {
		q += " and 'me' in owners"
	}
	if query != "" {
		q += fmt.Sprintf(" and name contains '%s'", query)
	}

	call := svc.Files.List().
		Q(q).
		Fields("files(id, name, createdTime, modifiedTime, webViewLink)").
		OrderBy("modifiedTime desc")

	if maxResults > 0 {
		call = call.PageSize(maxResults)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list documents: %v", err)
	}

	var result []DocumentInfo
	for _, f := range resp.Files {
		result = append(result, DocumentInfo{
			ID:           f.Id,
			Name:         f.Name,
			CreatedTime:  f.CreatedTime,
			ModifiedTime: f.ModifiedTime,
			WebViewLink:  f.WebViewLink,
		})
	}

	return result, nil
}

// DocumentContent represents the content of a Google Doc
type DocumentContent struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// GetDocumentRaw retrieves the raw Google Docs API document with tabs content
func GetDocumentRaw(svc *docs.Service, documentID string) (*docs.Document, error) {
	doc, err := svc.Documents.Get(documentID).IncludeTabsContent(true).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get document: %v", err)
	}
	return doc, nil
}

// FindTabBody finds a tab by ID and returns its DocumentTab.
// If tabID is empty, returns the first tab's DocumentTab.
// Searches recursively through child tabs.
func FindTabBody(doc *docs.Document, tabID string) (*docs.DocumentTab, error) {
	if len(doc.Tabs) == 0 {
		// Fallback for documents without tabs
		return nil, fmt.Errorf("no tabs found in document")
	}

	if tabID == "" {
		// Return first tab
		return doc.Tabs[0].DocumentTab, nil
	}

	// Search recursively
	tab := findTabByID(doc.Tabs, tabID)
	if tab == nil {
		return nil, fmt.Errorf("tab not found: %s", tabID)
	}
	return tab, nil
}

func findTabByID(tabs []*docs.Tab, tabID string) *docs.DocumentTab {
	for _, tab := range tabs {
		if tab.TabProperties != nil && tab.TabProperties.TabId == tabID {
			return tab.DocumentTab
		}
		if found := findTabByID(tab.ChildTabs, tabID); found != nil {
			return found
		}
	}
	return nil
}

// GetDocument retrieves a document by ID and extracts its text content
func GetDocument(svc *docs.Service, documentID string) (*DocumentContent, error) {
	doc, err := GetDocumentRaw(svc, documentID)
	if err != nil {
		return nil, err
	}

	body := extractText(doc.Body)

	return &DocumentContent{
		ID:    doc.DocumentId,
		Title: doc.Title,
		Body:  body,
	}, nil
}

// extractText extracts plain text from a Google Docs Body
func extractText(body *docs.Body) string {
	if body == nil {
		return ""
	}

	var sb strings.Builder
	for _, elem := range body.Content {
		if elem.Paragraph != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(pe.TextRun.Content)
				}
			}
		}
		if elem.Table != nil {
			for _, row := range elem.Table.TableRows {
				var cells []string
				for _, cell := range row.TableCells {
					var cellText strings.Builder
					for _, content := range cell.Content {
						if content.Paragraph != nil {
							for _, pe := range content.Paragraph.Elements {
								if pe.TextRun != nil {
									cellText.WriteString(pe.TextRun.Content)
								}
							}
						}
					}
					cells = append(cells, strings.TrimSpace(cellText.String()))
				}
				sb.WriteString(strings.Join(cells, "\t"))
				sb.WriteString("\n")
			}
		}
	}
	return sb.String()
}

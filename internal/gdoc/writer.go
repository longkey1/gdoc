package gdoc

import (
	"fmt"

	"google.golang.org/api/docs/v1"
)

// CreateDocument creates a new Google Doc with the given title.
func CreateDocument(svc *docs.Service, title string) (*docs.Document, error) {
	doc, err := svc.Documents.Create(&docs.Document{Title: title}).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create document: %v", err)
	}
	return doc, nil
}

// InsertText inserts text at the beginning of the document or tab body (index 1).
func InsertText(svc *docs.Service, docID string, text string, tabID string) error {
	req := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Text: text,
					Location: &docs.Location{
						Index: 1,
						TabId: tabID,
					},
				},
			},
		},
	}
	_, err := svc.Documents.BatchUpdate(docID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to insert text: %v", err)
	}
	return nil
}

// AppendText appends text at the end of the document or tab body.
func AppendText(svc *docs.Service, docID string, text string, tabID string) error {
	req := &docs.BatchUpdateDocumentRequest{
		Requests: []*docs.Request{
			{
				InsertText: &docs.InsertTextRequest{
					Text: text,
					EndOfSegmentLocation: &docs.EndOfSegmentLocation{
						TabId: tabID,
					},
				},
			},
		},
	}
	_, err := svc.Documents.BatchUpdate(docID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to append text: %v", err)
	}
	return nil
}

// PrependText inserts text at the beginning of the document or tab body.
func PrependText(svc *docs.Service, docID string, text string, tabID string) error {
	return InsertText(svc, docID, text, tabID)
}

// ReplaceContent replaces all content in the document or tab body with new text.
func ReplaceContent(svc *docs.Service, docID string, newContent string, tabID string) error {
	doc, err := GetDocumentRaw(svc, docID)
	if err != nil {
		return err
	}

	endIndex, err := GetTabEndIndex(doc, tabID)
	if err != nil {
		return err
	}

	var requests []*docs.Request

	// Delete existing content if there is any (endIndex > 2 means content beyond the trailing newline)
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

	// Insert new content
	if newContent != "" {
		requests = append(requests, &docs.Request{
			InsertText: &docs.InsertTextRequest{
				Text: newContent,
				Location: &docs.Location{
					Index: 1,
					TabId: tabID,
				},
			},
		})
	}

	if len(requests) == 0 {
		return nil
	}

	req := &docs.BatchUpdateDocumentRequest{Requests: requests}
	_, err = svc.Documents.BatchUpdate(docID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to replace content: %v", err)
	}
	return nil
}

// BatchUpdate sends a batch update request to the document.
func BatchUpdate(svc *docs.Service, docID string, requests []*docs.Request) error {
	if len(requests) == 0 {
		return nil
	}
	req := &docs.BatchUpdateDocumentRequest{Requests: requests}
	_, err := svc.Documents.BatchUpdate(docID, req).Do()
	if err != nil {
		return fmt.Errorf("unable to batch update document: %v", err)
	}
	return nil
}

// GetTabEndIndex returns the end index of the body content in the specified tab.
func GetTabEndIndex(doc *docs.Document, tabID string) (int64, error) {
	tab, err := FindTabBody(doc, tabID)
	if err != nil {
		return 0, err
	}

	if tab.Body == nil || len(tab.Body.Content) == 0 {
		return 1, nil
	}

	lastElem := tab.Body.Content[len(tab.Body.Content)-1]
	return lastElem.EndIndex, nil
}

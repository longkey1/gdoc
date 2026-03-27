package gdoc

import (
	"fmt"

	"google.golang.org/api/drive/v3"
)

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

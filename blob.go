package quip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type Blob struct {
	ID  string
	URL string
}

// XXX Return *Blob instead
func (q *Client) GetBlob(blobId, threadId string) ([]byte, error) {
	path := fmt.Sprintf("blob/%s/%s", threadId, blobId)
	resp, err := q.getJson(q.apiUrlResource(path), map[string]string{})
	return resp, err
}

func (q *Client) NewBlob(path string, threadId string) (*Blob, error) {

	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create form file param
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("blob", file.Name())
	if err != nil {
		return nil, err
	}

	// Copy file to form param
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	// Close writer
	if err := w.Close(); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("blob/%s", threadId)
	req, err := http.NewRequest("POST", q.apiUrlResource(endpoint), &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := q.doRequest(req, 1)
	if err != nil {
		return nil, err
	}
	var blob Blob
	if err := json.Unmarshal(resp, &blob); err != nil {
		return nil, err
	}
	return &blob, nil
}

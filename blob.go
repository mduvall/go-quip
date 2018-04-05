package quip

import (
	"fmt"
)

type Blob []byte

func (q *Client) GetBlob(blobId, threadId string) (Blob, error) {
	path := fmt.Sprintf("blob/%s/%s", threadId, blobId)
	resp, err := q.getJson(q.apiUrlResource(path), map[string]string{})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

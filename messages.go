package quip

import (
	"encoding/json"
)

type Message struct {
	AuthorId    string `json:"author_id"`
	CreatedUsec int    `json:"created_usec"`
	Id          string
	Text        string
}

type GetRecentMessagesParams struct {
	ThreadId       string
	Count          int
	MaxUpdatedUsec int
}

type NewMessageParams struct {
	ThreadId string
	Content  string
	Silent   bool
}

func (q *Client) GetRecentMessages(params *GetRecentMessagesParams) ([]*Message, error) {
	requestParams := make(map[string]string)

	required(params.ThreadId, "ThreadId is required for /messages/thread-id")
	setOptional(params.Count, "count", &requestParams)
	setOptional(params.MaxUpdatedUsec, "max_updated_usec", &requestParams)

	resp, err := q.getJson(q.apiUrlResource("messages/"+params.ThreadId), requestParams)
	if err != nil {
		return nil, err
	}
	var messages []*Message
	if err := json.Unmarshal(resp, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (q *Client) NewMessage(params *NewMessageParams) (*Message, error) {
	requestParams := make(map[string]string)

	setRequired(params.ThreadId, "thread_id", &requestParams, "ThreadID is required for /messages/new")
	setRequired(params.Content, "content", &requestParams, "Content is required for /messages/new")

	setOptional(params.Silent, "silent", &requestParams)

	resp, err := q.postJson(q.apiUrlResource("messages/new"), requestParams)
	if err != nil {
		return nil, err
	}
	var message Message
	if err := json.Unmarshal(resp, &message); err != nil {
		return nil, err
	}
	return &message, nil
}

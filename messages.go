package quip

import "github.com/mitchellh/mapstructure"

type Message struct {
	AuthorId    string `mapstructure:"author_id"`
	CreatedUsec int    `mapstructure:"created_usec"`
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

func (q *Client) GetRecentMessages(params *GetRecentMessagesParams) []*Message {
	requestParams := make(map[string]string)

	required(params.ThreadId, "ThreadId is required for /messages/thread-id")
	setOptional(params.Count, "count", &requestParams)
	setOptional(params.MaxUpdatedUsec, "max_updated_usec", &requestParams)

	resp := q.getJson(apiUrlResource("messages/"+params.ThreadId), requestParams)
	parsed := parseJsonArray(resp)

	return hydrateMessages(parsed)
}

func (q *Client) NewMessage(params *NewMessageParams) *Message {
	requestParams := make(map[string]string)

	setRequired(params.ThreadId, "thread_id", &requestParams, "ThreadID is required for /messages/new")
	setRequired(params.Content, "content", &requestParams, "Content is required for /messages/new")

	setOptional(params.Silent, "silent", &requestParams)

	resp := q.postJson(apiUrlResource("messages/new"), requestParams)
	parsed := parseJsonObject(resp)

	return hydrateMessage(parsed)
}

func hydrateMessage(resp interface{}) *Message {
	var message Message
	mapstructure.Decode(resp, &message)
	return &message
}

func hydrateMessages(resp []interface{}) []*Message {
	messages := make([]*Message, 0, len(resp))

	for _, body := range resp {
		messages = append(messages, hydrateMessage(body))
	}

	return messages
}

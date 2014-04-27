package quip

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Thread struct {
	ExpandedUserIds []string `mapstructure:"expanded_user_ids"`
	UserIds         []string `mapstructure:"user_ids"`
	SharedFolderIds []string `mapstructure:"shared_folder_ids"`
	Html            string
	Thread          map[string]string
}

type GetRecentThreadsParams struct {
	Count          int
	MaxUpdatedUsec int
}

type NewDocumentParams struct {
	Content   string
	Format    string
	Title     string
	MemberIds []string
}

type EditDocumentParams struct {
	ThreadId  string
	Content   string
	Format    string
	Location  string
	MemberIds []string
}

type AddMembersParams struct {
	ThreadId  string
	MemberIds []string
}

type RemoveMembersParams struct {
	ThreadId  string
	MemberIds []string
}

func (q *Client) GetThread(id string) *Thread {
	resp := q.getJson(apiUrlResource("threads/"+id), map[string]string{})
	parsed := parseJsonObject(resp)
	return hydrateThread(parsed)
}

func (q *Client) GetThreads(ids []string) []*Thread {
	qid := strings.Join(ids, ",")
	resp := q.getJson(apiUrlResource("threads/?ids="+qid), map[string]string{})
	parsed := parseJsonObject(resp)
	return hydrateThreads(parsed)
}

func (q *Client) GetRecentThreads(params *GetRecentThreadsParams) []*Thread {
	requestParams := make(map[string]string)

	setOptional(params.Count, "count", &requestParams)
	setOptional(params.MaxUpdatedUsec, "max_updated_usec", &requestParams)

	resp := q.getJson(apiUrlResource("threads/recent"), requestParams)
	parsed := parseJsonObject(resp)
	return hydrateThreads(parsed)
}

func (q *Client) NewDocument(params *NewDocumentParams) *Thread {
	requestParams := make(map[string]string)

	setRequired(params.Content, "content", &requestParams, "Content is required for /new-document")
	setOptional(params.Format, "format", &requestParams)
	setOptional(params.Title, "title", &requestParams)
	setOptional(strings.Join(params.MemberIds, ","), "member_ids", &requestParams)

	fmt.Println(requestParams)
	resp := q.postJson(apiUrlResource("threads/new-document"), requestParams)
	parsed := parseJsonObject(resp)
	return hydrateThread(parsed)
}

func (q *Client) EditDocument(params *EditDocumentParams) *Thread {
	requestParams := make(map[string]string)
	required(params.Content, "Content is required for /edit-document")
	requestParams["content"] = params.Content

	setOptional(params.Format, "format", &requestParams)
	setOptional(params.Location, "locatoin", &requestParams)
	setOptional(strings.Join(params.MemberIds, ","), "member_ids", &requestParams)

	resp := q.postJson(apiUrlResource("threads/edit-document"), requestParams)

	return hydrateThread(resp)
}

func (q *Client) AddMembers(params *AddMembersParams) *Thread {
	requestParams := make(map[string]string)
	required(params.ThreadId, "ThreadId is required for /add-members")
	required(params.MemberIds, "MemberIds is required for /add-members")

	requestParams["thread_id"] = params.ThreadId
	requestParams["member_ids"] = strings.Join(params.MemberIds, ",")

	resp := q.postJson(apiUrlResource("threads/add-members"), requestParams)
	parsed := parseJsonObject(resp)
	return hydrateThread(parsed)
}

func (q *Client) RemoveMembers(params *RemoveMembersParams) *Thread {
	requestParams := make(map[string]string)
	required(params.ThreadId, "ThreadId is required for /remove-members")
	required(params.MemberIds, "MemberIds is required for /remove-members")

	requestParams["thread_id"] = params.ThreadId
	requestParams["member_ids"] = strings.Join(params.MemberIds, ",")

	resp := q.postJson(apiUrlResource("threads/remove-members"), requestParams)
	parsed := parseJsonObject(resp)
	return hydrateThread(parsed)
}

func hydrateThread(resp interface{}) *Thread {
	var thread Thread
	mapstructure.Decode(resp, &thread)
	return &thread
}

func hydrateThreads(resp map[string]interface{}) []*Thread {
	threads := make([]*Thread, 0, len(resp))

	for _, body := range resp {
		threads = append(threads, hydrateThread(body))
	}

	return threads
}

package quip

import (
	"encoding/json"
	"strings"
)

const (
	APPEND          = "0"
	PREPEND         = "1"
	AFTER_SECTION   = "2"
	BEFORE_SECTION  = "3"
	REPLACE_SECTION = "4"
	DELETE_SECTION  = "5"
)

type ThreadSharing struct {
	CompanyID   string `json:"company_id"`
	CompanyMode string `json:"company_mode"`
}

type ThreadDetails struct {
	ID          string
	AuthorID    string `json:"author_id"`
	ThreadClass string `json:"thread_class"`
	Created     int64  `json:"created_usec"`
	Updated     int64  `json:"updated_usec"`
	Title       string
	Link        string
	Type        string
	Sharing     ThreadSharing
}

type Thread struct {
	ExpandedUserIds []string `json:"expanded_user_ids"`
	UserIds         []string `json:"user_ids"`
	SharedFolderIds []string `json:"shared_folder_ids"`
	Html            string
	Thread          ThreadDetails
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
	SectionId string
}

type AddMembersParams struct {
	ThreadId  string
	MemberIds []string
}

type RemoveMembersParams struct {
	ThreadId  string
	MemberIds []string
}

func (q *Client) GetThread(id string) (*Thread, error) {
	resp, err := q.getJson(q.apiUrlResource("threads/"+id), map[string]string{})
	if err != nil {
		return nil, err
	}
	var thread Thread
	if err := json.Unmarshal(resp, &thread); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (q *Client) GetThreads(ids []string) ([]*Thread, error) {
	var threads []*Thread

	if len(ids) == 0 {
		return threads, nil
	}

	resp, err := q.getJson(q.apiUrlResource("threads/"), map[string]string{
		"ids": strings.Join(ids, ","),
	})
	if err != nil {
		return nil, err
	}

	var threadMap map[string]*Thread
	if err := json.Unmarshal(resp, &threadMap); err != nil {
		return nil, err
	}
	for _, t := range threadMap {
		threads = append(threads, t)
	}

	return threads, nil
}

func (q *Client) GetRecentThreads(params *GetRecentThreadsParams) ([]*Thread, error) {
	requestParams := make(map[string]string)

	setOptional(params.Count, "count", &requestParams)
	setOptional(params.MaxUpdatedUsec, "max_updated_usec", &requestParams)

	resp, err := q.getJson(q.apiUrlResource("threads/recent"), requestParams)
	if err != nil {
		return nil, err
	}
	var threads []*Thread
	if err := json.Unmarshal(resp, &threads); err != nil {
		return nil, err
	}
	return threads, nil
}

func (q *Client) NewDocument(params *NewDocumentParams) (*Thread, error) {
	requestParams := make(map[string]string)

	setRequired(params.Content, "content", &requestParams, "Content is required for /new-document")
	setOptional(params.Format, "format", &requestParams)
	setOptional(params.Title, "title", &requestParams)
	setOptional(strings.Join(params.MemberIds, ","), "member_ids", &requestParams)

	resp, err := q.postJson(q.apiUrlResource("threads/new-document"), requestParams)
	if err != nil {
		return nil, err
	}
	var thread Thread
	if err := json.Unmarshal(resp, &thread); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (q *Client) EditDocument(params *EditDocumentParams) (*Thread, error) {
	requestParams := make(map[string]string)
	setRequired(params.Content, "content", &requestParams, "Content is required for /edit-document")
	setRequired(params.ThreadId, "thread_id", &requestParams, "Thread ID is required for /edit-document")

	setOptional(params.Format, "format", &requestParams)
	setOptional(params.Location, "location", &requestParams)
	setOptional(params.SectionId, "section_id", &requestParams)

	resp, err := q.postJson(q.apiUrlResource("threads/edit-document"), requestParams)
	if err != nil {
		return nil, err
	}
	var thread Thread
	if err := json.Unmarshal(resp, &thread); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (q *Client) AddMembers(params *AddMembersParams) (*Thread, error) {
	requestParams := make(map[string]string)
	required(params.ThreadId, "ThreadId is required for /add-members")
	required(params.MemberIds, "MemberIds is required for /add-members")

	requestParams["thread_id"] = params.ThreadId
	requestParams["member_ids"] = strings.Join(params.MemberIds, ",")

	resp, err := q.postJson(q.apiUrlResource("threads/add-members"), requestParams)
	if err != nil {
		return nil, err
	}
	var thread Thread
	if err := json.Unmarshal(resp, &thread); err != nil {
		return nil, err
	}
	return &thread, nil
}

func (q *Client) RemoveMembers(params *RemoveMembersParams) (*Thread, error) {
	requestParams := make(map[string]string)
	required(params.ThreadId, "ThreadId is required for /remove-members")
	required(params.MemberIds, "MemberIds is required for /remove-members")

	requestParams["thread_id"] = params.ThreadId
	requestParams["member_ids"] = strings.Join(params.MemberIds, ",")

	resp, err := q.postJson(q.apiUrlResource("threads/remove-members"), requestParams)
	if err != nil {
		return nil, err
	}
	var thread Thread
	if err := json.Unmarshal(resp, &thread); err != nil {
		return nil, err
	}
	return &thread, nil
}

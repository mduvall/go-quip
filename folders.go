package quip

import (
	"encoding/json"
	"strings"
)

type Folder struct {
	Info struct {
		Id          string
		Title       string
		CreatedUsec int `json:"created_usec"`
		UpdatedUsec int `json:"updated_usec"`
		Color       string
		ParentId    string `json:"parent_id"`
	} `json:"folder"`

	MemberIds []string `json:"member_ids"`
	Children  []map[string]string
}

type GetFolderParams struct {
	Id string
}

type GetFoldersParams struct {
	Ids []string
}

type AddFolderMembersParams struct {
	FolderId  string
	MemberIds []string
}

type RemoveFolderMembersParams struct {
	FolderId  string
	MemberIds []string
}

type NewFolderParams struct {
	Title     string
	ParentId  string
	Color     int
	MemberIds []string
}

func (q *Client) GetFolder(params *GetFolderParams) (*Folder, error) {
	required(params.Id, "Id is required for /folder/id")

	resp, err := q.getJson(q.apiUrlResource("folders/"+params.Id), map[string]string{})
	if err != nil {
		return nil, err
	}
	var folder Folder
	if err := json.Unmarshal(resp, &folder); err != nil {
		return nil, err
	}
	return &folder, nil
}

func (q *Client) GetFolders(params *GetFoldersParams) ([]*Folder, error) {
	required(params.Ids, "Ids is required for /folder/ids")

	resp, _ := q.getJson(q.apiUrlResource("folders/"+strings.Join(params.Ids, ",")), map[string]string{})
	var folders []*Folder
	if err := json.Unmarshal(resp, folders); err != nil {
		return nil, err
	}
	return folders, nil
}

func (q *Client) NewFolder(params *NewFolderParams) (*Folder, error) {
	requestParams := make(map[string]string)
	setRequired(params.Title, "title", &requestParams, "Title is required for /folders/new")
	setOptional(params.ParentId, "parent_id", &requestParams)
	setOptional(params.Color, "color", &requestParams)
	setOptional(params.MemberIds, "member_ids", &requestParams)

	resp, err := q.postJson(q.apiUrlResource("folders/new"), requestParams)
	if err != nil {
		return nil, err
	}
	var folder Folder
	if err := json.Unmarshal(resp, &folder); err != nil {
		return nil, err
	}
	return &folder, err
}

func (q *Client) AddFolderMembers(params *AddFolderMembersParams) (*Folder, error) {
	requestParams := make(map[string]string)
	setRequired(params.FolderId, "folder_id", &requestParams, "FolderId is required for /folder/add-members")
	setRequired(params.MemberIds, "member_ids", &requestParams, "MemberIds is required for /folder/add-members")

	resp, err := q.postJson(q.apiUrlResource("folders/add-members"), requestParams)
	if err != nil {
		return nil, err
	}
	var folder Folder
	if err := json.Unmarshal(resp, &folder); err != nil {
		return nil, err
	}
	return &folder, err
}

func (q *Client) RemoveFolderMembers(params *RemoveFolderMembersParams) (*Folder, error) {
	requestParams := make(map[string]string)
	setRequired(params.FolderId, "folder_id", &requestParams, "FolderId is required for /folder/remove-members")
	setRequired(params.MemberIds, "member_ids", &requestParams, "MemberIds is required for /folder/remove-members")

	resp, err := q.postJson(q.apiUrlResource("folders/remove-members"), requestParams)
	if err != nil {
		return nil, err
	}
	var folder Folder
	if err := json.Unmarshal(resp, &folder); err != nil {
		return nil, err
	}
	return &folder, err
}

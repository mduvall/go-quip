package quip

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Folder struct {
	Info struct {
		Id          string
		Title       string
		CreatedUsec int `mapstructure:"created_usec"`
		UpdatedUsec int `mapstructure:"updated_usec"`
		Color       int
		ParentId    string `mapstructure:"parent_id"`
	} `mapstructure:"folder"`

	MemberIds []string `mapstructure:"member_ids"`
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

func (q *Client) GetFolder(params *GetFolderParams) *Folder {
	required(params.Id, "Id is required for /folder/id")

	resp := q.getJson(apiUrlResource("folders/"+params.Id), map[string]string{})
	parsed := parseJsonObject(resp)

	return hydrateFolder(parsed)
}

func (q *Client) GetFolders(params *GetFoldersParams) []*Folder {
	required(params.Ids, "Ids is required for /folder/ids")

	resp := q.getJson(apiUrlResource("folders/"+strings.Join(params.Ids, ",")), map[string]string{})
	parsed := parseJsonObject(resp)

	return hydrateFolders(parsed)
}

func (q *Client) NewFolder(params *NewFolderParams) *Folder {
	requestParams := make(map[string]string)
	setRequired(params.Title, "title", &requestParams, "Title is required for /folders/new")
	setOptional(params.ParentId, "parent_id", &requestParams)
	setOptional(params.Color, "color", &requestParams)
	setOptional(params.MemberIds, "member_ids", &requestParams)

	resp := q.postJson(apiUrlResource("folders/new"), requestParams)
	parsed := parseJsonObject(resp)

	fmt.Println(string(resp))

	return hydrateFolder(parsed)
}

func (q *Client) AddFolderMembers(params *AddFolderMembersParams) *Folder {
	requestParams := make(map[string]string)
	setRequired(params.FolderId, "folder_id", &requestParams, "FolderId is required for /folder/add-members")
	setRequired(params.MemberIds, "member_ids", &requestParams, "MemberIds is required for /folder/add-members")

	resp := q.postJson(apiUrlResource("folders/add-members"), requestParams)
	parsed := parseJsonObject(resp)

	return hydrateFolder(parsed)
}

func (q *Client) RemoveFolderMembers(params *RemoveFolderMembersParams) *Folder {
	requestParams := make(map[string]string)
	setRequired(params.FolderId, "folder_id", &requestParams, "FolderId is required for /folder/remove-members")
	setRequired(params.MemberIds, "member_ids", &requestParams, "MemberIds is required for /folder/remove-members")

	resp := q.postJson(apiUrlResource("folders/remove-members"), requestParams)
	parsed := parseJsonObject(resp)

	return hydrateFolder(parsed)
}

func hydrateFolder(resp interface{}) *Folder {
	var folder Folder
	mapstructure.Decode(resp, &folder)
	return &folder
}

func hydrateFolders(resp map[string]interface{}) []*Folder {
	folders := make([]*Folder, 0, len(resp))

	for _, body := range resp {
		folders = append(folders, hydrateFolder(body))
	}

	return folders
}

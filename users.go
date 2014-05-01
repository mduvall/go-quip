package quip

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

type User struct {
	Id              string
	Name            string
	Affinity        float64
	ChatThreadId    string `mapstructure:"chat_thread_id"`
	DesktopFolderId string `mapstructure:"desktop_folder_id"`
	ArchiveFolderId string `mapstructure:"archive_folder_id"`
}

type GetUserParams struct {
	Id string
}

type GetUsersParams struct {
	Ids []string
}

func (q *Client) GetUser(params *GetUserParams) *User {
	required(params.Id, "Id is required for /users/id")

	resp := q.getJson(apiUrlResource("users/"+params.Id), map[string]string{})
	parsed := parseJsonObject(resp)

	return hydrateUser(parsed)
}

func (q *Client) GetUsers(params *GetUsersParams) []*User {
	required(params.Ids, "Ids is required for /users/ids")

	resp := q.getJson(apiUrlResource("users/"+strings.Join(params.Ids, ",")), map[string]string{})
	parsed := parseJsonObject(resp)

	return hydrateUsersMap(parsed)
}

func (q *Client) GetContacts() []*User {
	resp := q.getJson(apiUrlResource("users/contacts"), map[string]string{})
	parsed := parseJsonArray(resp)

	return hydrateUsersArray(parsed)
}

func (q *Client) GetAuthenticatedUser() *User {
	resp := q.getJson(apiUrlResource("users/current"), map[string]string{})
	parsed := parseJsonObject(resp)

	return hydrateUser(parsed)
}

func hydrateUser(resp interface{}) *User {
	var user User
	mapstructure.Decode(resp, &user)
	return &user
}

func hydrateUsersMap(resp map[string]interface{}) []*User {
	users := make([]*User, 0, len(resp))

	for _, body := range resp {
		users = append(users, hydrateUser(body))
	}

	return users
}

func hydrateUsersArray(resp []interface{}) []*User {
	users := make([]*User, 0, len(resp))

	for _, body := range resp {
		users = append(users, hydrateUser(body))
	}

	return users
}

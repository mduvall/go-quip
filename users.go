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

func (q *Client) GetUser(params *GetUserParams) (*User, error) {
	required(params.Id, "Id is required for /users/id")

	resp, err := q.getJson(q.apiUrlResource("users/"+params.Id), map[string]string{})
	if err != nil {
		return nil, err
	}
	parsed, err := parseJsonObject(resp)
	if err != nil {
		return nil, err
	}

	return hydrateUser(parsed)
}

func (q *Client) GetUsers(params *GetUsersParams) ([]*User, error) {
	required(params.Ids, "Ids is required for /users/ids")

	resp, err := q.getJson(q.apiUrlResource("users/"+strings.Join(params.Ids, ",")), map[string]string{})
	if err != nil {
		return nil, err
	}
	parsed, err := parseJsonObject(resp)
	if err != nil {
		return nil, err
	}

	return hydrateUsersMap(parsed)
}

func (q *Client) GetContacts() ([]*User, error) {
	resp, err := q.getJson(q.apiUrlResource("users/contacts"), map[string]string{})
	if err != nil {
		return nil, err
	}
	parsed, err := parseJsonArray(resp)
	if err != nil {
		return nil, err
	}

	return hydrateUsersArray(parsed)
}

func (q *Client) GetAuthenticatedUser() (*User, error) {
	resp, err := q.getJson(q.apiUrlResource("users/current"), map[string]string{})
	if err != nil {
		return nil, err
	}
	parsed, err := parseJsonObject(resp)
	if err != nil {
		return nil, err
	}

	return hydrateUser(parsed)
}

func hydrateUser(resp interface{}) (*User, error) {
	var user User
	if err := mapstructure.Decode(resp, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func hydrateUsersMap(resp map[string]interface{}) ([]*User, error) {
	users := make([]*User, 0, len(resp))

	for _, body := range resp {
		u, err := hydrateUser(body)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func hydrateUsersArray(resp []interface{}) ([]*User, error) {
	users := make([]*User, 0, len(resp))

	for _, body := range resp {
		u, err := hydrateUser(body)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

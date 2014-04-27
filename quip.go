package quip

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	BASE_API_URL = "https://platform.quip.com"
)

type Client struct {
	accessToken  string
	clientId     string
	clientSecret string
	redirectUri  string
}

func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
	}
}

func NewClientOAuth(accessToken string, clientId string, clientSecret string, redirectUri string) *Client {
	// TODO: need to get domain authentication to give this a swing
	return &Client{
		accessToken:  accessToken,
		clientId:     clientId,
		clientSecret: clientSecret,
		redirectUri:  redirectUri,
	}
}

func (q *Client) postJson(resource string, params map[string]string) []byte {
	req, err := http.NewRequest("POST", resource, mapToQueryString(params))
	if err != nil {
		log.Fatal("Bad url: " + resource)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return q.doRequest(req)
}

func (q *Client) getJson(resource string, params map[string]string) []byte {
	qs, err := ioutil.ReadAll(mapToQueryString(params))
	if err != nil {
		log.Fatal("Malformed query params %v", params)
	}

	queryString := string(qs)
	if queryString != "" {
		resource = resource + "?" + queryString
	}

	req, err := http.NewRequest("GET", resource, nil)
	if err != nil {
		log.Fatal("Bad url: " + resource)
	}

	return q.doRequest(req)
}

func (q *Client) doRequest(req *http.Request) []byte {
	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+q.accessToken)
	res, err := client.Do(req)
	if err != nil {
		// TODO: handle API errors here
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return body
}

func mapToQueryString(params map[string]string) *strings.Reader {
	body := url.Values{}
	for k, v := range params {
		body.Set(k, v)
	}
	return strings.NewReader(body.Encode())
}

func apiUrlResource(resource string) string {
	return BASE_API_URL + "/1/" + resource
}

func required(val interface{}, message string) {
	switch val := val.(type) {
	case string:
		if val == "" {
			log.Fatal(message)
		}
	case []string:
		if len(val) == 0 {
			log.Fatal(message)
		}
	}
}

func setOptional(val interface{}, key string, params *map[string]string) {
	switch val := val.(type) {
	case string:
		if val != "" {
			(*params)[key] = val
		}
	case []string:
		if len(val) != 0 {
			(*params)[key] = strings.Join(val, ",")
		}
	}
}

func setRequired(val interface{}, key string, params *map[string]string, message string) {
	required(val, message)
	setOptional(val, key, params)
}

func parseJsonObject(b []byte) map[string]interface{} {
	var val map[string]interface{}
	json.Unmarshal(b, &val)
	return val
}

func parseJsonArray(b []byte) []interface{} {
	var val []interface{}
	json.Unmarshal(b, &val)
	return val
}

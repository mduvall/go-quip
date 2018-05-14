package quip

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	accessToken  string
	clientId     string
	clientSecret string
	redirectUri  string
	apiUrl       string
}

func NewClient(accessToken string) *Client {
	return &Client{
		accessToken: accessToken,
		apiUrl:      "https://platform.quip.com",
	}
}

func NewClientOAuth(accessToken string, clientId string, clientSecret string, redirectUri string) *Client {
	// TODO: need to get domain authentication to give this a swing
	return &Client{
		accessToken:  accessToken,
		clientId:     clientId,
		clientSecret: clientSecret,
		redirectUri:  redirectUri,
		apiUrl:       "https://platform.quip.com",
	}
}

func (q *Client) SetApiUrl(url string) {
	q.apiUrl = url
}

func (q *Client) postJson(resource string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", resource, mapToQueryString(params))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return q.doRequest(req, 1)
}

func (q *Client) getJson(resource string, params map[string]string) ([]byte, error) {
	qs, err := ioutil.ReadAll(mapToQueryString(params))
	if err != nil {
		return nil, err
	}

	queryString := string(qs)
	if queryString != "" {
		resource = resource + "?" + queryString
	}

	req, err := http.NewRequest("GET", resource, nil)
	if err != nil {
		return nil, err
	}

	return q.doRequest(req, 1)
}

func (q *Client) doRequest(req *http.Request, attempt int) ([]byte, error) {

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+q.accessToken)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if retryDelay := retryAfterDelay(req, res); retryDelay > 0 {

		// Bail out, don't loop again
		if attempt > 3 {
			return []byte{}, errors.New("Too many failed HTTP requests")
		}

		log.Printf("Delaying for %s due to rate limit\n", retryDelay.Round(time.Second).String())
		time.Sleep(retryDelay)
		return q.doRequest(req, attempt+1)
	}

	if res.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("%s, in response to %s %s", res.Status, req.Method, req.URL)
	}

	return ioutil.ReadAll(res.Body)
}

// Quip API docs don't mention rate limiting currently. This code is based on
// retry_rate_limit in https://github.com/ConsenSys/quip-projects/blob/master/quip.py
// where Quip uses a 503 with a X-RateLimit-Reset header (not a 429 with Retry-After)
// "250 calls per hour/15 per minute" https://twitter.com/QuipSupport/status/527965994761744384
// Seems to give a 15s wait first then 44s waits for every ~48 rapid requests.
func retryAfterDelay(req *http.Request, res *http.Response) time.Duration {

	hdr := ""
	switch res.StatusCode {
	case http.StatusServiceUnavailable:
		hdr = res.Header.Get("X-RateLimit-Reset")
	case http.StatusTooManyRequests:
		hdr = res.Header.Get("Retry-After")
	default:
		return 0
	}

	var idempotent bool // idempotent methods we can safely retry
	switch req.Method {
	case "GET", "HEAD", "PUT", "DELETE", "OPTIONS", "TRACE":
		idempotent = true
	}

	delay, err := strconv.ParseInt(hdr, 10, 64)
	if err != nil { // missing retry delay header or a bad value
		if !idempotent {
			return 0 // not safe to retry
		}
		return 5 * time.Second // default retry delay for idempotent methods
	}

	// if we got a valid rate limit delay value then we trust that it's ok to retry
	// (Quip sends a 503 with X-RateLimit-Reset on a POST)
	if idempotent || req.Method == "POST" {
		return time.Duration(delay) * time.Second
	}

	return 0
}

func mapToQueryString(params map[string]string) *strings.Reader {
	body := url.Values{}
	for k, v := range params {
		body.Set(k, v)
	}
	return strings.NewReader(body.Encode())
}

func (q *Client) apiUrlResource(resource string) string {
	return q.apiUrl + "/1/" + resource
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

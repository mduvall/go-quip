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

	"github.com/beefsack/go-rate"
)

type Client struct {
	accessToken  string
	clientId     string
	clientSecret string
	redirectUri  string
	apiUrl       string
	throttle     *rate.RateLimiter
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

func (q *Client) Throttle(interval time.Duration) {
	if interval == 0 {
		q.throttle = nil
		return
	}
	q.throttle = rate.New(1, interval)
}

func (q *Client) postJson(resource string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("POST", resource, mapToQueryString(params))
	if err != nil {
		log.Fatal("Bad url: " + resource)
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

	// Bail out, don't loop
	if attempt > 3 {
		return []byte{}, errors.New("Too many failed HTTP requests")
	}

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+q.accessToken)

	if q.throttle != nil {
		q.throttle.Wait()
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 503 {
		header := res.Header.Get("X-RateLimit-Reset")
		if len(header) == 0 {
			return []byte{}, errors.New("Got 503, but no X-RateLimit-Reset header?!")
		}
		timestamp, err := strconv.ParseInt(header, 10, 64)
		if err != nil {
			return []byte{}, fmt.Errorf("Failed to convert %s to int64", err)
		}
		t := time.Unix(timestamp, 0).Local()
		until := time.Until(t)
		log.Printf("Asked to halt for %s...\n", until.Round(time.Second).String())
		time.Sleep(until)
		attempt = attempt + 1
		q.doRequest(req, attempt)
	}

	if res.StatusCode >= 400 {
		return []byte{}, fmt.Errorf("%s, in response to %s %s", res.Status, req.Method, req.URL)
	}

	return ioutil.ReadAll(res.Body)
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

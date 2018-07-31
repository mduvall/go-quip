package quip

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	accessToken       string
	clientId          string
	clientSecret      string
	redirectUri       string
	apiUrl            string
	maxRateLimitDelay int64
	// Number of seconds spent waiting due to rate limiting
	RateLimitDelays float64
}

func NewClient(accessToken string) *Client {
	return &Client{
		accessToken:       accessToken,
		apiUrl:            "https://platform.quip.com",
		maxRateLimitDelay: 60, // doubles each time a higher rate limit is hit
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
		return nil, errors.Wrapf(err, "http.NewRequest POST %s", resource)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return q.doRequest(req, 1)
}

func (q *Client) getJson(resource string, params map[string]string) ([]byte, error) {
	qs, err := ioutil.ReadAll(mapToQueryString(params))
	if err != nil {
		return nil, errors.Wrapf(err, "ioutil.ReadAll")
	}

	queryString := string(qs)
	if queryString != "" {
		resource = resource + "?" + queryString
	}

	req, err := http.NewRequest("GET", resource, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "http.NewRequest GET %s", resource)
	}

	return q.doRequest(req, 1)
}

func (q *Client) doRequest(req *http.Request, attempt int) ([]byte, error) {

	client := &http.Client{}
	req.Header.Set("Authorization", "Bearer "+q.accessToken)

	errWrap := func(err error) error {
		return errors.Wrapf(err, "client.Do %s %s", req.Method, req.URL.String())
	}

	res, err := client.Do(req)
	if err != nil {
		return []byte{}, errWrap(err)
	}
	defer res.Body.Close()

	if retryDelay := q.retryAfterDelay(req, res); retryDelay > 0 {

		// Bail out, don't recurse again
		if attempt > 3 {
			return []byte{}, errWrap(errors.New("Too many failed HTTP requests: " + res.Status))
		}

		log.Printf("Delaying for %s due to rate limit.\n", retryDelay.Round(time.Second).String())
		time.Sleep(retryDelay)
		q.RateLimitDelays += retryDelay.Seconds()
		return q.doRequest(req, attempt+1) // recurse
	}

	if res.StatusCode >= 400 {
		return []byte{}, errWrap(errors.New(res.Status))
	}

	return ioutil.ReadAll(res.Body)
}

// Quip API docs don't mention rate limiting currently. This code is based on
// retry_rate_limit in https://github.com/ConsenSys/quip-projects/blob/master/quip.py
// where Quip uses a 503 with a X-RateLimit-Reset header (not a 429 with Retry-After)
// "250 calls per hour/15 per minute" https://twitter.com/QuipSupport/status/527965994761744384
// Seems to give a 15s wait first then 44s waits for every ~48 rapid requests.
func (q *Client) retryAfterDelay(req *http.Request, res *http.Response) time.Duration {

	hdr := ""
	switch res.StatusCode {
	case http.StatusServiceUnavailable:
		hdr = res.Header.Get("X-RateLimit-Reset")
		//log.Printf("X-RateLimit-Reset: %s", hdr)
	case http.StatusTooManyRequests:
		hdr = res.Header.Get("Retry-After")
		//log.Printf("Retry-After: %s", hdr)
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

	// Quip used to return X-RateLimit-Reset with a small value
	// representing the number of seconds to delay. It later
	// switched to returning the more common UTC epoch time.
	if delay >= 1000000000 { // looks like an epoch
		delay = delay - time.Now().Unix() // convert epoch to delay
	}

	if delay > q.maxRateLimitDelay {
		delay = q.maxRateLimitDelay                   // be sensible
		q.maxRateLimitDelay = q.maxRateLimitDelay * 2 // be kind
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

func parseJsonObject(b []byte) (map[string]interface{}, error) {
	var val map[string]interface{}
	if err := unmarshal(b, &val); err != nil {
		return nil, err
	}
	return val, nil
}

func parseJsonArray(b []byte) ([]interface{}, error) {
	var val []interface{}
	if err := unmarshal(b, &val); err != nil {
		return nil, err
	}
	return val, nil
}

// unmarshal calls json.Unmarshal and gives richer error reports including the
// type being unmarshalled and at least some of the raw data bytes.
// Set the QUIP_API_ERR_DUMP_LEN env var to adjust the amount of data shown.
func unmarshal(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		dumpLen, _ := strconv.Atoi(os.Getenv("QUIP_API_ERR_DUMP_LEN"))
		if dumpLen == 0 {
			dumpLen = 20
		}
		suffix := ""
		if len(data) > dumpLen {
			suffix = fmt.Sprintf("... (showing first %d bytes, set QUIP_API_ERR_DUMP_LEN env var to adjust)", dumpLen)
		}
		return errors.Wrapf(err, "json.Unmarshal %T from %d bytes: %.*s%s", v, len(data), dumpLen, data, suffix)
	}
	return nil
}

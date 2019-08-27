package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// JSON is the expected data format of requests and responses
const JSON = "application/json"

// A Client is a client for the ndau REST API.
type Client struct {
	addr  *url.URL
	mutex sync.Mutex
	http  *http.Client
}

// NewClient creates a SDKClient.
func NewClient(node string) (*Client, error) {
	u, err := url.Parse(node)
	if err != nil {
		return nil, errors.Wrap(err, "parsing node address")
	}
	u.Path = ""
	return &Client{
		addr: u,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// SetTimeout updates the node's http timeout
//
// The default is 5 seconds
func (c *Client) SetTimeout(timeout time.Duration) {
	c.http.Timeout = timeout
}

// URL constructs a URL from a path
//
// It constructs the path from the supplied path and arguments using fmt.Sprintf.
func (c *Client) URL(path string, args ...interface{}) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// ensure we reset the base path
	defer func() { c.addr.Path = "" }()

	c.addr.Path = fmt.Sprintf(path, args...)
	return c.addr.String()
}

type params map[string]interface{}

// URLP constructs a URL from a path and adds some query parameters
func (c *Client) URLP(params params, path string, args ...interface{}) string {
	u := c.URL(path, args...)
	qfs := make([]string, 0, len(params))
	for k, v := range params {
		qfs = append(qfs, fmt.Sprintf(
			"%s=%s",
			url.QueryEscape(k),
			url.QueryEscape(fmt.Sprint(v)),
		))
	}
	q := strings.Join(qfs, "&")
	if q != "" {
		u += "?" + q
	}
	return u
}

func (c *Client) get(obj interface{}, url string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "constructing request")
	}
	req.Header.Set("Accept", JSON)
	response, err := c.http.Do(req)
	if err != nil {
		return errors.Wrap(err, "performing request")
	}
	defer response.Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return errors.Wrap(err, "unmarshaling response")
	}
	return nil
}

func (c *Client) post(req interface{}, resp interface{}, url string) error {
	var data []byte
	var err error
	if req != nil {
		data, err = json.Marshal(req)
		if err != nil {
			return errors.Wrap(err, "marshaling request body")
		}
	}
	buf := bytes.NewBuffer(data)
	request, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return errors.Wrap(err, "constructing request")
	}
	request.Header.Set("Content-Type", JSON)
	request.Header.Set("Accept", JSON)
	response, err := c.http.Do(request)
	if err != nil {
		return errors.Wrap(err, "performing request")
	}
	defer response.Body.Close()
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}
	err = json.Unmarshal(data, resp)
	if err != nil {
		return errors.Wrap(err, "unmarshaling response")
	}
	return nil
}

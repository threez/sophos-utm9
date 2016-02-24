package confd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// ConfdConn is the confd connection object
type ConfdConn struct {
	Host     string
	Port     int16
	user     string // currently not used
	passwd   string // currently not used
	facility string // currently not used
	Client   *http.Client
	session  string // currently not used
	id       int64  // json rpc counter
	err      error  // error cache
	mutex    sync.Mutex
}

// ConfdParams can be anything that renders to json
type ConfdParams interface{}

// ConfdResponse is used for custom response handling
// just include the type in your types to handle errors
type ConfdResponse struct {
	Error *string `json:"error"` // pointer since it can be omitted
	Id    int64   `json:"id"`
}

// ConfdSimpleResponse is used for SimpleResult requests, the return
// value is untyped
type ConfdSimpleResponse struct {
	ConfdResponse
	Result interface{} `json:"result"`
}

// ConfdRequest is used to construct requests
type ConfdRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Id     int64       `json:"id"`
}

// NewConfdConn creates a new confd connection (is not acually connecting)
func NewConfdConn() *ConfdConn {
	tr := &http.Transport{
		DisableCompression:  true,
		MaxIdleConnsPerHost: 8,
	}
	return &ConfdConn{
		Host:   "127.0.0.1",
		Port:   4472,
		Client: &http.Client{Transport: tr},
	}
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *ConfdConn) Connect() error {
	c.SimpleRequest("new", nil)
	c.SimpleRequest("get_SID", nil)
	return c.Err()
}

// SimpleRequest sends a simple request (untyped response) to the confd
func (c *ConfdConn) SimpleRequest(method string, params ConfdParams) (interface{}, error) {
	resp := &ConfdSimpleResponse{}
	err := c.Request(method, resp, params)
	return resp.Result, err
}

// Request allows to send request with typed (parsed with json) responses
func (c *ConfdConn) Request(method string, result interface{}, params ConfdParams) error {
	// skip anything on error
	if c.err != nil {
		return c.err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	setAndReturnErr := func(err error) error { c.err = err; return err }

	// request
	data := ConfdRequest{method, params, c.id}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&data)
	c.id++

	// send to remote side and recieve response
	url := fmt.Sprintf("http://%s:%d/", c.Host, c.Port)
	resp, err := c.Client.Post(url, "application/json", &buf)
	if err != nil {
		return setAndReturnErr(err)
	}
	defer resp.Body.Close()

	// decode response
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(result)
	if err != nil {
		return setAndReturnErr(err)
	}

	switch r := result.(type) {
	case ConfdResponse:
		setAndReturnErr(fmt.Errorf("Confd error: %s", *r.Error))
	case ConfdSimpleResponse:
		setAndReturnErr(fmt.Errorf("Confd error: %s", *r.Error))
	}

	return nil
}

// Err returns the last cached error value
func (c *ConfdConn) Err() error {
	return c.err
}

// ResetErr resets the set error value of the connection
func (c *ConfdConn) ResetErr() {
	c.err = nil
}

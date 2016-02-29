package confd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// Conn is the confd connection object
type Conn struct {
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
	write    sync.Mutex // prevent multiple write transactions
	read     sync.Mutex // prevent multiple read transactions
}

// Params can be anything that renders to json
type Params interface{}

// Response is used for custom response handling
// just include the type in your types to handle errors
type Response struct {
	Error *string `json:"error"` // pointer since it can be omitted
	ID    int64   `json:"id"`
}

// SimpleResponse is used for SimpleResult requests, the return
// value is untyped
type SimpleResponse struct {
	Response
	Result interface{} `json:"result"`
}

// Request is used to construct requests
type Request struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	ID     int64       `json:"id"`
}

// NewConn creates a new confd connection (is not acually connecting)
func NewConn() *Conn {
	tr := &http.Transport{
		DisableCompression:  true,
		MaxIdleConnsPerHost: 8,
	}
	return &Conn{
		Host:   "127.0.0.1",
		Port:   4472,
		Client: &http.Client{Transport: tr},
	}
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *Conn) Connect() error {
	c.SimpleRequest("new", nil)
	c.SimpleRequest("get_SID", nil)
	return c.Err()
}

// SimpleRequest sends a simple request (untyped response) to the confd
func (c *Conn) SimpleRequest(method string, params Params) (interface{}, error) {
	resp := &SimpleResponse{}
	err := c.Request(method, resp, params)
	return resp.Result, err
}

// Request allows to send request with typed (parsed with json) responses
func (c *Conn) Request(method string, result interface{}, params Params) error {
	// skip anything on error
	if c.err != nil {
		return c.err
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	setAndReturnErr := func(err error) error { c.err = err; return err }

	// request
	data := Request{method, params, c.id}
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
	case Response:
		setAndReturnErr(fmt.Errorf("Confd error: %s", *r.Error))
	case SimpleResponse:
		setAndReturnErr(fmt.Errorf("Confd error: %s", *r.Error))
	}

	return nil
}

// Err returns the last cached error value
func (c *Conn) Err() error {
	return c.err
}

// ResetErr resets the set error value of the connection
func (c *Conn) ResetErr() {
	c.err = nil
}

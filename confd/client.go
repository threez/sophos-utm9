package confd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// LocalConnection is used on the box
const LocalConnection = "127.0.0.1:4472"

// Conn is the confd connection object
type Conn struct {
	conn    *net.TCPConn
	URL     string
	id      int64 // json rpc counter
	err     error // error cache
	Timeout time.Duration
	mutex   sync.Mutex
	write   sync.Mutex // prevent multiple write transactions
	read    sync.Mutex // prevent multiple read transactions
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
func NewConn(URL string) *Conn {
	return &Conn{
		URL:     URL,
		Timeout: time.Second * 5,
	}
}

// NewDefaultConn creates a new confd connection (is not acually connecting) to
// http://127.0.0.1:4472/ (LocalConnection)
func NewDefaultConn() *Conn {
	return NewConn(LocalConnection)
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
	setAndReturnErr := func(err error) error {
		if c.conn != nil {
			c.conn.Close()
			c.conn = nil
		}
		c.err = err
		return err
	}

	// request
	data := Request{method, params, c.id}
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&data)
	c.id++

	// connect to server if not done yet
	if c.conn == nil {
		conn, err := net.Dial("tcp", c.URL)
		if err != nil {
			return setAndReturnErr(err)
		}
		c.conn = conn.(*net.TCPConn)
		c.connect() // initalize the session
	}

	// send request
	req := fmt.Sprintf("POST / HTTP/1.1\r\n"+
		"Content-Type: application/json\r\n"+
		"Content-Length: %d\r\n\r\n%s", buf.Len(), buf.String())

	// send to remote side and recieve response
	c.mutex.Lock()
	_, err = c.conn.Write([]byte(req[:]))
	if err != nil {
		return setAndReturnErr(err)
	}

	err = c.conn.SetReadDeadline(time.Now().Add(c.Timeout))
	if err != nil {
		return setAndReturnErr(err)
	}

	// read response
	reader := bufio.NewReader(c.conn)
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return setAndReturnErr(err)
	}
	c.mutex.Unlock()

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

// Close the confd connection
func (c *Conn) Close() error {
	c.SimpleRequest("disconnect", nil)
	c.err = c.conn.Close()
	return c.Err()
}

// Err returns the last cached error value
func (c *Conn) Err() error {
	return c.err
}

// ResetErr resets the set error value of the connection
func (c *Conn) ResetErr() {
	c.err = nil
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *Conn) connect() error {
	c.SimpleRequest("new", nil)
	c.SimpleRequest("get_SID", nil)
	return c.Err()
}

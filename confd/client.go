package confd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultTimeout confd workers will kill the process after 60 seconds
const DefaultTimeout = time.Second * 60

// DefaultFacility system can only be used for local connections
const DefaultFacility = "system"

const anonymousUser = ""
const anonymousPassword = ""
const localhost = "127.0.0.1"

// DefaultPort of the confd listener
const DefaultPort = 4472

// LocalConnection is used on the box
var LocalConnection = fmt.Sprintf("http://%s:%s@%s:%d/%s", anonymousUser,
	anonymousPassword, localhost, DefaultPort, DefaultFacility)

// Options define confd connection options
type Options struct {
	// Name of the client default os.Argv[0] (used for logging, on the server)
	Name     string `json:"client,omitempty"`
	Facility string `json:"facility,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// SID can be a string (login) or number (anonymous)
	SID interface{} `json:",omitempty"`
}

// Conn is the confd connection object
type Conn struct {
	conn        *net.TCPConn
	URL         *url.URL
	id          int64 // json rpc counter
	err         error // error cache
	Logger      *log.Logger
	Timeout     time.Duration
	mutex       sync.Mutex
	write       sync.Mutex // prevent multiple write transactions
	read        sync.Mutex // prevent multiple read transactions
	Options     Options
	LastRequest time.Time // last time a request was done
}

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
func NewConn(URL string) (conn *Conn, err error) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}
	username := anonymousUser
	password := anonymousPassword

	if u.User != nil {
		if u.User.Username() != "" {
			username = u.User.Username()
		}

		if passwd, _ := u.User.Password(); passwd != "" {
			password = passwd
		}
	}

	facility := strings.Replace(u.Path, "/", "", -1)
	if facility == DefaultFacility {
		facility = ""
	}

	conn = &Conn{
		URL:     u,
		Timeout: DefaultTimeout,
		Logger:  nil,
		Options: Options{
			Name:     os.Args[0],
			Facility: facility,
			Username: username,
			Password: password,
			SID:      nil,
		},
	}
	return
}

// NewAnonymousConn creates a new confd connection (is not acually connecting)
// to http://127.0.0.1:4472/ (LocalConnection)
func NewAnonymousConn() (conn *Conn) {
	// error is only for url parsing which can not happen here, therefore ignored
	conn, _ = NewConn(LocalConnection)
	return conn
}

// SimpleRequest sends a simple request (untyped response) to the confd
func (c *Conn) SimpleRequest(method string, params ...interface{}) (interface{}, error) {
	resp := &SimpleResponse{}
	err := c.Request(method, resp, params...)
	return resp.Result, err
}

// Request allows to send request with typed (parsed with json) responses
func (c *Conn) Request(method string, result interface{}, params ...interface{}) error {
	// skip anything on error
	if c.err != nil {
		return c.err
	}
	setAndReturnErr := func(err error) error {
		c.Logger.Printf("Err: %v", err)
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
		c.logf("Connect to %s", c.safeURL())
		conn, err := net.Dial("tcp", c.URL.Host)
		if err != nil {
			return setAndReturnErr(err)
		}
		c.conn = conn.(*net.TCPConn)
		c.ResetErr()
		c.connect() // initalize the session
	}

	// send request
	req := fmt.Sprintf("POST / HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Content-Type: application/json\r\n"+
		"Content-Length: %d\r\n\r\n%s", c.URL.Host, buf.Len(), buf.String())

	var reader *bufio.Reader
	var resp *http.Response

	// send to remote side and recieve response
	c.mutex.Lock()

	err = c.conn.SetDeadline(time.Now().Add(c.Timeout))
	if err != nil {
		goto unlock
	}

	c.logf("Send request %s", buf.String())
	_, err = c.conn.Write([]byte(req[:]))
	if err != nil {
		goto unlock
	}

	// read response
	reader = bufio.NewReader(c.conn)
	resp, err = http.ReadResponse(reader, nil)
	c.LastRequest = time.Now()

unlock:
	c.mutex.Unlock()
	if err != nil {
		return setAndReturnErr(err)
	}

	// decode response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return setAndReturnErr(err)
	}
	respBuf := bytes.NewBuffer(respBytes)
	resp.Body.Close()
	dec := json.NewDecoder(respBuf)
	c.logf("Decode response %s", respBuf.String())
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
func (c *Conn) Close() (err error) {
	c.logf("Disconnect from %s", c.safeURL())
	if c.conn != nil {
		c.SimpleRequest("detach")
		if c.conn != nil {
			err = c.conn.Close()
		}
		c.conn = nil
	}
	if err == nil {
		err = c.Err()
	}
	c.ResetErr()
	return
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
	c.SimpleRequest("new", c.Options)
	if c.Options.SID == nil {
		// if we got a sid we will use it next time
		c.Options.SID, _ = c.SimpleRequest("get_SID")
	}
	return c.Err()
}

func (c *Conn) logf(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Printf(format, args...)
	}
}

func (c *Conn) safeURL() string {
	if c.Options.Password != "" {
		return strings.Replace(c.URL.String(), c.Options.Password, "********", 1)
	}
	return c.URL.String()
}

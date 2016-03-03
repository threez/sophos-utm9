package confd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
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
	Logger      *log.Logger
	Timeout     time.Duration
	rwlock      sync.RWMutex // prevent multiple write/read transactions
	Options     Options
	LastRequest time.Time // last time a request was done
	sync.Mutex
}

// Response is used for custom response handling
// just include the type in your types to handle errors
type Response struct {
	Error error `json:"error"` // pointer since it can be omitted
	ID    int64 `json:"id"`
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

// HTTP retruns an http request as bytes
func (r *Request) HTTP(host string) (*http.Request, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(r)
	if err != nil {
		return nil, err
	}
	ru, err := http.NewRequest("POST", "/", &buf)
	if err != nil {
		return nil, err
	}
	ru.URL.Host = host
	ru.Header.Set("Content-Type", "application/json")
	ru.Header.Set("User-Agent", "")
	return ru, nil
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
	// make sure we have a connection to the server
	err := c.connect() // initalize the session
	if err != nil {
		return err
	}

	reportErrorAndCloseConnection := func(err error) error {
		c.Logger.Printf("Error: %v", err)
		c.Close()
		return err
	}

	// request
	r := Request{method, params, c.id}
	c.id++
	c.logf("Send request %v", r)
	req, err := r.HTTP(c.URL.Host)
	if err != nil {
		return err
	}

	// send request
	resp, err := c.doRequest(req)
	if err != nil {
		return reportErrorAndCloseConnection(err)
	}
	c.LastRequest = time.Now()

	// decode response
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(result)
	if err != nil {
		return reportErrorAndCloseConnection(err)
	}

	// General error handling
	c.logf("Received response %v", result)
	switch r := result.(type) {
	case *Response:
		if r.Error != nil {
			return reportErrorAndCloseConnection(r.Error)
		}
	case *SimpleResponse:
		if r.Error != nil {
			return reportErrorAndCloseConnection(r.Error)
		}
	}

	return nil
}

func (c *Conn) doRequest(req *http.Request) (resp *http.Response, err error) {
	c.Lock()
	defer c.Unlock()

	// send to remote side and recieve response
	err = c.conn.SetDeadline(time.Now().Add(c.Timeout))
	if err != nil {
		return
	}

	err = req.Write(c.conn)
	if err != nil {
		return
	}

	// read response
	resp, err = http.ReadResponse(bufio.NewReader(c.conn), nil)

	return
}

// Close the confd connection
func (c *Conn) Close() (err error) {
	if c.conn != nil {
		c.logf("Disconnect from %s", c.safeURL())
		_, _ = c.SimpleRequest("detach") // ignore if we can't detach
		if c.conn != nil {
			c.Lock()
			err = c.conn.Close()
			c.Unlock()
		}
		c.conn = nil
	}
	return
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *Conn) connect() (err error) {
	if c.conn != nil {
		return
	}
	c.logf("Connect to %s", c.safeURL())
	c.Lock()
	conn, err := net.Dial("tcp", c.URL.Host)
	if err != nil {
		c.logf("Unable to connect %v", err)
		return
	}
	c.conn = conn.(*net.TCPConn)
	c.Unlock()
	_, err = c.SimpleRequest("new", c.Options)
	if err == nil && c.Options.SID == nil {
		// if we got a sid we will use it next time
		c.Options.SID, err = c.SimpleRequest("get_SID")
	}
	if err != nil {
		c.logf("Unable to create session %v", err)
	}
	return
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

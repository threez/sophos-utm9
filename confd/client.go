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
	URL         *url.URL      // URL that the connection connects to
	Logger      *log.Logger   // Logger if specified, will log confd actions
	Timeout     time.Duration // Timeout specifies the conn read/write timeout
	Options     Options       // Options represent connection options
	LastRequest time.Time     // LastRequest last time a request was done

	conn      *net.TCPConn
	id        int64        // json rpc counter
	rwlock    sync.RWMutex // prevent multiple write/read transactions
	connMutex sync.Mutex   // prevent concurrent confd access
}

// Response is used for custom response handling
// just include the type in your types to handle errors
type Response struct {
	Error  error            `json:"error"` // pointer since it can be omitted
	ID     int64            `json:"id"`
	Result *json.RawMessage `json:"result"`
}

func (r *Response) String() string {
	if r.Error != nil {
		return fmt.Sprintf("[%d] Error: %s", r.ID, r.Error)
	}
	return fmt.Sprintf("[%d] Result: %s", r.ID, *r.Result)
}

// Request is used to construct requests
type Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int64         `json:"id"`
}

func (r *Request) String() string {
	params, _ := json.Marshal(r.Params)
	return fmt.Sprintf("[%d] %s(%s)", r.ID, r.Method, string(params[:]))
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
	result := new(interface{})
	err := c.Request(method, result, params...)
	return result, err
}

// Request allows to send request with typed (parsed with json) responses
func (c *Conn) Request(method string, result interface{}, params ...interface{}) (err error) {
	// make sure we have a connection to the server
	err = c.connect()
	if err != nil {
		return
	}

	c.connMutex.Lock()
	err = c.request(method, result, params...)
	c.connMutex.Unlock()

	if err != nil {
		c.Logger.Printf("Error: %v", err)
	}
	return
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *Conn) connect() (err error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	if c.conn != nil {
		return
	}
	c.logf("Connect to %s", c.safeURL())
	conn, err := net.Dial("tcp", c.URL.Host)
	if err != nil {
		c.logf("Unable to connect %v", err)
		return
	}
	c.conn = conn.(*net.TCPConn)
	err = c.request("new", nil, &c.Options)
	if err == nil && c.Options.SID == nil {
		// if we got a sid we will use it next time
		resp := new(interface{})
		err = c.request("get_SID", resp)
		c.Options.SID = resp
	}
	if err != nil {
		c.logf("Unable to create session %v", err)
	}
	return
}

func (c *Conn) request(method string, result interface{}, params ...interface{}) error {
	// request
	r := Request{method, params, c.id}
	c.id++
	c.logf("=> %s", r.String())
	req, err := r.HTTP(c.URL.Host)
	if err != nil {
		return err
	}

	// send request
	resp, err := c.sendRecv(req)
	if err != nil {
		// send receive operation failed, conenction will be closed
		_ = c.close() // ignore close errors
		return err
	}
	c.LastRequest = time.Now()

	// decode response
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	respObj := new(Response)
	err = dec.Decode(respObj)
	if err != nil {
		return err
	}
	if result != nil {
		err = json.Unmarshal(*respObj.Result, result)
		if err != nil {
			return err
		}
	}
	c.logf("<= %v", respObj)

	// General error handling
	if respObj.Error != nil {
		return respObj.Error
	}

	return nil
}

// Close the confd connection
func (c *Conn) Close() (err error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()
	if c.conn != nil {
		c.logf("Disconnect from %s", c.safeURL())
		_ = c.request("detach", nil) // ignore if we can't detach
		_ = c.close()                // ignore close errors
	}
	return
}

func (c *Conn) close() (err error) {
	if c.conn != nil {
		err = c.conn.Close()
	}
	c.conn = nil
	return
}

func (c *Conn) sendRecv(req *http.Request) (resp *http.Response, err error) {
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

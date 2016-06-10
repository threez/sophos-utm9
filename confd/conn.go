// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var safePasswordRegexp = regexp.MustCompile(`password":"[^"]+"`)

const (
	msgConnect = iota
	msgRequest
	msgClose
	msgQuit
)

type roundTripHandler func(*http.Request) (*http.Response, error)

// sessionMsg is the object send to the worker goroutine
type sessionMsg struct {
	Type     int
	Request  *http.Request
	Response *http.Response
	Error    error
	Done     chan bool
}

// BUG(threez) It currently requires to connect directly to the confd database.
// This can be done by connecting through an ssh tunnel and forward the port
// 4472, e.g.:
//
//      ssh -L 4472:127.0.0.1:4472 root@utm

// Conn is the confd connection object
type Conn struct {
	Transport              Transport
	AutomaticErrorHandling bool
	URL                    *url.URL    // URL that the connection connects to
	Logger                 *log.Logger // Logger if specified, will log confd actions
	Options                *Options    // Options represent connection options
	id                     struct {
		Value      uint64 // json rpc counter
		sync.Mutex        // prevent double counting
	}
	txMu   sync.Mutex // prevent multiple write/read transactions
	queue  chan *sessionMsg
	worker struct {
		refs uint64 // counts the references to the worker
		sync.RWMutex
	}
}

// NewConn creates a new confd connection (is not acually connecting)
func NewConn(URL string) (conn *Conn, err error) {
	u, err := url.Parse(URL)
	if err != nil {
		return
	}

	conn = &Conn{
		URL:       u,
		Logger:    nil,
		Options:   newOptions(u),
		Transport: &tcpTransport{Timeout: defaultTimeout},
		queue:     make(chan *sessionMsg),
		AutomaticErrorHandling: true,
	}

	return
}

// NewAnonymousConn creates a new confd connection (is not acually connecting)
// to http://127.0.0.1:4472/ (Local Connection)
func NewAnonymousConn() (conn *Conn) {
	// error is only for url parsing which can not happen here, therefore ignored
	conn, _ = NewConn(anonymousLocalConn)
	return conn
}

// NewSystemConn creates a new confd connection (is not acually connecting)
// to http://system@127.0.0.1:4472/ (Local Connection)
func NewSystemConn() (conn *Conn) {
	// error is only for url parsing which can not happen here, therefore ignored
	conn, _ = NewConn(systemLocalConn)
	return conn
}

// NewUserConn creates a new conn for the given user (is not acually connecting)
// to http://user:password@127.0.0.1:4472/ (Local Connection)
func NewUserConn(username, password, ip string) (conn *Conn) {
	// error is only for url parsing which can not happen here, therefore ignored
	conn = NewAnonymousConn()
	conn.Options.Facility = "webadmin"
	conn.Options.Username = username
	conn.Options.Password = password
	conn.Options.IP = ip
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
	c.requireWorker()
	defer c.releaseWorker()
	err = c.request(c.queuedExecution, method, result, params...)

	// automatic error handling
	if c.AutomaticErrorHandling &&
		(err == ErrEmptyResponse || err == ErrReturnCode) {
		c.logf("!! Started automatic error handling because of: %s", err)
		errs, errl := c.ErrList()
		if errl != nil {
			return errl
		}
		if len(errs) > 0 {
			return errors.New(errs[0].Error())
		}
	}

	if err != nil {
		c.logf("Error: %v", err)
	}
	return
}

// Connect creates a new confd session by calling new and get_SID confd calls.
// It is preffered to not use the call and create sessions if requests are made
func (c *Conn) Connect() (err error) {
	c.requireWorker()
	defer c.releaseWorker()
	c.logf("Connect to %s", c.safeURL())
	msg := sessionMsg{Type: msgConnect, Done: make(chan bool)}
	c.queue <- &msg
	<-msg.Done // Wait until request was processed
	return msg.Error
}

// Close the confd connection
func (c *Conn) Close() (err error) {
	c.requireWorker()
	defer c.releaseWorker()
	c.logf("Disconnect from %s", c.safeURL())
	_ = c.request(c.queuedExecution, "detach", nil) // ignore if we can't detach
	msg := sessionMsg{Type: msgClose, Done: make(chan bool)}
	c.queue <- &msg
	<-msg.Done // Wait until request was processed
	return msg.Error
}

func (c *Conn) request(handler roundTripHandler, method string, result interface{}, params ...interface{}) error {
	// make sure we are connected
	err := c.connect()
	if err != nil {
		return err
	}

	// request
	r, err := newRequest(method, params, c.nextID())
	if err != nil {
		return err
	}
	c.logf("=> %s", r.String())
	req, err := r.HTTP(c.URL.Host)
	if err != nil {
		return err
	}

	// send request
	resp, err := handler(req)
	if err != nil {
		return err
	}

	// decode response
	respObj, err := newResponse(resp.Body)
	if respObj != nil {
		c.logf("<= %v", respObj)
	}
	if err != nil {
		return err
	}

	err = respObj.Decode(result, method != "get_SID")
	if err != nil {
		return err
	}

	return nil
}

// run is the worker function, that does real work, all transport stuff (confd)
// interactions should be serialized though this worker. the function will
// spin up and down as work is required to be done
func (c *Conn) run() {
	var end chan bool
	running := true
	for running {
		select {
		case msg := <-c.queue:
			switch msg.Type {
			case msgConnect:
				msg.Error = c.connect()
			case msgRequest:
				msg.Response, msg.Error = c.directExecution(msg.Request)
			case msgClose:
				msg.Error = c.close()
			case msgQuit:
				end = msg.Done
				running = false // stop the worker
				continue        // skip done inside the loop
			}
			if msg.Done != nil {
				msg.Done <- true
			}
		}
	}
	end <- true
}

// requireWorker increments the worker references and starts the worker if
// the refs is at 0
func (c *Conn) requireWorker() {
	c.worker.Lock()
	if c.worker.refs == 0 {
		go c.run()
	}
	c.worker.refs++
	c.worker.Unlock()
}

// releaseWorker decrements the worker references and quits the worker at 0 refs
func (c *Conn) releaseWorker() {
	c.worker.Lock()
	c.worker.refs--
	if c.worker.refs == 0 {
		msg := sessionMsg{Type: msgQuit, Done: make(chan bool)}
		c.queue <- &msg
		<-msg.Done // Wait until request was processed
	}
	c.worker.Unlock()
}

// Connect creates a new confd session by calling new and get_SID confd calls.
// It is preffered to not use the call and create sessions if requests are made
func (c *Conn) connect() (err error) {
	if c.Transport.IsConnected() {
		return
	}
	err = c.Transport.Connect(c.URL)
	if err != nil {
		c.logf("Unable to connect %s", err)
		return
	}
	err = c.request(c.directExecution, "new", nil, c.Options)
	if err == nil && c.Options.SID == nil {
		// if we got a sid we will use it next time
		err = c.request(c.directExecution, "get_SID", &c.Options.SID)
	}
	if err != nil {
		c.logf("Unable to create session %v", err)
	}
	return
}

// Close the confd connection
func (c *Conn) close() (err error) {
	_ = c.Transport.Close() // ignore close errors
	return
}

// logf takes care of logging if a logger is present and removes password
// information of a given form
func (c *Conn) logf(format string, args ...interface{}) {
	if c.Logger != nil {
		str := fmt.Sprintf(format, args...)
		str = safePasswordRegexp.ReplaceAllString(str, `password":"********"`)
		c.Logger.Print(str)
	}
}

// Returns a url that doesn't contain a password
func (c *Conn) safeURL() string {
	if c.Options.Password != "" {
		return strings.Replace(c.URL.String(), c.Options.Password, "********", 1)
	}
	return c.URL.String()
}

func (c *Conn) queuedExecution(req *http.Request) (*http.Response, error) {
	msg := sessionMsg{Type: msgRequest, Request: req, Done: make(chan bool)}
	c.queue <- &msg
	<-msg.Done // Wait until request was processed

	if msg.Error != nil {
		return nil, msg.Error
	}
	return msg.Response, nil
}

func (c *Conn) directExecution(req *http.Request) (*http.Response, error) {
	resp, err := c.Transport.RoundTrip(req)
	// send receive operation failed, connection will be closed
	if err != nil {
		_ = c.close() // ignore close errors
	}
	return resp, err
}

func (c *Conn) nextID() uint64 {
	c.id.Lock()
	defer c.id.Unlock()
	next := c.id.Value
	c.id.Value++
	return next
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
)

// BUG(threez) It currently requires to connect directly to the confd database.
// This can be done by connecting through an ssh tunnel and forward the port
// 4472, e.g.:
//
//      ssh -L 4472:127.0.0.1:4472 root@utm

// Conn is the confd connection object
type Conn struct {
	URL       *url.URL    // URL that the connection connects to
	Logger    *log.Logger // Logger if specified, will log confd actions
	Options   *Options    // Options represent connection options
	id        uint64      // json rpc counter
	Transport Transport
	txMu      sync.Mutex // prevent multiple write/read transactions
	sessionMu sync.Mutex // prevent concurrent confd access
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
	}
	return
}

// NewAnonymousConn creates a new confd connection (is not acually connecting)
// to http://127.0.0.1:4472/ (LocalConnection)
func NewAnonymousConn() (conn *Conn) {
	// error is only for url parsing which can not happen here, therefore ignored
	conn, _ = NewConn(anonymousLocalConn)
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

	err = c.request(method, result, params...)

	// automatic error handling
	if err == ErrEmptyResponse {
		errs, _ := c.ErrList()
		if len(errs) > 0 {
			return errors.New(errs[0].Error())
		}
	}

	if err != nil {
		c.Logger.Printf("Error: %v", err)
	}
	return
}

// Connect creates a new confd session by calling new and get_SID confd calls
func (c *Conn) connect() (err error) {
	if c.Transport.IsConnected() {
		return
	}
	c.sessionMu.Lock()
	defer c.sessionMu.Unlock()
	c.logf("Connect to %s", c.safeURL())
	err = c.Transport.Connect(c.URL)
	if err != nil {
		c.logf("Unable to connect %s", err)
		return
	}
	err = c.request("new", nil, c.Options)
	if err == nil && c.Options.SID == nil {
		// if we got a sid we will use it next time
		err = c.request("get_SID", &c.Options.SID)
	}
	if err != nil {
		c.logf("Unable to create session %v", err)
	}
	return
}

func (c *Conn) request(method string, result interface{}, params ...interface{}) error {
	// request
	id := atomic.AddUint64(&c.id, 1)
	r, err := newRequest(method, params, id)
	if err != nil {
		return err
	}
	c.logf("=> %s", r.String())
	req, err := r.HTTP(c.URL.Host)
	if err != nil {
		return err
	}

	// send request
	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		// send receive operation failed, conenction will be closed
		_ = c.Transport.Close() // ignore close errors
		return err
	}

	// decode response
	respObj, err := newResponse(resp.Body)
	if err != nil {
		return err
	}
	err = respObj.Decode(result)
	if err != nil {
		return err
	}

	c.logf("<= %v", respObj)

	return nil
}

// Close the confd connection
func (c *Conn) Close() (err error) {
	if c.Transport.IsConnected() {
		c.sessionMu.Lock()
		defer c.sessionMu.Unlock()
		c.logf("Disconnect from %s", c.safeURL())
		_ = c.request("detach", nil) // ignore if we can't detach
		_ = c.Transport.Close()      // ignore close errors
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

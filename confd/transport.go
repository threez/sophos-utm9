// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Transport interface is used by connections to transport data to and from
// the confd
type Transport interface {
	Connect(url *url.URL) error
	IsConnected() bool
	io.Closer
	http.RoundTripper
}

// TCPTransport implements a tcp+http RoundTripper for confd connections
type tcpTransport struct {
	Timeout     time.Duration // Timeout specifies the conn read/write timeout
	LastRequest time.Time     // LastRequest last time a request was done
	conn        *net.TCPConn
	mu          sync.RWMutex
}

// Connect to the passed url
func (t *tcpTransport) Connect(url *url.URL) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	conn, err := net.Dial("tcp", url.Host)
	if err != nil {
		return err
	}
	t.conn = conn.(*net.TCPConn)
	return nil
}

// RoundTrip executes a request/response round trip
func (t *tcpTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// send to remote side and recieve response
	err = t.conn.SetDeadline(time.Now().Add(t.Timeout))
	if err != nil {
		return
	}

	err = req.Write(t.conn)
	if err != nil {
		return
	}

	// read response
	resp, err = http.ReadResponse(bufio.NewReader(t.conn), nil)
	t.LastRequest = time.Now()

	return
}

// IsConnected returns
func (t *tcpTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conn != nil
}

// Close the transport
func (t *tcpTransport) Close() (err error) {
	if !t.IsConnected() {
		return // we already disconnected
	}
	t.mu.Lock()
	err = t.conn.Close()
	t.conn = nil
	t.mu.Unlock()
	return
}

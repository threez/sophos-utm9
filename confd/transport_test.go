// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransportTimeout(t *testing.T) {
	done := make(chan bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-done
	}))

	tcp := tcpTransport{Timeout: time.Millisecond * 100}

	u, err := url.Parse(server.URL)
	assert.NoError(t, err)
	err = tcp.Connect(u)
	assert.True(t, tcp.IsConnected())
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	started := time.Now()
	req, err := http.NewRequest("POST", "/", buf)
	assert.NoError(t, err)
	resp, err := tcp.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	took := time.Since(started)

	assert.False(t, tcp.IsConnected())

	assert.True(t, took < time.Millisecond*200, "must to timeout after 200ms")
	done <- true
	server.Close()
}

func TestTransportClose(t *testing.T) {
	done := make(chan bool)
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.CloseClientConnections()
		<-done
	}))

	tcp := tcpTransport{Timeout: time.Millisecond * 100}

	u, err := url.Parse(server.URL)
	assert.NoError(t, err)
	err = tcp.Connect(u)
	assert.True(t, tcp.IsConnected())
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	started := time.Now()
	req, err := http.NewRequest("POST", "/", buf)
	assert.NoError(t, err)
	resp, err := tcp.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	took := time.Since(started)

	assert.False(t, tcp.IsConnected())

	assert.True(t, took < time.Millisecond*100, "must to timeout before 100ms")
	done <- true
	server.Close()
}

func TestTransportError(t *testing.T) {
	server, err := net.Listen("tcp", ":0")
	assert.NoError(t, err)

	tcp := tcpTransport{Timeout: time.Millisecond * 100}

	u, err := url.Parse(fmt.Sprintf("http://%s/", server.Addr()))
	assert.NoError(t, err)
	err = tcp.Connect(u)
	assert.True(t, tcp.IsConnected())
	assert.NoError(t, err)

	buf := &bytes.Buffer{}
	started := time.Now()

	go func() {
		_, e := server.Accept()
		assert.NoError(t, e)
	}()

	// First request
	req, err := http.NewRequest("POST", "/", buf)
	assert.NoError(t, err)
	resp, err := tcp.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	took := time.Since(started)
	assert.True(t, took < time.Millisecond*110, "must to timeout before 100ms")
	assert.False(t, tcp.IsConnected())

	// Subsequest request
	req, err = http.NewRequest("POST", "/", buf)
	assert.NoError(t, err)
	resp, err = tcp.RoundTrip(req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	took = time.Since(started)
	assert.True(t, took < time.Millisecond*110, "must to timeout before 100ms")

	assert.NoError(t, server.Close())
}

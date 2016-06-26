// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var tc, ts, tr = NewTestRestD()

// GET /network/host/ -> 200
func TestServerObjectGetAllAPI(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("getAllObject").URL("class", "network", "type", "host")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("GET", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

// GET /network/host/REF_UNKNOWN -> 404
func TestServerObjectGetAPIUnknown(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("getObject").URL("class", "network", "type", "host",
		"ref", "UNKNOWN")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("GET", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
}

// GET /network/host/REF_DefaultSuperAdminGroup -> 404
func TestServerObjectGetAPIWrongType(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("getObject").URL("class", "network", "type", "host",
		"ref", "DefaultSuperAdminNetwork")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("GET", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
}

// GET /network/any/REF_NetworkAny4 -> 200
func TestServerObjectGetAPICorrect(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("getObject").URL("class", "network", "type", "any",
		"ref", "NetworkAny4")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("GET", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

// UNLOCK /network/any/REF_NetworkAny4 -> 204
func TestServerObjectUnlockAPICorrect(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("unlockObject").URL("class", "network", "type", "any",
		"ref", "NetworkAny4")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("UNLOCK", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode)
}

// LOCK /network/any/REF_NetworkAny4 -> 204
func TestServerObjectLockAPICorrect(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("lockObject").URL("class", "network", "type", "any",
		"ref", "NetworkAny4")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("LOCK", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode)
}

// GET /swagger ...
func TestServerSwaggerAPI(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("swaggerDefinitions").URL()
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("GET", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func NewTestRestD() (*http.Client, *httptest.Server, *mux.Router) {
	server := &Server{
		apiLogger: negroni.NewLogger(),
		apiPrefix: "/api",
	}
	confdLogger = log.New(os.Stdout, "confd ", log.LstdFlags)
	r, err := server.Router()
	if err != nil {
		panic(err)
	}
	httpServer := httptest.NewServer(r)
	httpClient := &http.Client{Transport: &http.Transport{}}

	return httpClient, httpServer, server.router
}

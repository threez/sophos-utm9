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

func TestServerObjectGetAllAPI(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("getAllObject").URL("class", "aaa", "type", "group")
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("Get", url, &buf)
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "pppp")
	res, err := tc.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

// GET /REF_DefaultSuperAdminGroup
func TestServerSwaggerAPI(t *testing.T) {
	buf := bytes.Buffer{}
	path, err := tr.Get("swaggerDefinitions").URL()
	assert.NoError(t, err)
	url := ts.URL + path.String()
	req, err := http.NewRequest("Get", url, &buf)
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

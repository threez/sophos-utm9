// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

// Server implements the restd server component
type Server struct {
	apiLogger *negroni.Logger
	apiPrefix string
	addr      string
	router    *mux.Router
}

// ListenAndServe starts the http server
func (server *Server) ListenAndServe() {
	log.Printf("Starting restd...")
	handler, err := server.Router()
	if err != nil {
		log.Fatalf("Error fetch meta data: %s", err)
	}

	log.Printf("Listening on http://%s%s", server.addr, server.apiPrefix)
	err = http.ListenAndServe(server.addr, handler)
	if err != nil {
		log.Fatalf("Error can't start server: %s", err)
	}
}

// Router returns the restd router
func (server *Server) Router() (http.Handler, error) {
	server.router = mux.NewRouter()
	server.router.StrictSlash(true)

	api, err := NewSwaggerAPI(server.apiPrefix)
	if err != nil {
		return nil, err
	}

	oapi := &ObjectAPI{api}

	// all the api is located at `apiPrefix`
	s := server.router.PathPrefix(server.apiPrefix).Subrouter()

	// Register objects api
	oapi.RegisterAPI(s)

	// Register swagger ui and, /definitions and json endpoints
	api.RegisterSwaggerAPI(s)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(server.apiLogger)
	n.UseHandler(api.Cors(server.router))

	return n, nil
}

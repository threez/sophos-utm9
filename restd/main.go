// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var (
	apiPrefix     = "/api"
	listen        = ":3000"
	syslogEnabled = false
)

func main() {
	parseArguments()
	initLogger()
	log.Printf("Starting restd...")

	r := mux.NewRouter()
	r.StrictSlash(true)

	api, err := NewSwaggerAPI(apiPrefix)
	oapi := &ObjectAPI{api}

	if err != nil {
		log.Fatalf("Error fetch meta data: %s", err)
	}

	s := r.PathPrefix(apiPrefix).Subrouter()

	// GET single object
	s.HandleFunc("/objects/{class}/{type}/REF_{ref}",
		withConfdConnection(oapi.Get))

	// GET all objects
	s.HandleFunc("/objects/{class}/{type}",
		withConfdConnection(oapi.All))

	// Register swagger ui and, /classes and json endpoints
	api.RegisterSwaggerAPI(s)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(apiLogger)
	n.UseHandler(api.Cors(r))

	log.Printf("Listening on http://%s%s", listen, apiPrefix)
	err = http.ListenAndServe(listen, n)

	if err != nil {
		log.Fatalf("Error can't start server: %s", err)
	}
}

func parseArguments() {
	flag.BoolVar(&syslogEnabled, "syslog", syslogEnabled,
		"if enabled logs everything to the daemon log (level notice)")
	flag.StringVar(&apiPrefix, "api-prefix", apiPrefix,
		"defines the api location prefix")
	flag.StringVar(&listen, "listen", listen,
		"defines, where the server is started <interface:port>")
	flag.Parse()
}

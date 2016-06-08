// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
)

var (
	server        = &Server{}
	syslogEnabled = false
)

func main() {
	parseArguments()
	initLogger()
	server.ListenAndServe()
}

func parseArguments() {
	flag.BoolVar(&syslogEnabled, "syslog", syslogEnabled,
		"if enabled logs everything to the daemon log (level notice)")
	flag.StringVar(&server.apiPrefix, "api-prefix", "/api",
		"defines the api location prefix")
	flag.StringVar(&server.addr, "listen", ":3000",
		"defines, where the server is started <interface:port>")
	flag.Parse()
}

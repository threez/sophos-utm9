// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/threez/sophos-utm9/confd"
	"net/http"
	"strings"
)

// ipFromRemoteAddr retuns a ipv4/6 address from the client of the request
func ipFromRemoteAddr(r *http.Request) string {
	ip := r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")]
	ip = strings.Replace(ip, "[", "", 1) // remove [] from ipv6 addresses
	ip = strings.Replace(ip, "]", "", 1) // so that confd understands it
	return ip
}

// basicAuthConfdConn creates a confd connection based of the passed
// basic auth information. It errors, if the authentication was not passed or
// is invalid
func basicAuthConfdConn(r *http.Request) (*confd.Conn, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, fmt.Errorf("No authentication header found")
	}

	conn := confd.NewUserConn(username, password, ipFromRemoteAddr(r))

	err := conn.Connect()
	if err != nil || fmt.Sprint(conn.Options.SID) == "0" {
		_ = conn.Close()
		return nil, fmt.Errorf("Unable to authenticate")
	}

	return conn, nil
}

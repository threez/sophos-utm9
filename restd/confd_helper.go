// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"

	"github.com/threez/sophos-utm9/confd"
)

var (
	// must be initialized before use (normally done in `initLogger`)
	confdLogger *log.Logger
)

type confdHandler func(w http.ResponseWriter, r *http.Request, conn *confd.Conn)

func withConfdConnection(handler confdHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := basicAuthConfdConn(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		conn.Logger = confdLogger
		handler(w, r, conn)
		// ignore failures, since the connection is not used any more, and
		// if the connection would still be alive, confd would kill it anyway
		_ = conn.Close()
	}
}

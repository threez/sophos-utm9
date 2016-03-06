// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultTimeout confd workers will kill the process after 60 seconds
const defaultTimeout = time.Second * 60

// DefaultFacility system can only be used for local connections
const defaultFacility = "system"

const anonymousUser = ""
const anonymousPassword = ""
const localhost = "127.0.0.1"

// DefaultPort of the confd listener
const defaultPort = 4472

// LocalConnection is used on the box
var anonymousLocalConn = fmt.Sprintf("http://%s:%s@%s:%d/%s", anonymousUser,
	anonymousPassword, localhost, defaultPort, defaultFacility)

// Options define confd connection options
type Options struct {
	// Name of the client default os.Argv[0] (used for logging, on the server)
	Name     string `json:"client,omitempty"`
	Facility string `json:"facility,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// SID (SessionID) can be a string (login) or number (anonymous)
	SID interface{} `json:",omitempty"`
}

// newOptions are derived by defualts and the passed url
func newOptions(url *url.URL) *Options {
	username := anonymousUser
	password := anonymousPassword

	if url.User != nil {
		if url.User.Username() != "" {
			username = url.User.Username()
		}

		if passwd, _ := url.User.Password(); passwd != "" {
			password = passwd
		}
	}

	facility := strings.Replace(url.Path, "/", "", -1)
	if facility == defaultFacility {
		facility = ""
	}

	return &Options{
		Name:     os.Args[0],
		Facility: facility,
		Username: username,
		Password: password,
	}
}

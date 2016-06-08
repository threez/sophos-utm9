// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// request is used to construct requests
type request struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     uint64           `json:"id"`
}

// NewRequest creates a new request object
func newRequest(method string, params interface{}, id uint64) (req *request, err error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	rawMsg := json.RawMessage(data)
	r := &request{method, &rawMsg, id}
	return r, err
}

func (r *request) String() string {
	params := string(([]byte(*r.Params))[:])
	if params == "null" {
		params = ""
	}
	return fmt.Sprintf("[%d] %s(%s)", r.ID, r.Method, params)
}

// HTTP retruns an http request as bytes
func (r *request) HTTP(host string) (*http.Request, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(r)
	if err != nil {
		return nil, err
	}
	ru, err := http.NewRequest("POST", "/", &buf)
	if err != nil {
		return nil, err
	}
	ru.URL.Host = host
	ru.Header.Set("Content-Type", "application/json")
	ru.Header.Set("User-Agent", "")
	return ru, nil
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// ErrEmptyResponse is likly triggered by calling a function that isn't exported
var ErrEmptyResponse = errors.New("Empty response")

// ErrReturnCode is triggered by a 0 return value
var ErrReturnCode = errors.New("Returned 0, check errors")

// response is used for custom response handling
// just include the type in your types to handle errors
type response struct {
	Error  *string          `json:"error"` // pointer since it can be omitted
	ID     int64            `json:"id"`
	Result *json.RawMessage `json:"result"`
}

// newResponse based of the passed reader
func newResponse(reader io.ReadCloser) (resp *response, err error) {
	defer func() { _ = reader.Close() }() // ignore fail here
	dec := json.NewDecoder(reader)
	resp = new(response)
	err = dec.Decode(resp)
	if err != nil {
		return
	}
	return
}

// Decode the response into passed result or return request error
func (r *response) Decode(result interface{}, checkReturn bool) (err error) {
	if r.Error != nil {
		return errors.New(*r.Error)
	}
	if r.Result == nil {
		return ErrEmptyResponse
	}
	// disable return code cheking for confd.Bool values (they are encoded as
	// integer values)
	if checkReturn && result != nil &&
		reflect.TypeOf(result).String() == "*confd.Bool" {
		checkReturn = false
	}
	if checkReturn && string((*r.Result)[:]) == "0" {
		return ErrReturnCode
	}
	if result != nil {
		err = json.Unmarshal(*r.Result, result)
		if err != nil {
			return
		}
	}
	return
}

func (r *response) String() string {
	if r.Error != nil {
		return fmt.Sprintf("[%d] Error: %s", r.ID, *r.Error)
	}
	return fmt.Sprintf("[%d] Result: %s", r.ID, *r.Result)
}

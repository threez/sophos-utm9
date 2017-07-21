// Copyright 2017 Georg Fleig. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseStringEmpty(t *testing.T) {
	emptyResponse := &response{}
	var str string
	assert.NotPanics(t, func() {
		str = emptyResponse.String()
	})
	assert.Equal(t, "[0] Result: empty response", str)
}

func TestResponseStringError(t *testing.T) {
	err := "broken"
	errorResponse := &response{
		Error: &err,
	}
	var str string
	assert.NotPanics(t, func() {
		str = errorResponse.String()
	})
	assert.Equal(t, "[0] Error: broken", str)
}

func TestResponseStringOK(t *testing.T) {
	errorResponse := &response{
		Result: &json.RawMessage{},
	}
	var str string
	assert.NotPanics(t, func() {
		str = errorResponse.String()
	})
	assert.Equal(t, "[0] Result: ", str)
}

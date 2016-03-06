// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializeOptionsAnonymous(t *testing.T) {
	conn := NewAnonymousConn()
	conn.Options.Name = "test"
	bytes, err := json.Marshal(conn.Options)
	assert.NoError(t, err)
	assert.Equal(t, `{"client":"test"}`, string(bytes[:]))
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRights(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	rights, err := conn.GetRights()
	assert.NoError(t, err)
	assert.Equal(t, []string{"ANONYMOUS"}, rights)
}

func TestAdminRights(t *testing.T) {
	conn := NewSystemConn()
	defer conn.Close()

	rights, err := conn.GetRights()
	assert.NoError(t, err)
	assert.Equal(t, []string{"SUPERADMIN"}, rights)
}

func TestHasRight(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	ok, err := conn.HasRight("foo")
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = conn.HasRight("ANONYMOUS")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestHasOneOfRights(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	ok, err := conn.HasOneOfRights([]string{"foo"})
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = conn.HasOneOfRights([]string{"ANONYMOUS"})
	assert.NoError(t, err)
	assert.True(t, ok)
}

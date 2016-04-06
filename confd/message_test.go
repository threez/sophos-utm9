// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErr(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	tx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)
	defer tx.Rollback()

	err = conn.request("del_object", nil, "REF_DefaultInternal")
	assert.Equal(t, ErrReturnCode, err)

	num, err := conn.ErrIsFatal()
	assert.NoError(t, err)
	assert.True(t, num > 0)

	num, err = conn.ErrIsNoack()
	assert.NoError(t, err)
	assert.True(t, num > 0)

	errs, err := conn.ErrList()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)
	assert.Equal(t, "OBJECT_DELETE_PARENT_DEL", errs[0].MessageType)
	assert.Contains(t, errs[0].Error(),
		"Continuing will delete the latter object as well.")

	errs, err = conn.ErrListFatal()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)

	errs, err = conn.ErrListNoAck()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)
}

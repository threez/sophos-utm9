// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErr(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	tx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)
	defer tx.Rollback()

	ok, err := conn.DelObject("REF_DefaultInternal")
	assert.NoError(t, err)
	assert.False(t, ok)

	num, err := conn.ErrIsFatal()
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), num)

	num, err = conn.ErrIsNoack()
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), num)

	errs, err := conn.ErrList()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "OBJECT_DELETE_DENY", errs[0].MessageType)
	assert.Equal(t, "FATAL [OBJECT_DELETE_DENY] Permission denied to delete "+
		"ethernet standard interface object 'Internal'.", errs[0].Error())

	errs, err = conn.ErrListFatal()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "OBJECT_DELETE_DENY", errs[0].MessageType)

	errs, err = conn.ErrListNoAck()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(errs))
	assert.Equal(t, "OBJECT_DELETE_DENY", errs[0].MessageType)
}

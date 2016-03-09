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
	conn.Options.Username = "system"
	defer conn.Close()

	tx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)
	defer tx.Rollback()

	err = conn.request("del_object", nil, "REF_DefaultInternal")
	assert.Equal(t, ErrReturnCode, err)

	num, err := conn.ErrIsFatal()
	assert.Error(t, err)
	assert.Equal(t, uint64(0), num)

	num, err = conn.ErrIsNoack()
	assert.NoError(t, err)
	assert.Equal(t, uint64(7), num)

	errs, err := conn.ErrList()
	assert.NoError(t, err)
	assert.Equal(t, 7, len(errs))
	assert.Equal(t, "OBJECT_DELETE_PARENT_DEL", errs[0].MessageType)
	assert.Equal(t, "[OBJECT_DELETE_PARENT_DEL] The ethernet standard "+
		"interface object 'Internal' is required by the QoS interface object "+
		"'Internal'.\nContinuing will delete the latter object as well.",
		errs[0].Error())

	errs, err = conn.ErrListFatal()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(errs))

	errs, err = conn.ErrListNoAck()
	assert.NoError(t, err)
	assert.Equal(t, 7, len(errs))
}

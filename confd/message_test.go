// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErr(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	tx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	conn.requireWorker()
	// Try to delete an object that is protected/used. Deletion should throw a
	// non-acknowledgeable fatal error.
	err = conn.request(conn.queuedExecution, "del_object",
		nil, "REF_DefaultInternalNetwork")
	assert.Equal(t, ErrReturnCode, err)
	conn.releaseWorker()

	num, err := conn.ErrIsFatal()
	assert.NoError(t, err)
	assert.True(t, num > 0)

	num, err = conn.ErrIsNoack()
	assert.NoError(t, err)
	assert.True(t, num > 0)

	errs, err := conn.ErrList()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)
	assert.Equal(t, "OBJECT_DELETE_LOCKED", errs[0].MessageType)
	assert.Contains(t, errs[0].Error(),
		"The interface network object 'Internal (Network)' is protected from "+
			"deletion.")

	errs, err = conn.ErrListFatal()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)

	errs, err = conn.ErrListNoAck()
	assert.NoError(t, err)
	assert.True(t, len(errs) > 0)
}

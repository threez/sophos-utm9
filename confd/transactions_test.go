// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadTransactions(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	rtx, err := conn.BeginReadTransaction()
	assert.NoError(t, err)

	obj, err := conn.GetAnyObject("REF_AnonymousUser")
	assert.NoError(t, err)
	assert.Equal(t, "aaa", obj.Class)
	assert.Equal(t, "Anonymous user", obj.Data["comment"])

	rtx.Commit()
}

func TestReadRollbackTransactions(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	rtx, err := conn.BeginReadTransaction()
	assert.NoError(t, err)
	err = rtx.Rollback()
	assert.NoError(t, err)
}

func TestWriteTransactions(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	wtx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)

	err = wtx.Commit()
	assert.NoError(t, err)
}

func TestWriteRollbackTransactions(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	wtx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)

	err = wtx.Rollback()
	assert.NoError(t, err)
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAnyObject(t *testing.T) {
	conn := connHelper()
	defer func() { _ = conn.Close() }()

	obj, err := conn.GetAnyObject("REF_AnonymousUser")
	assert.NoError(t, err)
	assert.Equal(t, "aaa", obj.Class)
	assert.Equal(t, "Anonymous user", obj.Data["comment"])
}

func TestAffectedObjects(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	refs, err := conn.GetAffectedObjects([]string{"REF_DefaultInternalNetwork"})
	assert.NoError(t, err)
	assert.Contains(t, refs, "REF_DefaultInternal")
	assert.Contains(t, refs, "REF_ItfParamsDefaultInternal")
	assert.Contains(t, refs, "REF_DefaultInternalNetwork")
}

func TestFilterObjects(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	objects, err := conn.FilterObjects().
		ClassName("aaa").
		TypeName("user").
		Eq("status", 1).
		Gt("enabled", 0).
		Gte("enabled", 1).
		Lt("enabled", 2).
		Lte("enabled", 1).
		Default("status").
		Ne("hidden", 0).
		Or(conn.FilterObjects().Eq("name", "system").Eq("name", "Anonymous")).
		Matches("comment", "super").
		NotMatches("loc", "german").
		And(conn.FilterObjects().Not(conn.FilterObjects().Eq("name", "foo"))).
		Get()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(objects))
}

func TestAllObjects(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	objects, err := conn.GetAllObjects()
	assert.NoError(t, err)
	assert.True(t, len(objects) > 300)
}

func TestSetObject(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()
	tx, err := conn.BeginWriteTransaction()
	assert.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	var host = AnyObject{
		ObjectMeta: ObjectMeta{
			Class: "network",
			Type:  "host",
		},
		Data: map[string]interface{}{
			"address":  "8.8.8.8",
			"name":     "Google DNS",
			"resolved": BoolValue(true),
		},
	}

	ref, err := conn.SetObject(&host, true)
	assert.NoError(t, err)
	assert.Contains(t, ref, "REF_")

	obj, err := conn.GetAnyObject(ref)
	assert.NoError(t, err)
	assert.Equal(t, obj.Data["address"], "8.8.8.8")

	err = conn.MoveObject(ref, "REF_GOOGLEDNS")
	assert.NoError(t, err)

	obj, err = conn.GetAnyObject("REF_GOOGLEDNS")
	assert.NoError(t, err)
	assert.Equal(t, obj.Data["address"], "8.8.8.8")

	err = conn.LockObject("REF_GOOGLEDNS")
	assert.NoError(t, err)

	err = conn.UnlockObject("REF_GOOGLEDNS")
	assert.NoError(t, err)
}

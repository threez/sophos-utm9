// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetaObjects(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	ret, err := conn.GetMetaObjects()
	assert.NoError(t, err)

	assert.Equal(t, ret["dhcp"]["server"]["mappings"].ISA, "ARRAY")
	assert.Equal(t, ret["dhcp"]["server"]["mappings"].Type, "REF")
	assert.Equal(t, ret["dhcp"]["server"]["mappings"].Class, "network")
	assert.Equal(t, ret["dhcp"]["server"]["mappings"].Types[0], "host")
}

func TestClassesAndTypes(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	classes, err := conn.GetObjectClasses()
	assert.NoError(t, err)
	assert.Contains(t, classes, "aaa")

	types, err := conn.GetObjectTypes("aaa")
	assert.NoError(t, err)
	assert.Contains(t, types, "user")
}

func TestAvailableNodes(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	paths, err := conn.GetAvailableNodes()
	assert.NoError(t, err)
	assert.Contains(t, paths, "remote_access")

	paths, err = conn.GetAvailableNodes("remote_access")
	assert.NoError(t, err)
	assert.Contains(t, paths, "pptp")
}

func TestScalarsAndArrays(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	paths, err := conn.GetScalars("settings")
	assert.NoError(t, err)
	assert.Contains(t, paths, "country")

	paths, err = conn.GetArrays("remote_access", "pptp")
	assert.NoError(t, err)
	assert.Contains(t, paths, "aaa")
}

func TestMeta(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	data, err := conn.GetMeta()
	assert.NoError(t, err)
	assert.Equal(t, "^..$", data.Tree("settings").Tree("country")["_regex"])
}

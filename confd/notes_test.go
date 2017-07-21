// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()

	var validInterfaces = []string{"REF_NetNet100008", "REF_NetworkAny"}

	node, err := conn.GetNode("ssh")
	assert.NoError(t, err)
	assert.Equal(t, float64(22), node["port"])
	assert.Contains(t, validInterfaces,
		node["allowed_networks"].([]interface{})[0])

	nodev, err := conn.GetNodeValue("ssh", "allowed_networks")
	assert.NoError(t, err)
	assert.Contains(t, validInterfaces, nodev.([]interface{})[0])

	paths, err := conn.GetAffectedNodes("REF_NetworkAny")
	assert.NoError(t, err)
	assert.Contains(t, paths, NodePath{"epp", "allowed_networks"})

	afc, err := conn.GetNode("afc")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), afc["status"])
	afc["status"] = BoolValue(true)

	ok, err := conn.SetNode(afc, "afc")
	assert.NoError(t, err)
	assert.True(t, ok)

	afc, err = conn.GetNode("afc")
	assert.NoError(t, err)
	assert.Equal(t, float64(1), afc["status"])

	ok, err = conn.ResetNode("afc")
	assert.NoError(t, err)
	assert.True(t, ok)

	list, err := conn.GetNodes()
	assert.NoError(t, err)
	assert.Contains(t, list, NodeName("ssh"))
	assert.Contains(t, list, NodeName("http"))
}

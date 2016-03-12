// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNode(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	node, err := conn.GetNode("ssh")
	assert.NoError(t, err)
	assert.Equal(t, float64(22), node["port"])
	assert.Equal(t, "REF_DefaultInternalNetwork",
		node["allowed_networks"].([]interface{})[0])

	nodev, err := conn.GetNodeValue("ssh", "allowed_networks")
	assert.NoError(t, err)
	assert.Equal(t, "REF_DefaultInternalNetwork", nodev.([]interface{})[0])

	paths, err := conn.GetAffectedNodes("REF_DefaultInternalNetwork")
	assert.NoError(t, err)
	assert.Contains(t, paths, NodePath{"ssh", "allowed_networks"})

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

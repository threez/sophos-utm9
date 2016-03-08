// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAnyObject(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	obj, err := conn.GetAnyObject("REF_AnonymousUser")
	assert.NoError(t, err)
	assert.Equal(t, "aaa", obj.Class)
	assert.Equal(t, "Anonymous user", obj.Data["comment"])
}

func TestAffectedObjects(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	refs, err := conn.GetAffectedObjects([]string{"REF_DefaultInternalNetwork"})
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"REF_PacMasFromInterNetwo",
		"REF_DefaultHTTPProfile",
		"REF_DefaultInternalNetwork",
		"REF_DefaultInternal",
		"REF_QosItfDefaultInternal",
		"REF_ItfParamsDefaultInternal",
	}, refs)
}

func TestFilterObjects(t *testing.T) {
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

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
	conn := connHelper()
	conn.Options.Username = "system"
	defer conn.Close()

	objects, err := conn.GetAllObjects()
	assert.NoError(t, err)
	assert.True(t, len(objects) > 300)
}

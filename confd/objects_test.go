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

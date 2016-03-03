package confd

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestExports(t *testing.T) {
	conn := NewAnonymousConn()
	conn.Logger = log.New(os.Stdout, "confd ", log.LstdFlags)
	conn.Options.Name = "confd-package-test"
	conn.Timeout = time.Second * 1
	defer conn.Close()
	assert.Equal(t, nil, conn.Options.SID)
	exports, err := conn.Exports()
	assert.NoError(t, err)

	assert.True(t, len(exports) > 300, "Expected more than 300 methods to "+
		"be exported but found %d", len(exports))
	assert.Equal(t, "Session", exports["get_SID"].Module,
		"Expected method get_SID to be exported in module Session but got: '%s'",
		exports["get_SID"].Module)

	err = conn.Close()
	assert.NoError(t, err)

	obj, err := conn.GetAnyObject("REF_AnonymousUser")
	assert.NoError(t, err)
	assert.Equal(t, "aaa", obj.Class)
	assert.Equal(t, "Anonymous user", obj.Data["comment"])
	sid, err := conn.SimpleRequest("get_SID")
	assert.NoError(t, err)
	assert.Equal(t, sid, conn.Options.SID)
}

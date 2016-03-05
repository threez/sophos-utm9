package confd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExports(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	exports, err := conn.Exports()
	assert.NoError(t, err)
	assert.True(t, len(exports) > 300, "Expected more than 300 methods to "+
		"be exported but found %d", len(exports))
	assert.Equal(t, "Session", exports["get_SID"].Module,
		"Expected method get_SID to be exported in module Session but got: '%s'",
		exports["get_SID"].Module)
}

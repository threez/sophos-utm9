package confd

import (
	"testing"
)

func TestExports(t *testing.T) {
	conn := NewConn()
	err := conn.Connect()
	if err != nil {
		t.Error(err)
	}
	exports, err := conn.Exports()
	if err != nil {
		t.Error(err)
	}

	if len(exports) > 393 {
		t.Errorf("Expected more than 393 methods to be exported 393 but found %d",
			len(exports))
	}

	if exports["get_SID"].Module != "Session" {
		t.Errorf("Expected method get_SID to be exported in module "+
			"Session but got: '%s'", exports["get_SID"].Module)
	}
}

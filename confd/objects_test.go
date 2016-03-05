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

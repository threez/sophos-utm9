package confd

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func connHelper() *Conn {
	conn := NewAnonymousConn()
	conn.Logger = log.New(os.Stdout, "confd ", log.LstdFlags)
	conn.Options.Name = "confd-package-test"
	conn.Timeout = time.Second * 1
	return conn
}

func TestSID(t *testing.T) {
	conn := connHelper()
	defer conn.Close()

	sid, err := conn.SimpleRequest("get_SID")
	assert.NoError(t, err)
	assert.Equal(t, sid, conn.Options.SID)
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentAccess(t *testing.T) {
	// Starts three go routines, that concurrently access the connection
	conn := connHelper()
	defer func() { _ = conn.Close() }()
	wg := sync.WaitGroup{}

	wg.Add(3)
	work := func(conn *Conn) {
		for i := 0; i < 10; i++ {
			_ = conn.Connect()
			for i := 0; i < 5; i++ {
				_, _ = conn.SimpleRequest("get_SID")
			}
			_ = conn.Close()
		}
		wg.Done()
	}

	for i := 0; i < 3; i++ {
		go work(conn)
	}

	wg.Wait()
}

func TestConcurrentTransactionAccess(t *testing.T) {
	// Starts three go routines, that concurrently access the connection
	conn := systemConnHelper()
	defer func() { _ = conn.Close() }()
	wg := sync.WaitGroup{}
	sid, err := conn.SimpleRequest("get_SID")
	assert.NoError(t, err)

	wg.Add(20)
	work := func(conn *Conn) {
		for i := 0; i < 3; i++ {
			tx, err := conn.BeginWriteTransaction()
			assert.NoError(t, err)
			for i := 0; i < 1; i++ {
				value, err := conn.SimpleRequest("get_SID")
				assert.NoError(t, err)
				assert.Equal(t, sid, value)
			}
			assert.NoError(t, tx.Commit())
		}
		wg.Done()
	}

	for i := 0; i < 20; i++ {
		go work(conn)
	}

	wg.Wait()
}

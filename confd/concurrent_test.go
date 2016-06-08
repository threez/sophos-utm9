// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"sync"
	"testing"
)

func TestConcurrentAccess(t *testing.T) {
	// Starts three go routines, that concurrently access the connection
	conn := connHelper()
	defer func() { _ = conn.Close() }()
	wg := sync.WaitGroup{}

	wg.Add(3)
	work := func(conn *Conn) {
		for i := 0; i < 10; i++ {
			conn.Connect()
			for i := 0; i < 5; i++ {
				conn.SimpleRequest("get_SID")
			}
			conn.Close()
		}
		wg.Done()
	}

	go work(conn)
	go work(conn)
	go work(conn)

	wg.Wait()
}

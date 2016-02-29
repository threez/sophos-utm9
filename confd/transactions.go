package confd

// Transaction abstracts the read and write transactions to the confd
type Transaction interface {
	// Commit current transaction
	Commit() error
	// Rollback current Transaction
	Rollback() error
}

type writeTransaction struct{ *Conn }
type readTransaction struct{ *Conn }

// BeginReadTransaction starts new read transaction
func (c *Conn) BeginReadTransaction() (Transaction, error) {
	c.read.Lock()
	_, err := c.SimpleRequest("freeze", nil)
	if err != nil {
		return nil, err
	}
	return &readTransaction{c}, nil
}

// BeginWriteTransaction starts new write transaction
func (c *Conn) BeginWriteTransaction() (Transaction, error) {
	c.write.Lock()
	_, err := c.SimpleRequest("lock", nil)
	if err != nil {
		return nil, err
	}
	return &writeTransaction{c}, nil
}

func (t *readTransaction) Rollback() (err error) {
	_, err = t.SimpleRequest("thaw", nil)
	t.read.Unlock()
	return
}

func (t *readTransaction) Commit() (err error) {
	_, err = t.SimpleRequest("thaw", nil)
	t.read.Unlock()
	return
}

func (t *writeTransaction) Rollback() (err error) {
	_, err = t.SimpleRequest("unlock", nil)
	t.write.Unlock()
	return
}

func (t *writeTransaction) Commit() (err error) {
	_, err = t.SimpleRequest("commit", nil)
	t.write.Unlock()
	return
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"fmt"
)

// ErrDescription is returned by ErrList* functions and details the occured
// error
type ErrDescription struct {
	Name             string   `json:"name"`
	Rights           string   `json:"rights"`
	Attributes       []string `json:"attrs"`
	ObjectAttributes []string `json:"Oattrs"`
	ObjectName       string   `json:"objname"`
	Function         string   `json:"del_object"`
	Ref              string   `json:"ref"`
	MessageType      string   `json:"msgtype"`
	NeverHide        Bool     `json:"never_hide"`
	Format           string   `json:"format"`
	Fatal            Bool     `json:"fatal"`
	Class            string   `json:"class"`
	Type             string   `json:"type"`
	Permission       string   `json:"perms"`
}

func (e *ErrDescription) Error() string {
	if bool(e.Fatal) {
		return fmt.Sprintf("FATAL [%s] %s", e.MessageType, e.Name)
	}
	return fmt.Sprintf("[%s] %s", e.MessageType, e.Name)
}

// ErrAck add some error context patterns to the list of acknowledged errors.
// These errors will be ignored during the next public method call or for the
// time of the transaction.
func (c *Conn) ErrAck(errs []ErrDescription) error {
	return c.Request("err_ack", nil, errs)
}

// ErrAckAll expands to the catch-all pattern {}
func (c *Conn) ErrAckAll() error {
	return c.Request("err_ack", nil, "all")
}

// ErrAckLast expands to the result of ErrList()
func (c *Conn) ErrAckLast() error {
	return c.Request("err_ack", nil, "last")
}

// ErrAckNone clears the list of patterns.
func (c *Conn) ErrAckNone() error {
	return c.Request("err_ack", nil, "none")
}

// ErrIsFatal tells whether the last public method call had fatal errors.
// Returns the number of fatal errors during the last method call.
func (c *Conn) ErrIsFatal() (uint64, error) {
	var num uint64
	err := c.Request("err_is_fatal", &num)
	return num, err
}

// ErrIsNoack tells whether all errors during the last public method call
// had been acknowledged. Returns The number of non-acknowledged errors during
// the last method call.
func (c *Conn) ErrIsNoack() (uint64, error) {
	var num uint64
	err := c.Request("err_is_noack", &num)
	return num, err
}

// ErrList lists the errors that occurred since the last write transaction, or,
// when not in transaction, during the last transanction.
func (c *Conn) ErrList() ([]ErrDescription, error) {
	var errors []ErrDescription
	err := c.Request("err_list", &errors)
	return errors, err
}

// ErrListFatal lists all fatal errors that occurred since the last write
// transaction, or, when not in transaction, during the last transanction.
func (c *Conn) ErrListFatal() ([]ErrDescription, error) {
	var errors []ErrDescription
	err := c.Request("err_list_fatal", &errors)
	return errors, err
}

// ErrListNoAck lists unacknowledged errors that occurred since the last write
// transaction, or, when not in transaction, during the last transanction.
func (c *Conn) ErrListNoAck() ([]ErrDescription, error) {
	var errors []ErrDescription
	err := c.Request("err_list_noack", &errors)
	return errors, err
}

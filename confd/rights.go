// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// GetRights checks the rights of the currently logged in user.
func (c *Conn) GetRights() ([]string, error) {
	var rights []string
	err := c.Request("get_rights", &rights)
	return rights, err
}

// HasRight checks if the current user has the given right
func (c *Conn) HasRight(right string) (bool, error) {
	var ok Bool
	err := c.Request("get_rights", &ok, right)
	return bool(ok), err
}

// HasOneOfRights checks if the current user has one of the given rights
func (c *Conn) HasOneOfRights(rights []string) (bool, error) {
	var ok Bool
	err := c.Request("get_rights", &ok, rights)
	return bool(ok), err
}

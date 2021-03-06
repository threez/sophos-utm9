// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// Export represents an exported confd function
type Export struct {
	Write  Bool     `json:"write"`
	Deny   Bool     `json:"deny"`
	Module string   `json:"module"`
	Class  string   `json:"class"`
	Rights []string `json:"rights"`
	Doc    string   `json:"doc"`
}

// Exports returns all available exports (see definition of export)
func (c *Conn) Exports() (map[string]Export, error) {
	response := make(map[string]Export)
	err := c.Request("get_exports", &response)
	return response, err
}

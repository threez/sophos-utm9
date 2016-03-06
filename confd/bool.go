// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"encoding/json"
	"strings"
)

// Bool are not true or false but can be strings, arrays, maps, and ints
type Bool bool

// UnmarshalJSON handles unmarshaling of confd bools
func (b *Bool) UnmarshalJSON(bytes []byte) (err error) {
	var decoded interface{}
	err = json.Unmarshal(bytes, &decoded)
	if err != nil {
		return err
	}
	switch tv := decoded.(type) {
	case string:
		newValue := strings.Compare(tv, "") == 0
		*b = Bool(newValue)
	case float64:
		newValue := tv == 1
		*b = Bool(newValue)
	default:
		*b = true
	}
	return nil
}

// MarshalJSON marshals the bool as confd
func (b Bool) MarshalJSON() (bytes []byte, err error) {
	value := BoolValue(bool(b))
	bytes, err = json.Marshal(value)
	return
}

// BoolValue returns the confd representation of a bool
func BoolValue(value bool) int {
	if value {
		return 1
	}
	return 0
}

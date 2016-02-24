package confd

import (
	"encoding/json"
	"strings"
)

// Confd bools are not true or false but can be strings, arrays, maps, and ints
type ConfdBool bool

// Handles unmarshaling of confd bools
func (value *ConfdBool) UnmarshalJSON(bytes []byte) (err error) {
	var decoded interface{}
	err = json.Unmarshal(bytes, &decoded)
	if err != nil {
		return err
	}
	switch tv := decoded.(type) {
	case string:
		newValue := strings.Compare(tv, "") == 0
		*value = ConfdBool(newValue)
	case int:
		newValue := tv == 1
		*value = ConfdBool(newValue)
	default:
		*value = true
	}
	return nil
}

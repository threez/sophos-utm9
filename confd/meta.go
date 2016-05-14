// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

import (
	"encoding/json"
)

// AttrConstraint describes the contraints of an object attribute
type AttrConstraint struct {
	ISA              string          `json:"_isa"`
	Type             string          `json:"_type"`
	Class            string          `json:"_class"`
	DeleteWithParent Bool            `json:"_delete_with_parent"`
	Regex            string          `json:"_regex"`
	Require          string          `json:"_require"`
	Types            []string        `json:"_types"`
	NotTypes         []string        `json:"_not_types"`
	Keys             *AttrConstraint `json:"_keys"`
	Values           interface{}     `json:"_values"`
	Limits           []string        `json:"_limits"`
	Default          interface{}     `json:"_default"`
	NameTemplate     string
}

// AttrConstraintWrapper helps to break the JSON unmarshal loop
type AttrConstraintWrapper AttrConstraint

// UnmarshalJSON parses the json in AttrConstraints or templates
func (c *AttrConstraintWrapper) UnmarshalJSON(bytes []byte) (err error) {
	var cor AttrConstraint
	err = json.Unmarshal(bytes, &cor)
	if err != nil {
		err = json.Unmarshal(bytes, &cor.NameTemplate)
	}
	*c = AttrConstraintWrapper(cor)

	return err
}

// AttributeDefinition contains all attributes for a given type
type AttributeDefinition map[string]AttrConstraintWrapper

// TypeDefinition contains all type definitions for a given class
type TypeDefinition map[string]AttributeDefinition

// ObjectMetaTree contains all class definitions
type ObjectMetaTree map[string]TypeDefinition

// NodeTree contains all nodes and values in a single structure
type NodeTree map[string]interface{}

// GetMetaObjects returns objects meta-information.
func (c *Conn) GetMetaObjects() (ObjectMetaTree, error) {
	var ret ObjectMetaTree
	err := c.Request("get_meta_objects", &ret)
	return ret, err
}

// GetMetaNodes returns complete nodes information.
func (c *Conn) GetMetaNodes() (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := c.Request("get_nodes", &ret)
	return ret, err
}

// GetObjectClasses returns all available object classes
func (c *Conn) GetObjectClasses() ([]string, error) {
	var ret []string
	err := c.Request("get_object_classes", &ret)
	return ret, err
}

// GetObjectTypes returns a list of types for the given class name
func (c *Conn) GetObjectTypes(class string) ([]string, error) {
	var ret []string
	err := c.Request("get_object_types", &ret, class)
	return ret, err
}

// GetAvailableNodes returns a list of nodes starting at path
func (c *Conn) GetAvailableNodes(path ...NodeName) ([]string, error) {
	var ret []string
	args := pathToArgs(path)
	err := c.Request("get_nodes", &ret, args...)
	return ret, err
}

// GetMeta returns a map containing all possible nodes and their values
func (c *Conn) GetMeta() (NodeTree, error) {
	var ret NodeTree
	err := c.Request("get_meta", &ret)
	return ret, err
}

// GetScalars returns the available scalar values in the path
func (c *Conn) GetScalars(path ...NodeName) ([]string, error) {
	var ret []string
	args := pathToArgs(path)
	err := c.Request("get_scalars", &ret, args...)
	return ret, err
}

// GetArrays returns the available array values in the path
func (c *Conn) GetArrays(path ...NodeName) ([]string, error) {
	var ret []string
	args := pathToArgs(path)
	err := c.Request("get_arrays", &ret, args...)
	return ret, err
}

// Tree returns the next tree with the given name
func (t NodeTree) Tree(name string) NodeTree {
	return NodeTree(t[name].(map[string]interface{}))
}

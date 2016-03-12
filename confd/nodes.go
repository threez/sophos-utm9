// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// NodeName the name of a node
type NodeName string

// Node representation of node data
type Node map[NodeName]interface{}

// NodeValue representation of one node value
type NodeValue interface{}

// NodePath representation of the path to a Node / NodeValue
type NodePath []NodeName

// GetNode 5ead node data. Returned data type depends on called node.
func (c *Conn) GetNode(path ...NodeName) (Node, error) {
	var node Node
	err := c.Request("get", &node, pathToArgs(path)...)
	return node, err
}

// GetNodeValue 5ead node data. Returned data type depends on called node.
func (c *Conn) GetNodeValue(path ...NodeName) (NodeValue, error) {
	var node NodeValue
	err := c.Request("get", &node, pathToArgs(path)...)
	if err == ErrReturnCode {
		err = nil // ignore 0 return value as failure
	}
	return node, err
}

// GetAffectedNodes get a list of nodes that directly or indirectly use a list
// of given objects.
func (c *Conn) GetAffectedNodes(ref string) ([]NodePath, error) {
	var paths []NodePath
	err := c.Request("get_affected_nodes", &paths, ref)
	return paths, err
}

// ResetNode reset a node in the main tree to its default value.
// Returns true if successful, false otherwise
func (c *Conn) ResetNode(path ...NodeName) (bool, error) {
	var ok Bool
	err := c.Request("reset", &ok, pathToArgs(path)...)
	return bool(ok), err
}

// SetNode set node data in the main tree.
func (c *Conn) SetNode(node Node, path ...NodeName) (bool, error) {
	return c.SetNodeValue(node, path...)
}

// SetNodeValue set node data in the main tree.
func (c *Conn) SetNodeValue(node NodeValue, path ...NodeName) (bool, error) {
	var ok Bool
	args := make([]interface{}, len(path)+1)
	args[0] = node
	copy(args[1:], pathToArgs(path))
	err := c.Request("set", &ok, args...)
	return bool(ok), err
}

// GetNodes list of sub-nodes for a given node.
func (c *Conn) GetNodes(path ...NodeName) ([]NodeName, error) {
	var names []NodeName
	err := c.Request("get_nodes", &names, pathToArgs(path)...)
	return names, err
}

// pathToArgs converts the path of the node into interface{} values to
// comply with the request interface
func pathToArgs(path NodePath) []interface{} {
	args := make([]interface{}, len(path))
	for i, p := range path {
		args[i] = p
	}
	return args
}

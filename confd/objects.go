// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// ObjectMeta confd object metadata
type ObjectMeta struct {
	Ref      string `json:"ref"`
	Class    string `json:"class"`
	Type     string `json:"type"`
	Hidden   Bool   `json:"hidden"`
	Lock     string `json:"lock"`
	Nodel    string `json:"nodel"`
	Autoname Bool   `json:"autoname"`
}

// AnyObject a type that works with any confd object
type AnyObject struct {
	ObjectMeta
	Data map[string]interface{} `json:"data"`
}

// ChangeObject changes the object ref attributes
func (c *Conn) ChangeObject(ref string, attributes interface{}) (err error) {
	_, err = c.SimpleRequest("change_object", ref, attributes)
	return
}

// GetAnyObject returns a AnyObject for the given ref or nil
func (c *Conn) GetAnyObject(ref string) (*AnyObject, error) {
	response := new(AnyObject)
	err := c.Request("get_object", response, ref)
	return response, err
}

// DelObject deletes an object by ref
func (c *Conn) DelObject(ref string) (bool, error) {
	var ok Bool
	err := c.Request("del_object", &ok, ref)
	return bool(ok), err
}

// GetAffectedObjects get a list of objects that directly or indirectly use a
// list of given objects.
// Note - Since all objects carry references to themselves, the list submitted
// in the argument will also be included in the returned list.
func (c *Conn) GetAffectedObjects(refs []string) ([]string, error) {
	var affected []string
	err := c.Request("get_affected_objects", &affected, refs)
	return affected, err
}

// GetAllObjects returns all confd stored conf objects
func (c *Conn) GetAllObjects() ([]AnyObject, error) {
	return c.FilterObjects().Get()
}

// Copyright 2016 Vincent Landgraf. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confd

// ObjectMeta confd object metadata
type ObjectMeta struct {
	Ref      string `json:"ref,omitempty"`
	Class    string `json:"class,omitempty"`
	Type     string `json:"type,omitempty"`
	Nodel    string `json:"nodel,omitempty"`
	Hidden   Bool   `json:"hidden"`
	Lock     Bool   `json:"lock"`
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
	return err
}

// GetAnyObject returns an AnyObject for the given ref or nil
func (c *Conn) GetAnyObject(ref string) (*AnyObject, error) {
	response := new(AnyObject)
	return response, c.GetObject(ref, response)
}

// GetObject returns object for the given ref or nil
func (c *Conn) GetObject(ref string, object interface{}) error {
	err := c.Request("get_object", object, ref)
	return err
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

// LockObject sets the lockstate of the object to locked
func (c *Conn) LockObject(ref string) error {
	_, err := c.SimpleRequest("lock_object", ref, "user")
	return err
}

// UnlockObject sets the lockstate of the object to unlocked
func (c *Conn) UnlockObject(ref string) error {
	_, _ = c.SimpleRequest("lock_override", 1)
	_, err := c.SimpleRequest("lock_object", ref, BoolValue(false))
	_, _ = c.SimpleRequest("lock_override", 0)
	return err
}

// MoveObject change the reference string of an existing object,
// keeping all places where it is used consistent.
func (c *Conn) MoveObject(oldRef string, newRef string) error {
	_, err := c.SimpleRequest("move_object", oldRef, newRef)
	return err
}

// ResetObject reset an object to its state in the default storage.
func (c *Conn) ResetObject(ref string) error {
	_, err := c.SimpleRequest("reset_object", ref)
	return err
}

// SetObject create or update an object.
// When fuzzyName is true and the name is already taken, append the string
// " (2)" to the name and increment the number until a free name is found.
// Returns the ref of the created object
func (c *Conn) SetObject(obj interface{}, fuzzyName bool) (string, error) {
	ref, err := c.SimpleRequest("set_object", obj)
	if err != nil {
		return "", err
	}
	return (*ref.(*interface{})).(string), err
}

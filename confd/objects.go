package confd

// ChangeObject changes the object ref attributes
func (c *Conn) ChangeObject(ref string, attributes interface{}) (err error) {
	_, err = c.SimpleRequest("change_object", ref, attributes)
	return
}

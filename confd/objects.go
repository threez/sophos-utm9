package confd

func (c *Conn) ChangeObject(ref string, params Params) (err error) {
	_, err = c.SimpleRequest("change_object", []interface{}{ref, params})
	return
}

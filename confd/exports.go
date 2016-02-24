package confd

type ConfdExport struct {
	Write  ConfdBool `json:"write"`
	Deny   ConfdBool `json:"deny"`
	Module string    `json:"module"`
	Class  string    `json:"class"`
	Rights []string  `json:"rights"`
	Doc    string    `json:"doc"`
}

func (c *ConfdConn) Exports() (map[string]ConfdExport, error) {
	exports := new(struct {
		ConfdResponse
		Result map[string]ConfdExport `json:"result"`
	})
	err := c.Request("get_exports", exports, nil)
	return exports.Result, err
}

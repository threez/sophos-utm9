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
	response := new(struct {
		Response
		Result map[string]Export `json:"result"`
	})
	err := c.Request("get_exports", response)
	if err == nil && response.Error != nil {
		err = response.Error
	}
	return response.Result, err
}

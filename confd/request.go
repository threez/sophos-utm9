package confd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Request is used to construct requests
type Request struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	ID     uint64           `json:"id"`
}

// NewRequest creates a new request object
func NewRequest(method string, params interface{}, id uint64) (req *Request, err error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	rawMsg := json.RawMessage(data)
	r := &Request{method, &rawMsg, id}
	return r, err
}

func (r *Request) String() string {
	params := []byte(*r.Params)
	return fmt.Sprintf("[%d] %s(%s)", r.ID, r.Method, string(params[:]))
}

// HTTP retruns an http request as bytes
func (r *Request) HTTP(host string) (*http.Request, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(r)
	if err != nil {
		return nil, err
	}
	ru, err := http.NewRequest("POST", "/", &buf)
	if err != nil {
		return nil, err
	}
	ru.URL.Host = host
	ru.Header.Set("Content-Type", "application/json")
	ru.Header.Set("User-Agent", "")
	return ru, nil
}

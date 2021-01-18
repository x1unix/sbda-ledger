package ledger

import (
	"encoding/json"
	"fmt"
)

type ErrorResponse struct {
	Status    string `json:"-"`
	ErrorData struct {
		Error string          `json:"error"`
		Data  json.RawMessage `json:"data"`
	} `json:"error"`
}

func (rsp ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", rsp.Status, rsp.ErrorData.Error)
}

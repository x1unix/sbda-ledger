package ledger

import (
	"encoding/json"
	"fmt"
)

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Status     string `json:"-"`
	ErrorData  struct {
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	} `json:"error"`
}

func (rsp ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", rsp.Status, rsp.ErrorData.Message)
}

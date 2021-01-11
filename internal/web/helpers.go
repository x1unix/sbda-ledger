package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// UnmarshalJSON unmarshal request payload to destination value.
//
// Returns API error on failure.
func UnmarshalJSON(r io.ReadCloser, out interface{}) *APIError {
	defer r.Close()
	if err := json.NewDecoder(r).Decode(out); err != nil {
		var msg string
		if err == io.EOF {
			msg = "empty request"
		} else {
			msg = fmt.Sprintf("cannot read request: %s", err)
		}

		return &APIError{
			Status:  http.StatusBadRequest,
			Message: msg,
		}
	}
	return nil
}

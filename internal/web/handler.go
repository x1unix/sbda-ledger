package web

import (
	"net/http"
)

type Handler struct {
}

func (h Handler) Echo(rw http.ResponseWriter, r *http.Request) {
	_, _ = rw.Write([]byte("Hello"))
}

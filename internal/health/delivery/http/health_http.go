package http

import (
	"fmt"
	"net/http"

	"source.toby3d.me/toby3d/auth/internal/common"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Handler() http.Handler {
	return http.HandlerFunc(h.handleFunc)
}

func (h *Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
	fmt.Fprint(w, `👌`)
	w.WriteHeader(http.StatusOK)
}

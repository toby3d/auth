package http

import (
	"fmt"
	"net/http"

	"source.toby3d.me/toby3d/auth/internal/common"
	"source.toby3d.me/toby3d/auth/internal/middleware"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.HandlerFunc(middleware.HandlerFunc(h.handleFunc).Intercept(middleware.LogFmt())).ServeHTTP(w, r)
}

func (h *Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(common.HeaderContentType, common.MIMETextPlainCharsetUTF8)
	fmt.Fprint(w, `👌`)
	w.WriteHeader(http.StatusOK)
}

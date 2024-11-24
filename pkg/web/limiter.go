package web

import "net/http"

type Limiter struct {
	h http.Handler
	n int
}

func NewLimiter(h http.Handler, bytes int) http.Handler {
	return &Limiter{h, bytes}
}

func (h *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(h.n))
	h.h.ServeHTTP(w, r)
}

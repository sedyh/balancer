package controller

import (
	"balancer/internal/service"
	"balancer/pkg/conc"
	"balancer/pkg/data"
	"balancer/pkg/str"
	"balancer/pkg/web"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Balancer struct {
	server  *web.Server
	vault   *service.Vault
	upload  *service.SplitUpload
	keylock *conc.KeyLock
}

func NewBalancer(
	addr string,
	limit int,
	timeout time.Duration,
	vault *service.Vault,
	upload *service.SplitUpload,
) (*Balancer, error) {
	e := &Balancer{
		vault:   vault,
		upload:  upload,
		keylock: conc.NewKeyLock(),
	}

	m := http.NewServeMux()
	m.HandleFunc("POST /files/{name}", e.Upload)
	m.HandleFunc("GET /files/{name}", e.Download)

	server, err := web.NewServer(m, addr, limit, timeout)
	if err != nil {
		return nil, fmt.Errorf("serve: %w", err)
	}
	e.server = server

	return e, nil
}

func (e *Balancer) Upload(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !str.Filename.MatchString(name) {
		slog.Error("invalid name format", "name", name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	digest := r.Header.Get("Digest")
	if !str.Digest.MatchString(digest) {
		slog.Error("invalid digest format", "digest", digest)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e.keylock.Lock(name)
	e.keylock.Lock(digest)

	reader := data.NewProgressReader(
		r.Body, int(r.ContentLength),
		data.SlogProgress(name),
	)
	hash, size, err := e.vault.Write(reader, name)
	if err != nil {
		slog.Error("upload", "name", name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if hash != digest {
		slog.Error("corrupted data", "hash", hash, "digest", digest)
		w.WriteHeader(http.StatusBadRequest)
		e.vault.Remove(hash)
		return
	}

	go func() {
		e.upload.Upload(name, hash, size)
		e.keylock.Unlock(name)
		e.keylock.Unlock(digest)
	}()
}

func (e *Balancer) Download(w http.ResponseWriter, r *http.Request) {

}

func (e *Balancer) Close(ctx context.Context) error {
	return e.server.Close(ctx)
}

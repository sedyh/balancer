package controller

import (
	"balancer/internal/service"
	"balancer/pkg/data"
	"balancer/pkg/str"
	"balancer/pkg/web"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Storage struct {
	server *web.Server
	vault  *service.Vault
}

func NewStorage(
	addr string,
	limit int,
	timeout time.Duration,
	vault *service.Vault,
) (*Storage, error) {
	e := &Storage{vault: vault}

	m := http.NewServeMux()
	m.HandleFunc("POST /parts/{name}", e.Save)
	m.HandleFunc("GET /parts/{name}", e.Load)

	server, err := web.NewServer(m, addr, limit, timeout)
	if err != nil {
		return nil, fmt.Errorf("serve: %w", err)
	}
	e.server = server

	return e, nil
}

func (e *Storage) Save(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !str.Filename.MatchString(name) {
		slog.Error("invalid name format", "name", name)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reader := data.NewProgressReader(
		r.Body, int(r.ContentLength),
		data.SlogProgress(name),
	)

	_, _, err := e.vault.Write(reader, name)
	if err != nil {
		slog.Error("upload", "name", name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (e *Storage) Load(w http.ResponseWriter, r *http.Request) {

}

func (e *Storage) Close(ctx context.Context) error {
	return e.server.Close(ctx)
}

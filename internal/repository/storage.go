package repository

import (
	"balancer/pkg/data"
	"balancer/pkg/errs"
	"balancer/pkg/maglev"
	"balancer/pkg/validation"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Storage struct {
	timeout time.Duration
	hasher  *maglev.Hasher
}

func NewStorage(timeout time.Duration, backends []string) *Storage {
	s := &Storage{timeout: timeout}
	s.hasher = maglev.NewHasher(maglev.DefaultPrime)
	s.hasher.AddBackends(backends)
	return s
}

func (s *Storage) Save(name string, part int, r io.Reader, limit int) (e error) {
	flow := fmt.Sprintf("name-%s:part-%d", name, part)
	backend := s.hasher.GetBackend(flow)

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	url := fmt.Sprintf("http://%s/parts/%s", backend, flow)
	reader := data.NewProgressReader(
		&io.LimitedReader{R: r, N: int64(limit)},
		limit, data.SlogProgress(fmt.Sprintf("%s -> %s", flow, backend)),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer errs.Close(&e, res.Body.Close)

	if !validation.SuccessStatus(res.StatusCode) {
		return fmt.Errorf("error code %d", res.StatusCode)
	}

	return nil
}

func (s *Storage) Backends() int {
	return s.hasher.BackendsNum()
}

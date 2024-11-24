package repository

import (
	"balancer/pkg/data"
	"balancer/pkg/errs"
	"balancer/pkg/validation"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Balancer struct {
	base    string
	timeout time.Duration
}

func NewBalancer(base string, timeout time.Duration) *Balancer {
	s := &Balancer{
		base:    base,
		timeout: timeout,
	}
	return s
}

func (s *Balancer) Upload(name, hash string, r io.Reader, limit int) (e error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	url := fmt.Sprintf("http://%s/files/%s", s.base, name)
	reader := data.NewProgressReader(
		&io.LimitedReader{R: r, N: int64(limit)},
		limit, data.SlogProgress(name),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Digest", hash)

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

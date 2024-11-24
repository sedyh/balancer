package service

import (
	"balancer/pkg/data"
	"balancer/pkg/errs"
	"bytes"
	"fmt"
	"io"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

type SplitUpload struct {
	files    FileRepository
	storages StorageRepository
}

func NewSplitUpload(
	files FileRepository,
	storages StorageRepository,
) *SplitUpload {
	u := &SplitUpload{
		files:    files,
		storages: storages,
	}
	return u
}

func (u *SplitUpload) Upload(name, hash string, size int) {
	defer u.files.Remove(hash)

	backends := u.storages.Backends()
	average := size / backends
	smaller := data.PrevPowerOfTwo(int(average))
	remains := size - smaller*backends

	group := &errgroup.Group{}
	for part := 0; part < backends; part++ {
		group.Go(u.stream(name, hash, part, backends, smaller, remains))
	}

	if err := group.Wait(); err != nil {
		slog.Error("upload", "hash", hash, "error", err)
		return
	}

	slog.Info("uploaded", "name", name, "hash", hash)
}

func (u *SplitUpload) stream(name, hash string, part, backends, smaller, remains int) func() error {
	return func() (e error) {
		offset := part * smaller
		reader, err := u.files.Seek(hash, offset)
		if err != nil {
			return fmt.Errorf("seek offset %d: %w", offset, err)
		}
		defer errs.Close(&e, reader.Close)

		limit := smaller
		if part == backends-1 {
			limit = remains
		}
		var prepend []byte
		if part == 0 {
			prepend = []byte{byte(backends)}
			limit++
		}
		combined := io.MultiReader(bytes.NewReader(prepend), reader)
		if err := u.storages.Save(name, part, combined, limit); err != nil {
			return fmt.Errorf("save on storage %d: %w", offset, err)
		}
		return nil
	}
}

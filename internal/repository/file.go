package repository

import (
	"balancer/pkg/data"
	"balancer/pkg/errs"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lithammer/shortuuid/v4"
)

type File struct {
	path string
}

func NewFile(path string) *File {
	return &File{path: path}
}

func (f *File) Write(r io.Reader) (hash string, size int, e error) {
	if err := data.EnsureDir(f.path); err != nil {
		return "", 0, fmt.Errorf("create dir: %w", err)
	}

	was := filepath.Join(f.path, shortuuid.New())
	hash, size, err := data.Stream(was, r)
	defer errs.Close(&e, data.SilentRemoveCloser(was))
	if err != nil {
		return "", 0, fmt.Errorf("stream data: %w", err)
	}

	now := filepath.Join(f.path, hash)
	if err = os.Rename(was, now); err != nil {
		return "", 0, fmt.Errorf("rename file: %w", err)
	}

	return hash, size, nil
}

func (f *File) Read(hash string) (r io.ReadCloser, e error) {
	now := filepath.Join(f.path, hash)
	file, err := os.Open(now)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

func (f *File) Seek(hash string, offset int) (r io.ReadCloser, e error) {
	now := filepath.Join(f.path, hash)
	file, err := os.Open(now)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	if _, err = file.Seek(int64(offset), 0); err != nil {
		return nil, fmt.Errorf("seek file: %w", err)
	}

	return file, nil
}

func (f *File) Import(path string) (hash string, size int, e error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("open file: %w", err)
	}
	defer errs.Close(&e, file.Close)

	return f.Write(file)
}

func (f *File) Remove(hash string) {
	now := filepath.Join(f.path, hash)
	data.SilentRemove(now)
}

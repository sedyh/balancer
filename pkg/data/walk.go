package data

import (
	"balancer/pkg/errs"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Walk(path, name string) (string, error) {
	var location string
	if err := filepath.WalkDir(path, visit(name, &location)); err != nil {
		return "", err
	}

	if location == "" {
		return "", fmt.Errorf("search %s: %w", name, fs.ErrNotExist)
	}

	return location, nil
}

func visit(name string, location *string) fs.WalkDirFunc {
	return func(p string, f fs.DirEntry, e error) error {
		if e != nil {
			return e
		}

		if f.Name() == name {
			*location = p
			return filepath.SkipAll
		}

		return nil
	}
}

func Stream(path string, r io.Reader) (hash string, size int, err error) {
	const buffer = 32000

	f, e := os.Create(path)
	if e != nil {
		return "", 0, fmt.Errorf("create file: %w", e)
	}
	defer errs.Close(&err, f.Close)

	n, s := 0, 0
	h := sha256.New()
	buf := make([]byte, buffer)
	for {
		n, e = r.Read(buf)
		if n > 0 {
			if _, e = h.Write(buf[:n]); e != nil {
				return "", 0, fmt.Errorf("write hash: %w", e)
			}
			if _, e = f.Write(buf[:n]); e != nil {
				return "", 0, fmt.Errorf("write file: %w", e)
			}
			s += n
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			return "", 0, fmt.Errorf("read file: %w", e)
		}
	}
	sum := h.Sum(nil)

	return hex.EncodeToString(sum[:]), s, nil
}

func Exist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func EnsureDir(path string) error {
	if !Exist(path) {
		return os.MkdirAll(path, 0o700)
	}

	return nil
}

func SilentRemoveCloser(path string) func() error {
	return func() error {
		SilentRemove(path)
		return nil
	}
}

func SilentRemove(path string) {
	_ = os.Remove(path)
}

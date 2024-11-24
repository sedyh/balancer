package service

import (
	"io"
)

type Vault struct {
	files FileRepository
}

func NewVault(files FileRepository) *Vault {
	return &Vault{files: files}
}

func (f *Vault) Write(r io.Reader, name string) (string, int, error) {
	return f.files.Write(r)
}

func (f *Vault) Read(hash string) (r io.ReadCloser, e error) {
	return f.files.Read(hash)
}

func (f *Vault) Remove(hash string) {
	f.files.Remove(hash)
}

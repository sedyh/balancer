package service

import "io"

type FileRepository interface {
	Write(r io.Reader) (hash string, size int, e error)
	Read(hash string) (r io.ReadCloser, e error)
	Seek(hash string, offset int) (r io.ReadCloser, e error)
	Import(path string) (hash string, size int, e error)
	Remove(hash string)
}

type StorageRepository interface {
	Save(name string, part int, r io.Reader, limit int) (e error)
	Backends() int
}

type BalancerRepository interface {
	Upload(name, hash string, r io.Reader, limit int) (e error)
}

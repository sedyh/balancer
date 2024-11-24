package service

import (
	"balancer/pkg/errs"
	"fmt"
)

const limit = 10000000000

type PlainUpload struct {
	file     FileRepository
	balancer BalancerRepository
}

func NewPlainUpload(
	file FileRepository,
	balancer BalancerRepository,
) *PlainUpload {
	u := &PlainUpload{
		file:     file,
		balancer: balancer,
	}
	return u
}

func (u *PlainUpload) Upload(name string) (e error) {
	hash, size, err := u.file.Import(name)
	if err != nil {
		return fmt.Errorf("import file: %w", err)
	}
	defer u.file.Remove(hash)

	if size > limit {
		return fmt.Errorf("file size is too large, should be lower than 10GB")
	}

	reader, err := u.file.Read(hash)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	defer errs.Close(&e, reader.Close)

	if err := u.balancer.Upload(name, hash, reader, size); err != nil {
		return fmt.Errorf("upload file: %w", err)
	}

	return nil
}

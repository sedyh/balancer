package errs

import (
	"errors"
)

func Close(dest *error, c func() error) {
	*dest = errors.Join(*dest, c())
}

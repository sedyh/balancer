package validation

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func SuccessStatus(code int) bool {
	return code >= 200 && code <= 299
}

func Pretty(err error) error {
	var vs validator.ValidationErrors
	if !errors.As(err, &vs) {
		return err
	}

	return separate(vs)
}

func separate(vs validator.ValidationErrors) error {
	fields := make([]string, 0, len(vs))
	for _, v := range vs {
		key := strings.ReplaceAll(strings.ToLower(v.Namespace()), ".", "_")
		tag := strings.ToLower(v.ActualTag())
		par := v.Param()

		var buf strings.Builder
		buf.WriteString(opt("invalid ", key))
		buf.WriteString(opt(" should be ", tag))
		buf.WriteString(opt("=", par))
		fields = append(fields, buf.String())
	}

	return errors.New(strings.Join(fields, ", "))
}

func opt(str, v string) string {
	if v == "" {
		return ""
	}

	return str + v
}

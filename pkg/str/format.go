package str

import "regexp"

var (
	Filename = regexp.MustCompile(`^[^|/\s]+$`)
	Digest   = regexp.MustCompile(`(?i)[A-F0-9]{64}`)
)

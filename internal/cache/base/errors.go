package base

import "errors"

var (
	// ErrNotFound is returned when the key is not found
	ErrNotFound = errors.New("not found")
)

package kv

import (
	"errors"
)

var (
	ErrInvalidType      = errors.New("invalid type")
	ErrValuerEmptySlice = errors.New("empty byte slice")
)

var ErrKeyNil = errors.New("key does not exist")

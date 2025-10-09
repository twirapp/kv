package kv

import (
	"errors"
)

var ErrInvalidType = errors.New("invalid type")
var ErrValuerEmptySlice = errors.New("empty byte slice")

var KeyNil = errors.New("key does not exist")

package kv

import (
	"context"

	kvoptions "github.com/twirapp/kv/options"
)

type KV interface {
	Get(ctx context.Context, key string) Valuer
	Set(ctx context.Context, key string, value any, options ...kvoptions.Option) error
	SetMany(ctx context.Context, values []SetMany) error
	Delete(ctx context.Context, key string) error
	DeleteMany(ctx context.Context, keys []string) error
	Exists(ctx context.Context, key string) (bool, error)
	// ExistsMany returns a slice of bools indicating whether each key exists.
	// The order of the bools corresponds to the order of the keys provided.
	ExistsMany(ctx context.Context, keys []string) ([]bool, error)
	GetKeysByPattern(ctx context.Context, pattern string) ([]string, error)
}

type Valuer interface {
	Int() (int64, error)
	String() (string, error)
	Bytes() ([]byte, error)
	Bool() (bool, error)
	Float() (float64, error)
	Scan(dest any) error
	Err() error
}

type SetMany struct {
	Key     string
	Value   any
	Options []kvoptions.Option
}

package kvotter

import (
	"context"
	"fmt"
	"strings"

	"github.com/maypok86/otter/v2"
	"github.com/twirapp/kv"
	"github.com/twirapp/kv/internal/matchpattern"
	"github.com/twirapp/kv/internal/tobytes"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
)

var _ kv.KV = (*Otter)(nil)

func New() *Otter {
	cache := otter.Must(&otter.Options[string, []byte]{})

	return &Otter{o: cache}
}

type Otter struct {
	o *otter.Cache[string, []byte]
}

func (c *Otter) Get(_ context.Context, key string) kv.Valuer {
	v, ok := c.o.GetIfPresent(key)
	if !ok {
		return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
	}

	return &kvvaluer.Valuer{Value: v}
}

func (c *Otter) Set(_ context.Context, key string, value any, options ...kvoptions.Option) error {
	b, err := tobytes.ToBytes(value)
	if err != nil {
		return err
	}

	v, ok := c.o.Set(key, b)
	if !ok && v == nil {
		return fmt.Errorf("failed to set value for key %s", key)
	}

	o := kvoptions.Construct(options...)
	if o.Expire > 0 {
		c.o.SetExpiresAfter(key, o.Expire)
	}

	return nil
}

func (c *Otter) SetMany(ctx context.Context, values []kv.SetMany) error {
	for _, v := range values {
		if err := c.Set(ctx, v.Key, v.Value, v.Options...); err != nil {
			return err
		}
	}

	return nil
}

func (c *Otter) Delete(_ context.Context, key string) error {
	_, _ = c.o.Invalidate(key)

	return nil
}

func (c *Otter) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (c *Otter) Exists(_ context.Context, key string) (bool, error) {
	_, ok := c.o.GetIfPresent(key)
	return ok, nil
}

func (c *Otter) ExistsMany(_ context.Context, keys []string) ([]bool, error) {
	results := make([]bool, len(keys))
	for i, key := range keys {
		_, ok := c.o.GetIfPresent(key)
		results[i] = ok
	}

	return results, nil
}

func (c *Otter) GetKeysByPattern(_ context.Context, pattern string) ([]string, error) {
	var (
		keys         []string
		patternParts = strings.Split(pattern, ":")
	)

	for key := range c.o.Keys() {
		keyParts := strings.Split(key, ":")
		if matchpattern.MatchPattern(patternParts, keyParts) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

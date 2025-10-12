package kvmemcached

import (
	"context"
	"errors"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/twirapp/kv"
	"github.com/twirapp/kv/internal/tobytes"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
)

var _ kv.KV = (*KvMemcached)(nil)

type KvMemcached struct {
	mc *memcache.Client
}

func New(mc *memcache.Client) *KvMemcached {
	return &KvMemcached{
		mc: mc,
	}
}

func (c *KvMemcached) Get(_ context.Context, key string) kv.Valuer {
	item, err := c.mc.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
		}
		return &kvvaluer.Valuer{Error: err}
	}
	return &kvvaluer.Valuer{Value: item.Value}
}

func (c *KvMemcached) Set(
	_ context.Context,
	key string,
	value any,
	options ...kvoptions.Option,
) error {
	o := kvoptions.Construct(options...)
	valueBytes, err := tobytes.ToBytes(value)
	if err != nil {
		return fmt.Errorf("failed to convert value to bytes: %w", err)
	}
	item := &memcache.Item{
		Key:   key,
		Value: valueBytes,
	}
	if o.Expire > 0 {
		item.Expiration = int32(o.Expire.Seconds())
	}
	err = c.mc.Set(item)
	return err
}

func (c *KvMemcached) SetMany(ctx context.Context, values []kv.SetMany) error {
	for _, v := range values {
		if err := c.Set(ctx, v.Key, v.Value, v.Options...); err != nil {
			return err
		}
	}
	return nil
}

func (c *KvMemcached) Delete(_ context.Context, key string) error {
	err := c.mc.Delete(key)
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}
	return nil
}

func (c *KvMemcached) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (c *KvMemcached) Exists(_ context.Context, key string) (bool, error) {
	_, err := c.mc.Get(key)
	if err != nil {
		if errors.Is(err, memcache.ErrCacheMiss) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *KvMemcached) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	results := make([]bool, len(keys))
	for i, key := range keys {
		exists, err := c.Exists(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("error checking existence for key %s: %w", key, err)
		}
		results[i] = exists
	}
	return results, nil
}

func (c *KvMemcached) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	return nil, fmt.Errorf("GetKeysByPattern is not supported in Memcached")
}

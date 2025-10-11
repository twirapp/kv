package kvinmemory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/twirapp/kv"
	"github.com/twirapp/kv/internal/tobytes"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
)

var _ kv.KV = (*InMemory)(nil)

type inMemoryValue struct {
	value  []byte
	expire time.Duration
}

type InMemory struct {
	storage map[string]inMemoryValue
	mu      sync.RWMutex
}

func New() *InMemory {
	return &InMemory{
		storage: make(map[string]inMemoryValue),
		mu:      sync.RWMutex{},
	}
}

func (c *InMemory) Get(_ context.Context, key string) kv.Valuer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.storage[key]
	if !ok {
		return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
	}

	return &kvvaluer.Valuer{Value: v.value}
}

func (c *InMemory) Set(_ context.Context, key string, value any, options ...kvoptions.Option) error {
	b, err := tobytes.ToBytes(value)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	o := kvoptions.Construct(options...)
	c.storage[key] = inMemoryValue{
		value:  b,
		expire: o.Expire,
	}

	return nil
}

func (c *InMemory) SetMany(_ context.Context, values []kv.SetMany) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, v := range values {
		b, err := tobytes.ToBytes(v.Value)
		if err != nil {
			return err
		}

		o := kvoptions.Construct(v.Options...)
		c.storage[v.Key] = inMemoryValue{
			value:  b,
			expire: o.Expire,
		}
	}

	return nil
}

func (c *InMemory) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.storage, key)

	return nil
}

func (c *InMemory) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

func (c *InMemory) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, ok := c.storage[key]
	return ok, nil
}

func (c *InMemory) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]bool, len(keys))
	for i, key := range keys {
		_, ok := c.storage[key]
		results[i] = ok
	}

	return results, nil
}

func (c *InMemory) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var (
		keys         []string
		patternParts = strings.Split(pattern, ":")
	)

	for key := range c.storage {
		keyParts := strings.Split(key, ":")
		if matchPattern(patternParts, keyParts) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func matchPattern(patternParts, keyParts []string) bool {
	if len(patternParts) != len(keyParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if patternParts[i] != "*" && patternParts[i] != keyParts[i] {
			return false
		}
	}
	return true
}

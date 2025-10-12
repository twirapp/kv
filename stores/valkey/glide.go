package valkey

import (
	"context"
	"sync"

	"github.com/twirapp/kv"
	"github.com/twirapp/kv/internal/tobytes"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/constants"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

var _ kv.KV = (*GlideStore)(nil)

func NewGlide(client *glide.Client) *GlideStore {
	return &GlideStore{
		cl: client,
	}
}

type GlideStore struct {
	cl *glide.Client
}

func (c *GlideStore) Get(ctx context.Context, key string) kv.Valuer {
	result, err := c.cl.Get(ctx, key)
	if err != nil {
		return &kvvaluer.Valuer{Error: err}
	}
	if result.IsNil() {
		return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
	}

	return &kvvaluer.Valuer{Value: []byte(result.Value())}
}

func (c *GlideStore) Set(ctx context.Context, key string, value any, options ...kvoptions.Option) error {
	o := kvoptions.Construct(options...)

	bytes, err := tobytes.ToBytes(value)
	if err != nil {
		return err
	}

	_, err = c.cl.Set(ctx, key, string(bytes))
	if err != nil {
		return err
	}

	if o.Expire > 0 {
		_, err = c.cl.Expire(ctx, key, o.Expire)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *GlideStore) SetMany(ctx context.Context, values []kv.SetMany) error {
	setMap := make(map[string]string, len(values))
	for _, v := range values {
		bytes, err := tobytes.ToBytes(v.Value)
		if err != nil {
			return err
		}
		setMap[v.Key] = string(bytes)
	}

	_, err := c.cl.MSet(ctx, setMap)
	if err != nil {
		return err
	}

	for _, v := range values {
		o := kvoptions.Construct(v.Options...)
		if o.Expire > 0 {
			_, err = c.cl.Expire(ctx, v.Key, o.Expire)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *GlideStore) Delete(ctx context.Context, key string) error {
	_, err := c.cl.Del(ctx, []string{key})
	return err
}

func (c *GlideStore) DeleteMany(ctx context.Context, keys []string) error {
	_, err := c.cl.Del(ctx, keys)
	return err
}

func (c *GlideStore) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.cl.Exists(ctx, []string{key})
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (c *GlideStore) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	var (
		results = make([]bool, len(keys))
		wg      sync.WaitGroup
	)

	wg.Add(len(keys))

	for i, key := range keys {
		go func(i int, key string) {
			defer wg.Done()
			exists, err := c.Exists(ctx, key)
			if err != nil {
				results[i] = false
				return
			}
			results[i] = exists
		}(i, key)
	}

	wg.Wait()
	return results, nil
}

func (c *GlideStore) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	opts := options.NewScanOptions().SetMatch(pattern).SetType(constants.ObjectTypeString)
	result, err := c.cl.ScanWithOptions(ctx, models.NewCursor(), *opts)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

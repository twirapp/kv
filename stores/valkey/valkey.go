package valkey

import (
	"context"
	"errors"

	"github.com/twirapp/kv"
	"github.com/twirapp/kv/internal/tobytes"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
	"github.com/valkey-io/valkey-go"
)

var _ kv.KV = (*ValkeyStore)(nil)

func New(client valkey.Client) *ValkeyStore {
	return &ValkeyStore{
		cl: client,
	}
}

type ValkeyStore struct {
	cl valkey.Client
}

func (c *ValkeyStore) Get(ctx context.Context, key string) kv.Valuer {
	result, err := c.cl.Do(ctx, c.cl.B().Get().Key(key).Build()).AsBytes()
	if err != nil {
		if errors.Is(err, valkey.Nil) {
			return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
		}
		return &kvvaluer.Valuer{Error: err}
	}

	return &kvvaluer.Valuer{Value: result}
}

func (c *ValkeyStore) Set(ctx context.Context, key string, value any, options ...kvoptions.Option) error {
	o := kvoptions.Construct(options...)

	bytes, err := tobytes.ToBytes(value)
	if err != nil {
		return err
	}

	cmd := c.cl.B().Set().Key(key).Value(string(bytes))

	var finalCmd valkey.Completed
	if o.Expire > 0 {
		finalCmd = cmd.Ex(o.Expire).Build()
	} else {
		finalCmd = cmd.Build()
	}

	err = c.cl.Do(ctx, finalCmd).Error()
	if err != nil {
		return err
	}

	return nil
}

func (c *ValkeyStore) SetMany(ctx context.Context, values []kv.SetMany) error {
	cmds := make(valkey.Commands, 0, len(values))
	for _, v := range values {
		o := kvoptions.Construct(v.Options...)

		bytes, err := tobytes.ToBytes(v.Value)
		if err != nil {
			return err
		}

		cmd := c.cl.B().Set().Key(v.Key).Value(string(bytes))

		var finalCmd valkey.Completed
		if o.Expire > 0 {
			finalCmd = cmd.Ex(o.Expire).Build()
		} else {
			finalCmd = cmd.Build()
		}

		cmds = append(cmds, finalCmd)
	}

	for _, resp := range c.cl.DoMulti(ctx, cmds...) {
		if err := resp.Error(); err != nil {
			return err
		}
	}

	return nil
}

func (c *ValkeyStore) Delete(ctx context.Context, key string) error {
	err := c.cl.Do(ctx, c.cl.B().Del().Key(key).Build()).Error()
	return err
}

func (c *ValkeyStore) DeleteMany(ctx context.Context, keys []string) error {
	err := c.cl.Do(ctx, c.cl.B().Del().Key(keys...).Build()).Error()
	return err
}

func (c *ValkeyStore) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.cl.Do(ctx, c.cl.B().Exists().Key(key).Build()).AsInt64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (c *ValkeyStore) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	cmds := make(valkey.Commands, 0, len(keys))
	for _, key := range keys {
		cmd := c.cl.B().Exists().Key(key)
		cmds = append(cmds, cmd.Build())
	}

	results := make([]bool, 0, len(keys))
	for _, resp := range c.cl.DoMulti(ctx, cmds...) {
		result, err := resp.AsInt64()
		if err != nil {
			return nil, err
		}
		results = append(results, result == 1)
	}

	return results, nil
}

func (c *ValkeyStore) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	result, err := c.cl.Do(ctx, c.cl.B().Scan().Cursor(0).Match(pattern).Build()).AsScanEntry()
	if err != nil {
		return nil, err
	}
	return result.Elements, nil
}

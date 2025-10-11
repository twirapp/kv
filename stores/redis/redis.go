package kvredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	kv "github.com/twirapp/kv"
	kvoptions "github.com/twirapp/kv/options"
	kvvaluer "github.com/twirapp/kv/valuer"
)

var _ kv.KV = (*KvRedis)(nil)

type KvRedis struct {
	r *redis.Client
}

func New(r *redis.Client) *KvRedis {
	return &KvRedis{
		r: r,
	}
}

func (c *KvRedis) Get(ctx context.Context, key string) kv.Valuer {
	result, err := c.r.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return &kvvaluer.Valuer{Error: kv.ErrKeyNil}
		}
		return &kvvaluer.Valuer{Error: err}
	}

	return &kvvaluer.Valuer{Value: result}
}

func (c *KvRedis) Set(
	ctx context.Context,
	key string,
	value any,
	options ...kvoptions.Option,
) error {
	o := kvoptions.Construct(options...)

	return c.r.Set(ctx, key, value, o.Expire).Err()
}

func (c *KvRedis) SetMany(ctx context.Context, values []kv.SetMany) error {
	pipe := c.r.Pipeline()

	for _, v := range values {
		o := kvoptions.Construct(v.Options...)
		if err := pipe.Set(ctx, v.Key, v.Value, o.Expire).Err(); err != nil {
			return err
		}
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *KvRedis) Delete(ctx context.Context, key string) error {
	exists, err := c.Exists(ctx, key)
	if err != nil {
		return err
	}
	if !exists {
		return kv.ErrKeyNil
	}

	return c.r.Del(ctx, key).Err()
}

func (c *KvRedis) DeleteMany(
	ctx context.Context,
	keys []string,
) error {
	return c.r.Del(ctx, keys...).Err()
}

func (c *KvRedis) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.r.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (c *KvRedis) ExistsMany(ctx context.Context, keys []string) ([]bool, error) {
	results := make([]bool, len(keys))

	cmds, err := c.r.Pipelined(
		ctx,
		func(pipe redis.Pipeliner) error {
			for _, key := range keys {
				pipe.Exists(ctx, key)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		exists, err := cmd.(*redis.IntCmd).Result()
		if err != nil {
			return nil, fmt.Errorf("error checking existence for key %s: %w", keys[i], err)
		}

		results[i] = exists == 1
	}

	return results, nil
}

func (c *KvRedis) GetKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	var keys []string

	iter := c.r.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

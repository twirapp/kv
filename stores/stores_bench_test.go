package stores

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twirapp/kv"
	kvredis "github.com/twirapp/kv/stores/redis"
)

var additionalBenchImplementations = []struct {
	name   string
	create func() kv.KV
}{
	{
		name: "Dragonfly (redis client)",
		create: func() kv.KV {
			ctx := context.Background()
			rc, err := tc.Run(
				ctx,
				"docker.dragonflydb.io/dragonflydb/dragonfly:latest",
				tc.WithExposedPorts("6379"),
				tc.WithWaitStrategy(
					wait.ForListeningPort("6379"),
				),
			)
			if err != nil {
				fmt.Printf("Could not start dragonfly container: %v\n", err)
				os.Exit(1)
			}
			connString, err := rc.Endpoint(ctx, "redis")
			if err != nil {
				fmt.Printf("Could not get dragonfly connection string: %v\n", err)
				os.Exit(1)
			}
			rOpts, err := redis.ParseURL(connString)
			if err != nil {
				fmt.Printf("Could not parse dragonfly connection string: %v\n", err)
				os.Exit(1)
			}

			containersLock.Lock()
			containers = append(containers, rc)
			containersLock.Unlock()

			return kvredis.New(redis.NewClient(rOpts))
		},
	},
}

func BenchmarkGet(b *testing.B) {
	impls := append(implementations, additionalBenchImplementations...)

	for _, impl := range impls {
		store := impl.create()
		key := "test_key"
		value := "test_value"

		ctx := context.Background()

		if err := store.Set(ctx, key, value); err != nil {
			b.Fatalf("failed to set initial value: %v", err)
		}

		b.ResetTimer()

		b.Run(impl.name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					v := store.Get(ctx, key)
					casted, err := v.String()
					if err != nil {
						b.Fatalf("failed to get value: %v", err)
					}

					if casted != value {
						b.Fatalf("unexpected value: got %s, want %s", casted, value)
					}
				}
			})
		})
	}
}

func BenchmarkSet(b *testing.B) {
	impls := append(implementations, additionalBenchImplementations...)

	for _, impl := range impls {
		store := impl.create()
		key := "test_key"
		value := "test_value"

		b.ResetTimer()
		b.Run(impl.name, func(b *testing.B) {
			ctx := context.Background()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {

					for i := 0; i < b.N; i++ {
						if err := store.Set(ctx, fmt.Sprintf("%s:%s:%d", impl.name, key, i), value); err != nil {
							b.Fatalf("failed to set value: %v", err)
						}
					}
				}
			})
		})
	}
}

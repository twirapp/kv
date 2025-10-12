package stores

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/redis/go-redis/v9"
	tcmemcached "github.com/testcontainers/testcontainers-go/modules/memcached"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/twirapp/kv"
	kvinmemory "github.com/twirapp/kv/stores/inmemory"
	kvmemcached "github.com/twirapp/kv/stores/memcached"
	kvotter "github.com/twirapp/kv/stores/otter"
	kvredis "github.com/twirapp/kv/stores/redis"
)

type redisContainer struct {
	container *tcredis.RedisContainer
	opts      *redis.Options
}

type memcachedContainer struct {
	container *tcmemcached.Container
	opts      *redis.Options
}

var (
	redisCreateLock sync.Mutex
	redisContainers []redisContainer

	memcachedCreateLock sync.Mutex
	memcachedContainers []memcachedContainer

	implementations = []struct {
		name   string
		create func() kv.KV
	}{
		{
			name: "InMemory",
			create: func() kv.KV {
				return kvinmemory.New()
			},
		},
		{
			name: "Otter",
			create: func() kv.KV {
				return kvotter.New()
			},
		},
		{
			name: "Redis",
			create: func() kv.KV {
				ctx := context.Background()
				rc, err := tcredis.Run(ctx, "redis:7")
				if err != nil {
					fmt.Printf("Could not start redis container: %v\n", err)
					os.Exit(1)
				}
				connString, err := rc.ConnectionString(ctx)
				if err != nil {
					fmt.Printf("Could not get redis connection string: %v\n", err)
					os.Exit(1)
				}
				rOpts, err := redis.ParseURL(connString)
				if err != nil {
					fmt.Printf("Could not parse redis connection string: %v\n", err)
					os.Exit(1)
				}

				redisCreateLock.Lock()
				redisContainers = append(redisContainers, redisContainer{
					container: rc,
					opts:      rOpts,
				})
				redisCreateLock.Unlock()

				return kvredis.New(redis.NewClient(rOpts))
			},
		},
		{
			name: "Memcached",
			create: func() kv.KV {
				ctx := context.Background()
				mc, err := tcmemcached.Run(ctx, "memcached:1.6")
				if err != nil {
					fmt.Printf("Could not start memcached container: %v\n", err)
					os.Exit(1)
				}
				endpoint, err := mc.HostPort(ctx)
				if err != nil {
					fmt.Printf("Could not get memcached endpoint: %v\n", err)
					os.Exit(1)
				}

				memcachedCreateLock.Lock()
				memcachedContainers = append(memcachedContainers, memcachedContainer{
					container: mc,
				})
				memcachedCreateLock.Unlock()

				return kvmemcached.New(memcache.New(endpoint))
			},
		},
	}
)

func TestMain(m *testing.M) {
	// Run all the tests in the package
	exitCode := m.Run()

	var redisCleanUpWg sync.WaitGroup
	for _, c := range redisContainers {
		redisCleanUpWg.Add(1)
		go func() {
			defer redisCleanUpWg.Done()
			if err := c.container.Terminate(context.TODO()); err != nil {
				fmt.Printf("Could not terminate redis container: %v\n", err)
			}
		}()
	}
	redisCleanUpWg.Wait()

	var memcachedCleanUpWg sync.WaitGroup
	for _, c := range memcachedContainers {
		memcachedCleanUpWg.Add(1)
		go func() {
			defer memcachedCleanUpWg.Done()
			if err := c.container.Terminate(context.TODO()); err != nil {
				fmt.Printf("Could not terminate memcached container: %v\n", err)
			}
		}()
	}
	memcachedCleanUpWg.Wait()

	os.Exit(exitCode)
}

func TestStore_Delete(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		before  func(kv.KV, args, *testing.T)
		after   func(kv.KV, args, *testing.T)
	}{
		{
			name: "key exists and deleted",
			args: args{
				ctx: context.Background(),
				key: "key1",
			},
			wantErr: false,
			before: func(memory kv.KV, a args, t *testing.T) {
				err := memory.Set(context.Background(), a.key, "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
			},
			after: func(c kv.KV, a args, t *testing.T) {
				if err := c.Get(context.TODO(), "key1").Err(); err == nil {
					t.Errorf("key 'key1' was not deleted")
				}
			},
		},
		{
			name: "key does not exist",
			args: args{
				ctx: context.Background(),
				key: "nonexistent",
			},
			wantErr: false,
		},
	}

	for _, impl := range implementations {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s: %s", impl.name, tt.name), func(t *testing.T) {
				c := impl.create()

				if tt.before != nil {
					tt.before(c, tt.args, t)
				}

				if err := c.Delete(tt.args.ctx, tt.args.key); (err != nil) != tt.wantErr {
					t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				}

				if tt.after != nil {
					tt.after(c, tt.args, t)
				}
			})
		}
	}

}

func TestStore_DeleteMany(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		before  func(kv.KV, args, *testing.T)
		after   func(kv.KV, args, *testing.T)
	}{
		{
			name: "multiple keys deleted",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "key2", "key3"},
			},
			wantErr: false,
			before: func(memory kv.KV, a args, t *testing.T) {
				for _, key := range a.keys {
					err := memory.Set(context.Background(), key, "")
					if err != nil {
						t.Fatalf("failed to set up test: %v", err)
					}
				}
			},
			after: func(c kv.KV, a args, t *testing.T) {
				for _, key := range a.keys {
					if err := c.Get(context.TODO(), key).Err(); err == nil {
						t.Errorf("key '%s' was not deleted", key)
					}
				}
			},
		},
		{
			name: "some keys do not exist",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "key3"},
			},
			wantErr: false,
			before: func(memory kv.KV, a args, t *testing.T) {
				err := memory.Set(context.Background(), "key1", "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
				err = memory.Set(context.Background(), "key3", "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
			},
			after: func(c kv.KV, a args, t *testing.T) {
				if err := c.Get(context.TODO(), "key1").Err(); err == nil {
					t.Errorf("key 'key1' should not exist")
				}

				if err := c.Get(context.TODO(), "key3").Err(); err == nil {
					t.Errorf("key 'key3' should not exist")
				}
			},
		},
	}

	for _, impl := range implementations {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s: %s", tt.name, tt.name), func(t *testing.T) {
				c := impl.create()

				if tt.before != nil {
					tt.before(c, tt.args, t)
				}

				if err := c.DeleteMany(tt.args.ctx, tt.args.keys); (err != nil) != tt.wantErr {
					t.Errorf("%s: DeleteMany() error = %v, wantErr %v", impl.name, err, tt.wantErr)
				}

				if tt.after != nil {
					tt.after(c, tt.args, t)
				}
			})
		}
	}

}

func TestStore_Exists(t *testing.T) {
	t.Parallel()

	for _, impl := range implementations {
		t.Run(fmt.Sprintf("%s: Exists", impl.name), func(t *testing.T) {
			c := impl.create()

			if err := c.Set(context.Background(), "key1", "value1"); err != nil {
				t.Fatalf("failed to set up test: %v", err)
			}

			exists, err := c.Exists(context.Background(), "key1")
			if err != nil {
				t.Errorf("Exists() error = %v, wantErr %v", err, false)
			} else if !exists {
				t.Errorf("Exists() got = %v, want %v", exists, true)
			}

			exists, err = c.Exists(context.Background(), "nonexistent")
			if err != nil {
				t.Errorf("Exists() error = %v, wantErr %v", err, false)
			} else if exists {
				t.Errorf("Exists() got = %v, want %v", exists, false)
			}
		})
	}
}

func TestStore_ExistsMany(t *testing.T) {
	t.Parallel()

	for _, impl := range implementations {
		t.Run(fmt.Sprintf("%s: ExistsMany", impl.name), func(t *testing.T) {
			c := impl.create()

			keysToSet := []string{"key1", "key2", "key3"}
			for _, key := range keysToSet {
				if err := c.Set(context.Background(), key, "value"); err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
			}

			keysToCheck := []string{"key1", "key2", "nonexistent", "key3", "anothernonexistent"}
			expected := []bool{true, true, false, true, false}

			results, err := c.ExistsMany(context.Background(), keysToCheck)
			if err != nil {
				t.Errorf("ExistsMany() error = %v, wantErr %v", err, false)
				return
			}

			if !reflect.DeepEqual(results, expected) {
				t.Errorf("ExistsMany() got = %v, want %v", results, expected)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	t.Parallel()

	for _, impl := range implementations {
		t.Run(fmt.Sprintf("%s: Get", impl.name), func(t *testing.T) {
			c := impl.create()

			if err := c.Set(context.Background(), "key1", "value1"); err != nil {
				t.Fatalf("failed to set up test: %v", err)
			}

			val := c.Get(context.Background(), "key1")
			if val.Err() != nil {
				t.Errorf("Get() error = %v, wantErr %v", val.Err(), false)
			} else {
				str, err := val.String()
				if err != nil {
					t.Errorf("Get() String() error = %v, wantErr %v", err, false)
				} else if str != "value1" {
					t.Errorf("Get() got = %v, want %v", str, "value1")
				}
			}

			err := c.Get(context.Background(), "nonexistent").Err()
			if err == nil {
				t.Errorf("Get() error = %v, wantErr %v", err, true)
			}

		})
	}

}

func TestStore_GetKeysByPattern(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx     context.Context
		pattern string
	}

	tests := []struct {
		name    string
		args    args
		want    []string
		forSet  []string
		wantErr bool
	}{
		{
			name: "simple pattern match",
			args: args{
				ctx:     context.Background(),
				pattern: "user:*",
			},
			forSet: []string{
				"user:1",
				"user:2",
				"admin:1",
				"user:profile:1",
			},
			want:    []string{"user:1", "user:2", "user:profile:1"},
			wantErr: false,
		},
		{
			name: "no matches",
			args: args{
				ctx:     context.Background(),
				pattern: "guest:*",
			},
			forSet: []string{
				"user:1",
				"user:2",
				"admin:1",
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "complex pattern match",
			args: args{
				ctx:     context.Background(),
				pattern: "user:*:profile",
			},
			forSet: []string{
				"user:1:profile",
				"user:2:profile",
				"user:3:data",
				"admin:1:profile",
			},
			want:    []string{"user:1:profile", "user:2:profile"},
			wantErr: false,
		},
		{
			name: "exact match without wildcard",
			args: args{
				ctx:     context.Background(),
				pattern: "user:1",
			},
			forSet: []string{
				"user:1",
				"user:2",
				"admin:1",
			},
			want:    []string{"user:1"},
			wantErr: false,
		},
		{
			name: "pattern with multiple wildcards",
			args: args{
				ctx:     context.Background(),
				pattern: "*:*",
			},
			forSet: []string{
				"user:1",
				"admin:1",
				"guest:1",
			},
			want:    []string{"user:1", "admin:1", "guest:1"},
			wantErr: false,
		},
		{
			name: "pattern longer than keys",
			args: args{
				ctx:     context.Background(),
				pattern: "user:*:profile",
			},
			forSet: []string{
				"user:1",
				"user:2",
				"admin:1",
			},
			want:    []string{},
			wantErr: false,
		},
	}

	for _, impl := range implementations {
		for _, tt := range tests {
			if impl.name == "Memcached" {
				// Skip Memcached as it does not support GetKeysByPattern
				continue
			}

			t.Run(fmt.Sprintf("%s: %s", impl.name, tt.name), func(t *testing.T) {
				c := impl.create()

				k := make([]kv.SetMany, len(tt.forSet))
				for i, key := range tt.forSet {
					k[i] = kv.SetMany{
						Key:   key,
						Value: "value",
					}
				}

				if err := c.SetMany(context.Background(), k); err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}

				got, err := c.GetKeysByPattern(tt.args.ctx, tt.args.pattern)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetKeysByPattern() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				for i := range got {
					found := false
					for j := range tt.want {
						if got[i] == tt.want[j] {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("GetKeysByPattern() got unexpected key = %v", got[i])
					}
				}

				if len(got) != len(tt.want) {
					t.Errorf("GetKeysByPattern() got = %v, want %v", got, tt.want)
				}
			})
		}
	}

}

func TestStore_Set(t *testing.T) {
	t.Parallel()

	for _, impl := range implementations {
		t.Run(fmt.Sprintf("%s: Set", impl.name), func(t *testing.T) {
			c := impl.create()

			if err := c.Set(context.Background(), "key1", "value1"); err != nil {
				t.Errorf("Set() error = %v, wantErr %v", err, false)
			}

			val := c.Get(context.Background(), "key1")
			if val.Err() != nil {
				t.Errorf("Get() after Set() error = %v, wantErr %v", val.Err(), false)
			} else {
				str, err := val.String()
				if err != nil {
					t.Errorf("Get() String() after Set() error = %v, wantErr %v", err, false)
				} else if str != "value1" {
					t.Errorf("Get() after Set() got = %v, want %v", str, "value1")
				}
			}

		})
	}

}

func TestStore_SetMany(t *testing.T) {
	t.Parallel()

	for _, impl := range implementations {
		t.Run(fmt.Sprintf("%s: SetMany", impl.name), func(t *testing.T) {
			c := impl.create()

			items := []kv.SetMany{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
				{Key: "key3", Value: "value3"},
			}

			if err := c.SetMany(context.Background(), items); err != nil {
				t.Errorf("SetMany() error = %v, wantErr %v", err, false)
			}

			for _, item := range items {
				val := c.Get(context.Background(), item.Key)
				if val.Err() != nil {
					t.Errorf("Get() after SetMany() error = %v, wantErr %v", val.Err(), false)
				} else {
					str, err := val.String()
					if err != nil {
						t.Errorf("Get() String() after SetMany() error = %v, wantErr %v", err, false)
					} else if str != item.Value {
						t.Errorf("Get() after SetMany() got = %v, want %v", str, item.Value)
					}
				}
			}
		})
	}
}

package kvinmemory

import (
	"context"
	"reflect"
	"testing"

	"github.com/twirapp/kv"
)

func TestInMemory_Delete(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		before  func(*InMemory, args, *testing.T)
		after   func(*InMemory, args, *testing.T)
	}{
		{
			name: "key exists and deleted",
			args: args{
				ctx: context.Background(),
				key: "key1",
			},
			wantErr: false,
			before: func(memory *InMemory, a args, t *testing.T) {
				err := memory.Set(context.Background(), a.key, "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
			},
			after: func(c *InMemory, a args, t *testing.T) {
				if _, exists := c.storage["key1"]; exists {
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
			wantErr: true,
			before:  nil,
			after: func(c *InMemory, a args, t *testing.T) {
				if len(c.storage) != 0 {
					t.Errorf("storage should still be empty")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()

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

func TestInMemory_DeleteMany(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		before  func(*InMemory, args, *testing.T)
		after   func(*InMemory, args, *testing.T)
	}{
		{
			name: "multiple keys deleted",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "key2", "key3"},
			},
			wantErr: false,
			before: func(memory *InMemory, a args, t *testing.T) {
				for _, key := range a.keys {
					err := memory.Set(context.Background(), key, "")
					if err != nil {
						t.Fatalf("failed to set up test: %v", err)
					}
				}
			},
			after: func(c *InMemory, a args, t *testing.T) {
				for _, key := range a.keys {
					if _, exists := c.storage[key]; exists {
						t.Errorf("key '%s' was not deleted", key)
					}
				}
			},
		},
		{
			name: "some keys do not exist",
			args: args{
				ctx:  context.Background(),
				keys: []string{"key1", "nonexistent", "key3"},
			},
			wantErr: false,
			before: func(memory *InMemory, a args, t *testing.T) {
				err := memory.Set(context.Background(), "key1", "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
				err = memory.Set(context.Background(), "key3", "")
				if err != nil {
					t.Fatalf("failed to set up test: %v", err)
				}
			},
			after: func(c *InMemory, a args, t *testing.T) {
				if _, exists := c.storage["key1"]; exists {
					t.Errorf("key 'key1' was not deleted")
				}
				if _, exists := c.storage["key3"]; exists {
					t.Errorf("key 'key3' was not deleted")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()

			if tt.before != nil {
				tt.before(c, tt.args, t)
			}

			if err := c.DeleteMany(tt.args.ctx, tt.args.keys); (err != nil) != tt.wantErr {
				t.Errorf("DeleteMany() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.after != nil {
				tt.after(c, tt.args, t)
			}
		})
	}
}

func TestInMemory_Exists(t *testing.T) {
	t.Parallel()

	c := New()

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
}

func TestInMemory_ExistsMany(t *testing.T) {
	t.Parallel()

	c := New()

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
}

func TestInMemory_Get(t *testing.T) {
	t.Parallel()

	c := New()

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
}

func TestInMemory_GetKeysByPattern(t *testing.T) {
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
			want:    []string{"user:1", "user:2"},
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()

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

func TestInMemory_Set(t *testing.T) {
	t.Parallel()

	c := New()

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
}

func TestInMemory_SetMany(t *testing.T) {
	t.Parallel()

	c := New()

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
}

# KV

A key-value store implementation in Go.

## Overview

KV is a lightweight key-value storage solution built with Go, designed for simple and efficient data persistence.

# Backends

- InMemory
- Otter
- Redis
- Memcached
- Valkey
- Valkey glide

## Installation

```bash
go get github.com/twirapp/kv
```

# Usage

```go
package main

import (
	"github.com/go-redis/redis/v9"
	"github.com/twirapp/kv"
	kvredis "github.com/twirapp/kv/stores/redis"
	"context"
	"fmt"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

	kvStorage := kvredis.New(redisClient)
	doSomeWork(kvStorage)
}

func doSomeWork(storage kv.KV) {
	err := storage.Set(context.TODO(), "somekey", 12345)
    if err != nil {
        // handle error
    }

	data, err := storage.Get(context.TODO(), "somekey").Int64()
	if err != nil {
		// handle error
	}

	fmt.Println(data)
}

```

# Benchmarks

### Get

```
cpu: AMD Ryzen 5 5600 6-Core Processor
BenchmarkGet/InMemory-12 30565016 35.76 ns/op
BenchmarkGet/Otter-12 49937456 23.78 ns/op
BenchmarkGet/Redis-12 93295 15257 ns/op
BenchmarkGet/Memcached-12 92832 13554 ns/op
BenchmarkGet/Valkey_Glide-12 131566 8749 ns/op
BenchmarkGet/Valkey-12 104168 11323 ns/op
```

### Set

```
cpu: AMD Ryzen 5 5600 6-Core Processor
BenchmarkSet/InMemory-12                 5751238               208.3 ns/op
BenchmarkSet/Otter-12                    4138132               290.2 ns/op
BenchmarkSet/Redis-12                      78825             14400 ns/op
BenchmarkSet/Memcached-12                  73033             15546 ns/op
BenchmarkSet/Valkey_Glide-12              125948              9242 ns/op
BenchmarkSet/Valkey-12                    119925             10249 ns/op
```

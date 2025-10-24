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
BenchmarkGet/InMemory-12                32833582                36.89 ns/op
BenchmarkGet/Otter-12                   43721808                25.31 ns/op
BenchmarkGet/Redis-12                      90266             13027 ns/op
BenchmarkGet/Memcached-12                  99343             13220 ns/op
BenchmarkGet/Valkey_Glide-12              143864              8541 ns/op
BenchmarkGet/Valkey-12                    117400             10261 ns/op
BenchmarkGet/Redis-Glide-12               129991              9679 ns/op
BenchmarkGet/Dragonfly_(redis_client)-12                   55845             19551 ns/op
```

### Set

```
cpu: AMD Ryzen 5 5600 6-Core Processor
BenchmarkSet/InMemory-12                 5903462               203.9 ns/op
BenchmarkSet/Otter-12                    4391866               281.6 ns/op
BenchmarkSet/Redis-12                      80160             13843 ns/op
BenchmarkSet/Memcached-12                  78168             15544 ns/op
BenchmarkSet/Valkey_Glide-12              130809              9455 ns/op
BenchmarkSet/Valkey-12                    118170             10083 ns/op
BenchmarkSet/Redis-Glide-12               113458              9772 ns/op
BenchmarkSet/Dragonfly_(redis_client)-12                   73546             16483 ns/op
```

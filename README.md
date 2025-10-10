# KV

A key-value store implementation in Go.

## Overview

KV is a lightweight key-value storage solution built with Go, designed for simple and efficient data persistence.

Is can use various backends, but for now only redis and inmemory implemented

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
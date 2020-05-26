# Distribute Lock In Redis

<img align="right" width="159px" src="https://raw.githubusercontent.com/gin-gonic/logo/master/color.png">

RedisLock is a distribute lock tool written in Go (Golang). If you need performance and good productivity, you will love RedisLock.


## Installation

To install RedisLock package, you need to install Go and set your Go workspace first.

1. The first need [Go](https://golang.org/) installed , then you can use the below Go command to install Gin.

```sh
$ go get -u github.com/shileislslsl/redislock 
```

2. Import it in your code:

```go
import "github.com/shileislslsl/redislock "
```

3. (Optional) Import `github.com/go-redis/redis`. This is required to connect redis.

```go
import "github.com/go-redis/redis"
```

```go
package main

import (
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/shileislslsl/redislock"
)

func main() {
	cli = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	rlock := redislock.NewClient(cli)
	//lock take three params,lock key,default expiration time,auto refresh key expiration time before unlock the key
	_, err := rlock.Lock("test_key2", 10*time.Second, true)
	if err != nil {
		log.Error(err)
	}
	_, err = rlock.Unlock()
	if err != nil {
		log.Error(err)
	}
}
```


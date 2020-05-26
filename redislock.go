package redislock

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/go-redis/redis"
)

const (
	//SUC relaese lock success
	SUC = iota
	//ERR relaese lock failed
	ERR
	luaScript = `local val = redis.call("get", KEYS[1]) 
	if val == false then return -1 
	end 
	if val ~= ARGV[1] then 
		return 0
	else 
		redis.call("del", KEYS[1]) 
		return 1 
	end`
)

var (
	//NILClient redis client init error or close
	NILClient = errors.New("nil client")
	//LOCKFailed get lock failed
	LOCKFailed = errors.New("set key failed")
	//UNLOCKVALUEERROR unlock value not match set value
	UNLOCKVALUEERROR = errors.New("unlock value error")
	//UNLOCKKEYNOTFOUND unlock key not found
	UNLOCKKEYNOTFOUND = errors.New("unlock key not found error")
)

//RedisLock base struct
type RedisLock struct {
	unLock chan int64
	res    chan error
	cli    *redis.Client
	ctx    context.Context
	key    string
	val    int64
}

//NewRedisLock init a new RedisLock struct
func NewRedisLock(cli *redis.Client) *RedisLock {
	if cli == nil {
		panic(NILClient)
	}
	rlock := RedisLock{}
	rlock.cli = cli
	rlock.unLock = make(chan int64, 0)
	rlock.res = make(chan error, 0)
	return &rlock
}

//SetContext set context,so can use cancle function to canle lock
func (r *RedisLock) SetContext(ctx context.Context) {
	r.ctx = ctx
}

//Lock set redis lock
func (r *RedisLock) Lock(key string, timeOut time.Duration, renew bool) error {
	rand.Seed(time.Now().UnixNano())
	value := rand.Int63()
	res := r.cli.SetNX(key, value, timeOut).Val()
	if !res {
		return LOCKFailed
	}
	r.key = key
	r.val = value
	if renew {
		if r.ctx == nil {
			go r.autoRenew(timeOut)
		} else {
			go r.autoRenewWitchCtx(timeOut)
		}
	}
	return nil
}

func (r *RedisLock) autoRenewWitchCtx(timeout time.Duration) {
	ticker := time.NewTicker(timeout / 2)
	defer ticker.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			r.cli.Expire(r.key, timeout)
		case <-r.unLock:
			r.checkValAndDel()
			break loop
		case <-r.ctx.Done():
			go r.checkValAndDel()
			break loop
		}
	}
}

func (r *RedisLock) autoRenew(timeout time.Duration) {
	ticker := time.NewTicker(timeout / 2)
	defer ticker.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			r.cli.Expire(r.key, timeout)
		case <-r.unLock:
			r.checkValAndDel()
			break loop
		}
	}
}
func (r *RedisLock) checkValAndDel() {
	v, err := r.cli.Eval(luaScript, []string{r.key}, r.val).Result()
	if err != nil {
		r.res <- NILClient
	}
	intv := v.(int64)
	switch intv {
	case 1:
		r.res <- nil
	case 0:
		r.res <- UNLOCKVALUEERROR
	case -1:
		r.res <- UNLOCKKEYNOTFOUND
	}
	return
}

//UnLock releaselock
func (r *RedisLock) UnLock() error {
	r.unLock <- r.val
	return <-r.res

}

/*TTL get lock remain time
-1: no expiration time
-2: key not exists
*/
func (r *RedisLock) TTL() time.Duration {
	return r.cli.TTL(r.key).Val()
}

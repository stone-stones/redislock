package redislock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

var tCli *redis.Client

func initClient() {
	tCli = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	s := tCli.Ping()
	if err := s.Err(); err != nil {
		panic(err)
	}
	return
}

func TestMain(m *testing.M) {
	initClient()
	m.Run()
}

func TestLockSetError(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key2", 10*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found")
	}
	_, err = rlock.Lock("test_key2", 1*time.Second, false)
	if err != LOCKFailed {
		t.Errorf("mute set success")
	}
}

func TestKeyTimeOut(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key3", 200*time.Millisecond, false)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	time.Sleep(210 * time.Millisecond)
	if tCli.Exists("test_key3").Val() == 1 {
		t.Error("key exists")
	}
}
func TestKeyNotTimeOut(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key4", 1*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if tCli.Exists("test_key4").Val() == 0 {
		t.Error("key  not exists")
	}
}

func TestUnlock(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key5", 2*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	err = rlock.UnLock()
	if err != nil {
		t.Error("unlock key error", err)
	}
}

func TestUnlockValueChange(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key6", 2*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	tCli.Set("test_key6", 1111, 10*time.Second)
	err = rlock.UnLock()
	if err != UNLOCKVALUEERROR {
		t.Error("unlock key error", err)
	}
}

func TestMuteLockKey(t *testing.T) {
	for i := 0; i < 20; i++ {
		s := i
		go LockAndUnlock(s, t)
	}
	time.Sleep(3 * time.Second)

}

func LockAndUnlock(i int, t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock(fmt.Sprintf("test_key%d", i+10), 2*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	time.Sleep(2100 * time.Millisecond)
	err = rlock.UnLock()
	if err != nil {
		t.Error("unlock key error", err)
	}
}

func TestCtxLock(t *testing.T) {
	backCtx := context.Background()
	ctx, cancel := context.WithCancel(backCtx)
	rlock := NewRedisLock(tCli)
	rlock.SetContext(ctx)
	_, err := rlock.Lock("test_key7", 2*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	cancel()
	if tCli.Exists("test_key3").Val() == 1 {
		t.Error("key exists")
	}

}

func TestTTl(t *testing.T) {
	rlock := NewRedisLock(tCli)
	_, err := rlock.Lock("test_key8", 2*time.Second, true)
	if err != nil {
		t.Error("set lock params error not found", err)
	}
	err = rlock.UnLock()
	if err != nil {
		t.Error("unlock key error", err)
	}
	if rlock.TTL() != -2*time.Second {
		t.Errorf("ttl result error")
	}
}

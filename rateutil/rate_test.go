package rateutil

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRate(t *testing.T) {
	ins, err := RateInstanceFactory(TokenBucket, "test1", 1000, 100)
	fmt.Println(ins)
	fmt.Println(err)
}

func TestTryAllow(t *testing.T) {
	var a int64
	var b int64
	var wg sync.WaitGroup
	for i := 0; i < 100000; i++ {
		time.Sleep(100 * time.Microsecond)
		wg.Add(1)
		go func() {
			ins, err := RateInstanceFactory(TokenBucket, "test1", 1000, 100)
			if err != nil {
				fmt.Println("err:", err)
				return
			}
			ok := ins.TryAllow()
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&b, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println(a)
	fmt.Println(b)
}

func TestAllowTimeOut(t *testing.T) {
	var a int64
	var b int64
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		time.Sleep(100)
		wg.Add(1)
		go func(ind int) {
			ins, err := RateInstanceFactory(TokenBucket, "test1", 1000, 100)
			if err != nil {
				fmt.Println("err:", err)
				return
			}
			ok := ins.Allow(time.Millisecond*10, false)
			//ok := ins.Allow(time.Duration(0), false)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&b, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println(a)
	fmt.Println(b)
}

func TestAllowLeaky(t *testing.T) {
	var a int64
	var b int64
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(ind int) {
			ins, err := RateInstanceFactory(LeakyBucket, "test1", 1000)
			if err != nil {
				fmt.Println("err:", err)
				return
			}
			ok := ins.Allow(time.Millisecond*10, true)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&b, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println(a)
	fmt.Println(b)
}

func TestRedisNoWait(t *testing.T) {
	var a int64
	var b int64
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func(ind int) {
			ins, err := RateInstanceFactory(RedisTokenBucketNoWait, "cache", "redis_rate_key", 1000, 10)
			if err != nil {
				fmt.Println("err:", err)
				return
			}
			ok := ins.Allow(time.Second*5, false)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&b, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println(a)
	fmt.Println(b)
}

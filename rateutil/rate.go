package rateutil

import (
	"sync"
	"time"
)

type RateType int

const (
	TokenBucket RateType = iota + 1
	LeakyBucket
	SlidingWindow
	RedisTokenBucket
	RedisLeakyBucket
	RedisTokenBucketNoWait
)

type RateInstanceMap struct {
	lock    *sync.Mutex
	rateMap map[string]RateInstance
}

type RateInstance interface {
	TryAllow() bool
	TryAllowN(n int) bool
	Allow(d time.Duration, withSleep bool) bool
	AllowN(d time.Duration, withSleep bool, n int) bool
	Last() time.Time
}

var rMap RateInstanceMap

func init() {
	rMap.lock = new(sync.Mutex)
	rMap.rateMap = make(map[string]RateInstance, 0)
	go clearRateMap()
}

func RateInstanceFactory(rType RateType, key string, args ...interface{}) (RateInstance, error) {
	var prefix string
	var fn func(args []interface{}) (RateInstance, error)
	var instance RateInstance
	var err error
	switch rType {
	case TokenBucket:
		prefix = "[TokenBucket]"
		fn = newTokenBucket
	case LeakyBucket:
		prefix = "[LeakyBucket]"
		fn = newLeakyBucket
	case SlidingWindow:
		prefix = "[SlidingWindow]"
		fn = newSlidingWindow
	case RedisTokenBucket:
		prefix = "[RedisTokenBucket]"
		fn = func(args []interface{}) (RateInstance, error) {
			return newRedisTokenBucket(args)
		}
	case RedisLeakyBucket:
		prefix = "[RedisLeakyBucket]"
		fn = func(args []interface{}) (RateInstance, error) {
			return newRedisLeakyBucket(args)
		}
	case RedisTokenBucketNoWait:
		prefix = "[RedisTokenBucketNoWait]"
		fn = func(args []interface{}) (RateInstance, error) {
			return newRedisTokenBucketWithoutWait(args)
		}
	}
	rMap.lock.Lock()
	defer rMap.lock.Unlock()
	if instance, ok := rMap.rateMap[prefix+key]; ok {
		return instance, nil
	}
	instance, err = fn(args)
	if err != nil {
		return nil, err
	}
	rMap.rateMap[prefix+key] = instance
	return instance, nil
}

func clearRateMap() {
	timer := time.NewTicker(time.Minute * 5)
	for {
		select {
		case <-timer.C:
			rMap.lock.Lock()
			for k, ins := range rMap.rateMap {
				if time.Now().Sub(ins.Last()) > time.Minute*10 {
					delete(rMap.rateMap, k)
				}
			}
			rMap.lock.Unlock()
		}
	}
}

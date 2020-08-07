package rateutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cast"
	redisdao "github.com/tal-tech/xredis"
)

var initScript = `
                    local exist=redis.call("EXISTS",KEYS[1])
                    if exist == 0 then
                        redis.call("HSET",KEYS[1],"rate",ARGV[1])        
                        redis.call("HSET",KEYS[1],"burst",ARGV[2])        
                        redis.call("HSET",KEYS[1],"curr_permits",ARGV[3])        
	                end
					return 1
                 `

var reversescript = `  
                    local ratelimit_info=redis.pcall("HMGET",KEYS[1],"last_mill_second","curr_permits","burst","rate")
					local last_mill_second=ratelimit_info[1]
					local curr_permits=tonumber(ratelimit_info[2])
					local max_burst=tonumber(ratelimit_info[3])
					local rate=tonumber(ratelimit_info[4])

                    local local_curr_permits=max_burst
                     
					if(type(last_mill_second) ~='boolean' and last_mill_second ~=nil) then
					    local reverse_permits=math.floor((ARGV[2]-last_mill_second)*rate/1000)
						if(reverse_permits<0) then
						    reverse_permits=0
						end
						if(reverse_permits>0) then
						    redis.pcall("HMSET",KEYS[1],"last_mill_second",ARGV[2])
                        end
						local expect_curr_permits=reverse_permits+curr_permits
						local_curr_permits=math.min(expect_curr_permits,max_burst);
					else
						redis.pcall("HMSET",KEYS[1],"last_mill_second",ARGV[2])
					end

					local result=-1
					if(local_curr_permits-ARGV[1]>=0) then
					    result=0
						redis.pcall("HMSET",KEYS[1],"curr_permits",local_curr_permits-ARGV[1])
					else
					    local expect_wait_second=math.floor((ARGV[1]-local_curr_permits)*1000/rate)
						if(ARGV[3]-expect_wait_second>=0) then
						    result=expect_wait_second
						    redis.pcall("HMSET",KEYS[1],"curr_permits",local_curr_permits-ARGV[1])
                        else
						    redis.pcall("HMSET",KEYS[1],"curr_permits",local_curr_permits)
						end
					end
					return result 
                `

type RedisBucket struct {
	last        time.Time
	instance    string
	key         string
	burst       int
	limit       int
	withoutWait bool
}

func newRedisTokenBucket(args []interface{}) (*RedisBucket, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("args length less than 4")
	}
	instance, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("instance need a string")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("key need a string")
	}
	limit, ok := args[2].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	burst, ok := args[3].(int)
	if !ok {
		return nil, fmt.Errorf("burst need a int")
	}
	this := new(RedisBucket)
	this.instance = instance
	this.burst = burst
	this.limit = limit
	this.key = key + "_" + cast.ToString(limit) + "_" + cast.ToString(burst)
	err := this.init()
	return this, err
}

func newRedisTokenBucketWithoutWait(args []interface{}) (*RedisBucket, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("args length less than 4")
	}
	instance, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("instance need a string")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("key need a string")
	}
	limit, ok := args[2].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	burst, ok := args[3].(int)
	if !ok {
		return nil, fmt.Errorf("burst need a int")
	}
	this := new(RedisBucket)
	this.instance = instance
	this.burst = burst
	this.limit = limit
	this.key = key + "_" + cast.ToString(limit) + "_" + cast.ToString(burst)
	this.withoutWait = true
	err := this.init()
	return this, err
}

func newRedisLeakyBucket(args []interface{}) (*RedisBucket, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("args length less than 3")
	}
	instance, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("instance need a string")
	}
	key, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("key need a string")
	}
	limit, ok := args[2].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	this := new(RedisBucket)
	this.instance = instance
	this.burst = 1
	this.limit = limit
	this.key = key + "_" + cast.ToString(limit) + "_1"
	err := this.init()
	return this, err
}

func (this *RedisBucket) init() error {
	_, err := redisdao.NewXesRedis().Eval(this.instance, initScript, []string{this.key}, this.limit, this.burst, 0)
	return err
}

func (this *RedisBucket) Last() time.Time {
	return this.last
}

func (this *RedisBucket) TryAllow() bool {
	return this.TryAllowN(1)
}

func (this *RedisBucket) TryAllowN(n int) bool {
	this.last = time.Now()
	_, err := this.reserveN(time.Now(), n, 0)
	return err == nil
}

func (this *RedisBucket) Allow(d time.Duration, withSleep bool) bool {
	return this.AllowN(d, withSleep, 1)
}

func (this *RedisBucket) AllowN(d time.Duration, withSleep bool, n int) bool {
	this.last = time.Now()
	ctx := context.Background()
	if int64(d) > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}
	err := this.waitN(ctx, n)
	if err != nil && err.Error() != "context deadline exceeded" && withSleep {
		time.Sleep(d)
	}
	return err == nil
}

func (this *RedisBucket) waitN(ctx context.Context, n int) (err error) {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	now := time.Now()
	waitLimit := InfDuration
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(now)
	}

	wait, err := this.reserveN(now, n, waitLimit)
	if err != nil {
		return err
	}
	t := time.NewTimer(wait)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (this *RedisBucket) reserveN(now time.Time, n int, t time.Duration) (time.Duration, error) {
	var wait time.Duration

	nowMilliSecond := now.UnixNano() / 1e6
	waitMilliSecond := int(t / time.Millisecond)

	res, err := redisdao.NewXesRedis().Eval(this.instance, reversescript, []string{this.key}, n, nowMilliSecond, waitMilliSecond)

	if err != nil {
		return wait, err
	}

	if cast.ToInt(res) < 0 {
		return wait, errors.New("get token fail")
	}

	if this.withoutWait {
		return time.Duration(0), nil
	}

	wait = time.Millisecond * time.Duration(cast.ToInt(res))

	return wait, nil
}

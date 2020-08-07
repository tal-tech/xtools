package rateutil

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

type Bucket struct {
	limiter *rate.Limiter
	last    time.Time
}

func newTokenBucket(args []interface{}) (RateInstance, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("args length less than 2")
	}
	limit, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	burst, ok := args[1].(int)
	if !ok {
		return nil, fmt.Errorf("burst need a int")
	}
	instance := new(Bucket)
	instance.limiter = rate.NewLimiter(rate.Limit(limit), burst)
	instance.last = time.Now()
	return instance, nil
}

func newLeakyBucket(args []interface{}) (RateInstance, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("args length less than 1")
	}
	limit, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	instance := new(Bucket)
	instance.limiter = rate.NewLimiter(rate.Limit(limit), 1)
	instance.last = time.Now()
	return instance, nil
}

func (this *Bucket) TryAllow() bool {
	return this.TryAllowN(1)
}

func (this *Bucket) TryAllowN(n int) bool {
	this.last = time.Now()
	return this.limiter.AllowN(time.Now(), n)
}

func (this *Bucket) Allow(d time.Duration, withSleep bool) bool {
	return this.AllowN(d, withSleep, 1)
}

func (this *Bucket) AllowN(d time.Duration, withSleep bool, n int) bool {
	this.last = time.Now()
	ctx := context.Background()
	if int64(d) > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}
	err := this.limiter.WaitN(ctx, n)
	if err != nil && err.Error() != "context deadline exceeded" && withSleep {
		time.Sleep(d)
	}
	return err == nil
}

func (this *Bucket) Last() time.Time {
	return this.last
}

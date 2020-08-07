package rateutil

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRedisBucket(t *testing.T) {
	ins, err := newRedisTokenBucket([]interface{}{"cache", "testkeytoken", 100, 50})
	fmt.Println(ins)
	fmt.Println(err)
}

func runRedis(t *testing.T, lim *RedisBucket, allows []allow) {
	for i, allow := range allows {
		ok := lim.TryAllowN(allow.n)
		if ok != allow.ok {
			t.Errorf("step %d: lim.AllowN(%v, %v) = %v want %v",
				i, allow.t, allow.n, ok, allow.ok)
		}
	}
}
func TestRedisBucket1(t *testing.T) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkeytoken", 10, 10})
	runRedis(t, lim, []allow{
		{t1, 1, true},
		{t1, 20, false},
		{t2, 2, true},
		{t2, 5, true},
		{t3, 5, false},
		{t3, 2, true},
		{t11, 5, true},
		{t11, 6, false},
		{t11, 1, true},
	})
}

func TestRedisBucket3(t *testing.T) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkeytoken", 10, 5})
	runRedis(t, lim, []allow{
		{t1, 4, true},
		{t1, 2, false},
		{t2, 1, true},
		{t4, 3, true},
		{t4, 1, false},
		{t4, 1, false},
		{t11, 3, true},
		{t11, 4, false},
		{t11, 2, true},
	})
}

func TestRedisSimultaneousRequests(t *testing.T) {
	const (
		key         = "testkey"
		limit       = 10
		burst       = 2
		numRequests = 100
	)
	var (
		wg    sync.WaitGroup
		numOK = uint32(0)
	)

	// Very slow replenishing bucket.
	lim, _ := newRedisTokenBucket([]interface{}{"cache", key, limit, burst})

	// Tries to take a token, atomically updates the counter and decreases the wait
	// group counter.
	f := func() {
		defer wg.Done()
		if ok := lim.TryAllow(); ok {
			atomic.AddUint32(&numOK, 1)
		}
	}

	wg.Add(numRequests)
	for i := 0; i < numRequests; i++ {
		go f()
	}
	wg.Wait()
	if numOK != limit {
		t.Errorf("numOK = %d, want %d", numOK, limit)
	}
}

func TestRedisLongRunningQPS(t *testing.T) {
	const (
		key   = "testkey"
		limit = 100
		burst = 1
	)
	var numOK = int32(0)

	lim, _ := newRedisTokenBucket([]interface{}{"cache", key, limit, burst})

	var wg sync.WaitGroup
	f := func() {
		if ok := lim.TryAllow(); ok {
			atomic.AddInt32(&numOK, 1)
		}
		wg.Done()
	}

	start := time.Now()
	end := start.Add(5 * time.Second)
	for time.Now().Before(end) {
		wg.Add(1)
		go f()

		// This will still offer ~500 requests per second, but won't consume
		// outrageous amount of CPU.
		time.Sleep(2 * time.Millisecond)
	}
	wg.Wait()
	elapsed := time.Since(start)
	ideal := (limit * float64(elapsed) / float64(time.Second))

	// We should never get more requests than allowed.
	if want := int32(ideal + 1); numOK > want {
		t.Errorf("numOK = %d, want %d (ideal %f)", numOK, want, ideal)
	}
	// We should get very close to the number of requests allowed.
	if want := int32(0.999 * ideal); numOK < want {
		t.Errorf("numOK = %d, want %d (ideal %f)", numOK, want, ideal)
	}
}

func runRedisAllowN(t *testing.T, lim *RedisBucket, w wait) {
	ok := lim.AllowN(w.d, w.withsleep, w.n)
	if ok != w.ok {
		t.Errorf("lim.AllowN(%v, %v) = %v want %v",
			w.d, w.n, ok, w.ok)
	}
}

func TestRedisWaitSimple(t *testing.T) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkey", 10, 1})

	runRedisAllowN(t, lim, wait{time.Millisecond * 2000, false, 21, false})
	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 2, true})
	time.Sleep(time.Millisecond * 200)
	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 5, true})
	time.Sleep(time.Millisecond * 200)
	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 4, true})
	runRedisAllowN(t, lim, wait{time.Millisecond * 700, false, 6, true})
	runRedisAllowN(t, lim, wait{time.Millisecond * 300, false, 3, false})

}

func TestRedisWaitSleep(t *testing.T) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkey", 10, 5})

	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 2, true})
	time.Sleep(time.Millisecond * 200)
	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 5, true})
	time.Sleep(time.Millisecond * 200)
	runRedisAllowN(t, lim, wait{time.Millisecond * 1000, false, 4, true})
	time.Sleep(time.Millisecond * 400)
	runRedisAllowN(t, lim, wait{time.Millisecond * 300, true, 10, false})
	runRedisAllowN(t, lim, wait{time.Millisecond * 500, false, 9, true})

}

func TestRedisWlf(t *testing.T) {
	lim, _ := newRedisTokenBucketWithoutWait([]interface{}{"cache", "testkey", 100, 10})
	var a int64
	var c int64
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			//ctx, _ := context.WithTimeout(context.Background(), 0)
			ok := lim.AllowN(time.Second*3, false, 1)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&c, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("a:", a)
	fmt.Println("c:", c)

}

func BenchmarkRedisAllowN(b *testing.B) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkey", 1000, 1000})
	b.ReportAllocs()
	b.ResetTimer()
	var a int64
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ok := lim.AllowN(time.Millisecond, false, 1)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&c, 1)
			}
		}
	})
	fmt.Println("a:", a)
	fmt.Println("c:", c)
}

func BenchmarkRedisTryAllowN(b *testing.B) {
	lim, _ := newRedisTokenBucket([]interface{}{"cache", "testkey", 1000, 10})
	b.ReportAllocs()
	b.ResetTimer()
	var a int64
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ok := lim.TryAllowN(1)
			if ok {
				atomic.AddInt64(&a, 1)
			} else {
				atomic.AddInt64(&c, 1)
			}
		}
	})
	fmt.Println("a:", a)
	fmt.Println("c:", c)
}

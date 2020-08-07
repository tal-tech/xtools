package rateutil

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	d = 100 * time.Millisecond
)

var (
	t0  = time.Now()
	t1  = t0.Add(time.Duration(1) * d)
	t2  = t0.Add(time.Duration(2) * d)
	t3  = t0.Add(time.Duration(3) * d)
	t4  = t0.Add(time.Duration(4) * d)
	t5  = t0.Add(time.Duration(5) * d)
	t9  = t0.Add(time.Duration(9) * d)
	t11 = t0.Add(time.Duration(11) * d)
)

type allow struct {
	t  time.Time
	n  int
	ok bool
}

func TestNewWindow(t *testing.T) {
	ins, err := newSlidingWindow([]interface{}{10, 2})
	fmt.Println(ins)
	fmt.Println(err)
}

func run(t *testing.T, lim *Window, allows []allow) {
	for i, allow := range allows {
		ok := lim.TryAllowN(allow.n)
		if ok != allow.ok {
			t.Errorf("step %d: lim.AllowN(%v, %v) = %v want %v",
				i, allow.t, allow.n, ok, allow.ok)
		}
	}
}
func TestWindow1(t *testing.T) {
	run(t, NewWindow(10, 1), []allow{
		{t1, 1, true},
		{t1, 20, false},
		{t2, 2, true},
		{t2, 5, true},
		{t3, 5, false},
		{t3, 2, true},
		{t11, 5, true},
		{t11, 5, true},
		{t11, 1, false},
	})
}

func TestWindow3(t *testing.T) {
	run(t, NewWindow(10, 5), []allow{
		{t1, 4, true},
		{t2, 1, true},
		{t1, 1, true},
		{t4, 4, true},
		{t4, 1, false},
		{t4, 1, false},
		{t11, 3, true},
		{t11, 4, false},
		{t11, 2, true},
	})
}

func TestSimultaneousRequests(t *testing.T) {
	const (
		limit       = 10
		windows     = 5
		numRequests = 100
	)
	var (
		wg    sync.WaitGroup
		numOK = uint32(0)
	)

	// Very slow replenishing bucket.
	lim := NewWindow(limit, windows)

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

func TestLongRunningQPS(t *testing.T) {
	const (
		limit   = 200
		windows = 1000
	)
	var numOK = int32(0)

	lim := NewWindow(limit, windows)

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

type request struct {
	t   time.Time
	n   int
	act time.Time
	ok  bool
}

type wait struct {
	d         time.Duration
	withsleep bool
	n         int
	ok        bool
}

func runAllowN(t *testing.T, lim *Window, w wait) {
	ok := lim.AllowN(w.d, w.withsleep, w.n)
	if ok != w.ok {
		t.Errorf("lim.AllowN(%v, %v) = %v want %v",
			w.d, w.n, ok, w.ok)
	}
}

func TestWaitSimple(t *testing.T) {
	lim := NewWindow(10, 5)

	runAllowN(t, lim, wait{time.Millisecond * 2000, false, 21, false})
	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 2, true})
	time.Sleep(time.Millisecond * 200)
	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 5, true})
	time.Sleep(time.Millisecond * 200)
	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 4, true})
	runAllowN(t, lim, wait{time.Millisecond * 300, false, 6, true})
	runAllowN(t, lim, wait{time.Millisecond * 300, false, 4, false})

}

func TestWaitSleep(t *testing.T) {
	lim := NewWindow(10, 5)

	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 2, true})
	time.Sleep(time.Millisecond * 200)
	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 5, true})
	time.Sleep(time.Millisecond * 200)
	runAllowN(t, lim, wait{time.Millisecond * 1000, false, 4, true})
	time.Sleep(time.Millisecond * 400)
	runAllowN(t, lim, wait{time.Millisecond * 300, false, 10, false})
	runAllowN(t, lim, wait{time.Millisecond * 500, false, 9, true})

}

func TestWlf(t *testing.T) {
	lim := NewWindow(100, 10)
	var a int64
	var c int64
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			//ctx, _ := context.WithTimeout(context.Background(), 0)
			ok := lim.AllowN(time.Second*3, true, 1)
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

func BenchmarkAllowN(b *testing.B) {
	lim := NewWindow(1000, 10)
	b.ReportAllocs()
	b.ResetTimer()
	var a int64
	var c int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ok := lim.AllowN(time.Millisecond*10, false, 1)
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

func BenchmarkTryAllowN(b *testing.B) {
	lim := NewWindow(1000, 10)
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

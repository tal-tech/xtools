package rateutil

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const SLICEPREBUFFER int = 5

type LimitNode struct {
	nodeEvent time.Time
	max       int
	used      int
}

type Window struct {
	last     time.Time
	windows  int
	limit    int
	datasize int
	duration time.Duration
	mu       sync.Mutex
	data     []LimitNode
	index    int
	preindex int
}

func newSlidingWindow(args []interface{}) (RateInstance, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("args length less than 2")
	}
	limit, ok := args[0].(int)
	if !ok {
		return nil, fmt.Errorf("limit need a int")
	}
	windows, ok := args[1].(int)
	if !ok {
		return nil, fmt.Errorf("windows need a int")
	}
	this := new(Window)
	this.windows = windows
	this.limit = limit
	this.duration = time.Second / time.Duration(windows)
	if time.Second%time.Duration(windows) != 0 {
		return nil, fmt.Errorf("time.Second %% window != 0")
	}
	this.datasize = windows * SLICEPREBUFFER
	this.data = make([]LimitNode, windows*SLICEPREBUFFER)
	this.data[this.index].nodeEvent = time.Now()
	return this, nil
}

func NewWindow(limit, windows int) *Window {
	this := new(Window)
	this.windows = windows
	this.limit = limit
	this.duration = time.Second / time.Duration(windows)
	this.datasize = windows * SLICEPREBUFFER
	this.data = make([]LimitNode, this.datasize)
	this.data[this.index].nodeEvent = time.Now()
	this.data[this.index].max = limit
	return this
}

func (this *Window) Last() time.Time {
	return this.last
}

func (this *Window) TryAllow() bool {
	return this.TryAllowN(1)
}

func (this *Window) TryAllowN(n int) bool {
	this.last = time.Now()
	_, err := this.reserveN(time.Now(), n, 0)
	return err == nil
}

func (this *Window) Allow(d time.Duration, withSleep bool) bool {
	return this.AllowN(d, withSleep, 1)
}

func (this *Window) AllowN(d time.Duration, withSleep bool, n int) bool {
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

const InfDuration = time.Duration(1<<63 - 1)

func (this *Window) waitN(ctx context.Context, n int) (err error) {
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

func (this *Window) reserveN(now time.Time, n int, t time.Duration) (time.Duration, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	wait, start, err := this.advance(now, n, t)
	if err != nil {
		return 0, err
	}
	this.addused(n, start)
	return wait, nil
}

func (this *Window) advance(now time.Time, n int, t time.Duration) (time.Duration, int, error) {
	this.initindex(now)

	var wait time.Duration
	var start int

	if this.data[this.preindex].nodeEvent.After(now) {
		wait = this.data[this.preindex].nodeEvent.Sub(now)
		start = this.preindex
	} else {
		start = this.index
	}

	i := start
	for {
		if wait > t {
			return 0, 0, fmt.Errorf("wait > inputWaitDuration")
		}
		if (i+this.windows-1)%this.datasize == this.index && i != this.index {
			return 0, 0, fmt.Errorf("wait > maxAvailableDuration")
		}
		if n > this.data[i].max-this.data[i].used {
			n = n - this.data[i].max + this.data[i].used
			lastNodeEvent := this.data[i].nodeEvent
			i = (i + 1) % this.datasize
			this.data[i].nodeEvent = lastNodeEvent.Add(this.duration)
			this.fill(i, true)
			wait += this.duration
		} else {
			return wait, start, nil
		}
	}
}

func (this *Window) initindex(now time.Time) int {
	forward := now.Sub(this.data[this.index].nodeEvent) / this.duration
	index := (int(forward) + this.index) % this.datasize
	if index != this.index {
		this.data[index].nodeEvent = this.data[this.index].nodeEvent.Add(forward * this.duration)
		this.index = index
		this.fill(index, false)
	}
	return index
}

func (this *Window) fill(index int, pre bool) {
	if !this.data[this.preindex].nodeEvent.Before(this.data[index].nodeEvent) {
		return
	}
	max := this.limit
	for i := 1; i < this.windows; i++ {
		if this.data[index].nodeEvent.Sub(this.data[(index-i+this.datasize)%this.datasize].nodeEvent) < time.Second {
			if !this.data[(index-i+this.datasize)%this.datasize].nodeEvent.Before(this.data[this.index].nodeEvent) && pre {
				max = max - this.data[(index-i+this.datasize)%this.datasize].max
			} else {
				max = max - this.data[(index-i+this.datasize)%this.datasize].used
			}
		}
	}
	this.data[index].max = max
	this.data[index].used = 0
	return
}

func (this *Window) addused(n, index int) {
	for {
		if n <= this.data[index].max-this.data[index].used {
			this.preindex = index
			this.data[index].used += n
			return
		} else {
			n = n - this.data[index].max + this.data[index].used
			this.data[index].used = this.data[index].max
			index = (index + 1) % this.datasize
		}
	}
	return
}

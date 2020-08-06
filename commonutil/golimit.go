/*===============================================================
*   Copyright (C) 2018 All rights reserved.
*
*   FileName：golimit.go
*   Author：WuGuoFu
*   Date： 2018-09-30
*   Description：
*
================================================================*/
package commonutil

import (
	"errors"
	"sync"
	"time"
)

var ErrPoolTimeout = errors.New("routine pool timeout")

type GoLimit struct {
	queue    chan struct{}
	timers   *sync.Pool
	funcPool *sync.Pool
	timeOut  time.Duration
}

func NewGoLimit(queueNum, timeout int) (gl *GoLimit) {
	gl = new(GoLimit)
	gl.timeOut = time.Duration(timeout)
	gl.queue = make(chan struct{}, queueNum)

	gl.timers = &sync.Pool{
		New: func() interface{} {
			t := time.NewTimer(time.Hour)
			t.Stop()
			return t
		},
	}

	gl.funcPool = &sync.Pool{
		New: func() interface{} {
			f := func(call func()) {
				go call()
			}
			return f
		},
	}

	return
}

/*
func GoGo() {
	for i := 0; i < 100000; i++ {
		gogo, err := Get()
		if err != nil {
			fmt.Println(err)
			continue
		}
		gogo.(func(call func()))(func() {
			fmt.Println("1111111")
			Put(gogo)
		})
	}

}
*/
func (gl *GoLimit) Put(c interface{}) {
	gl.funcPool.Put(c)
	<-gl.queue
}
func (gl *GoLimit) Get() (interface{}, error) {
	select {
	case gl.queue <- struct{}{}:
	default:
		timer := gl.timers.Get().(*time.Timer)
		timer.Reset(time.Millisecond * gl.timeOut)

		select {
		case gl.queue <- struct{}{}:
			if !timer.Stop() {
				<-timer.C
			}
			gl.timers.Put(timer)
		case <-timer.C:
			gl.timers.Put(timer)
			return nil, ErrPoolTimeout
		}
	}

	return gl.funcPool.Get(), nil
}

/*
func main() {
	go func() {
		for {
			fmt.Println(len(queue))
		}
	}()
	GoGo()
}
*/

package commonutil

import (
	"sync"

	"context"

	"github.com/tal-tech/xtools/confutil"
)

//	A common goroutine wrapper intended for future controll over goroutines
func XesGo(funcArg func()) {
	go funcArg()
}

func GetWaitGroup(ctx context.Context) *sync.WaitGroup {
	var waitGroup *sync.WaitGroup
	if ctx != nil {
		if val := ctx.Value("WaitGroup"); val != nil {
			if wg, ok := val.(*sync.WaitGroup); ok {
				waitGroup = wg
			}
		}
	}
	return waitGroup
}
func SetWaitGroup(ctx context.Context) (newctx context.Context, wg *sync.WaitGroup, ok bool) {
	if confutil.GetConf("", "GoWaitAll") == "true" {
		wg := new(sync.WaitGroup)
		ctx = context.WithValue(ctx, "WaitGroup", wg)
		return ctx, wg, true
	}
	return ctx, nil, false
}

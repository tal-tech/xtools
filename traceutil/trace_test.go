package traceutil

import (
	"context"
	"testing"
	"time"
)

func TestTrace1(t *testing.T) {
	ctx := context.Background()
	f1(ctx)
	time.Sleep(time.Millisecond * 2000)
}

func TestTrace2(t *testing.T) {
	newCtx := context.Background()
	span, newCtx := Trace(newCtx, "trace2")
	if span != nil {
		span.Finish()
	}
	time.Sleep(time.Millisecond * 2000)
}

func f1(ctx context.Context) {
	span, ctx := Trace(ctx, "f1") //返回复制的context，覆盖ctx变量,下层调用使用
	if span != nil {
		defer span.Finish()
	}
	time.Sleep(time.Millisecond * 10)
	f2(ctx)
}

func f2(ctx context.Context) {
	span, ctx := Trace(ctx, "f2")
	if span != nil {
		defer span.Finish()
	}
	time.Sleep(time.Millisecond * 10)
	f3(ctx)
	metadata := make(map[string]string, 0)
	span2, newctx := TraceRPCXInject(ctx, "rpc")
	if span2 != nil {
		defer span2.Finish()
	}

	f4(newctx, metadata)
}

func f3(ctx context.Context) {
	span, ctx := Trace(ctx, "f3")
	if span != nil {
		defer span.Finish()
	}
	time.Sleep(time.Millisecond * 10)
}

func f4(ctx context.Context, metadata map[string]string) {
	//span, ctx := TraceRPCXExtract(ctx, "f4")
	span, ctx := Trace(ctx, "f4")
	if span != nil {
		defer span.Finish()
	}
	time.Sleep(time.Millisecond * 10)
}

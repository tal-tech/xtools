package traceutil

import (
	"context"
	"time"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/smallnest/rpcx/share"
)

func TraceRPCXExtract(ctx context.Context, name string) (zipkin.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}
	sc := Extract(ExtractRPCX(ctx))
	// create Span using SpanContext if found
	sp := tracer.StartSpan(
		name,
		zipkin.Parent(sc),
	)
	sp.Annotate(time.Now(), "TraceTime")

	ctx = context.WithValue(ctx, SpanKey, sp)

	return sp, ctx
}

func TraceRPCXInject(ctx context.Context, name string) (zipkin.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}

	options := make([]zipkin.SpanOption, 0)
	if s, ok := ctx.Value(SpanKey).(zipkin.Span); ok {
		options = append(options, zipkin.Parent(s.Context()))
	}

	span := tracer.StartSpan(name, options...)
	span.Annotate(time.Now(), "TraceTime")

	meta := ctx.Value(share.ReqMetaDataKey)
	if meta == nil {
		ctx = context.WithValue(ctx, share.ReqMetaDataKey, make(map[string]string))
	}

	_ = InjectRPCX(ctx)(span.Context())

	return span, ctx

}

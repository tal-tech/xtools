package traceutil

import (
	"context"
	"time"

	zipkin "github.com/openzipkin/zipkin-go"
	"github.com/smallnest/rpcx/share"
)

func Metadata2Rpcx(ctx context.Context, metadata map[string]string) context.Context {
	meta, ok := ctx.Value(share.ReqMetaDataKey).(map[string]string)
	if !ok {
		ctx = context.WithValue(ctx, share.ReqMetaDataKey, metadata)
	} else {
		for k, v := range metadata {
			meta[k] = v
		}
	}
	return ctx
}

func TraceMetadataInject(ctx context.Context, name string, metadata map[string]string) (zipkin.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}

	span, ctx := Trace(ctx, name)
	span.Annotate(time.Now(), "TraceTime")

	_ = InjectMetadata(metadata)(span.Context())

	return span, ctx
}

func TraceMetadataExtract(ctx context.Context, name string, metadata map[string]string) (zipkin.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}
	sc := Extract(ExtractMetadata(metadata))
	// create Span using SpanContext if found
	sp := tracer.StartSpan(
		name,
		zipkin.Parent(sc),
	)
	sp.Annotate(time.Now(), "TraceTime")

	return sp, context.WithValue(ctx, SpanKey, sp)
}

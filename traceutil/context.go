package traceutil

import (
	"context"
	"time"

	zipkin "github.com/openzipkin/zipkin-go"
)

const SpanKey string = "traceSpan"

func StartSpanFromContext(ctx context.Context, name string) (zipkin.Span, context.Context) {
	options := make([]zipkin.SpanOption, 0)
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		options = append(options, zipkin.Parent(parentSpan.Context()))
	}
	span := tracer.StartSpan(name, options...)
	span.Annotate(time.Now(), "TraceTime")
	return span, context.WithValue(ctx, SpanKey, span)
}

func SpanFromContext(ctx context.Context) zipkin.Span {
	if s, ok := ctx.Value(SpanKey).(zipkin.Span); ok {
		return s
	}
	return nil
}

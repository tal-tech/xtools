package traceutil

import (
	"context"
	"errors"

	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/smallnest/rpcx/share"
)

// B3 header format.
func ExtractRPCX(ctx context.Context) propagation.Extractor {
	return func() (*model.SpanContext, error) {

		var Metadata map[string]string
		meta := ctx.Value(share.ReqMetaDataKey)
		if meta == nil {
			Metadata = make(map[string]string)
		} else {
			Metadata = meta.(map[string]string)
		}
		var (
			traceIDHeader      = Metadata[b3.TraceID]
			spanIDHeader       = Metadata[b3.SpanID]
			parentSpanIDHeader = Metadata[b3.ParentSpanID]
			sampledHeader      = Metadata[b3.Sampled]
			flagsHeader        = Metadata[b3.Flags]
		)

		return b3.ParseHeaders(
			traceIDHeader, spanIDHeader, parentSpanIDHeader, sampledHeader,
			flagsHeader,
		)
	}
}

// InjectRPCX will inject a span.Context into a RPCX Request
func InjectRPCX(ctx context.Context) propagation.Injector {
	return func(sc model.SpanContext) error {
		if (model.SpanContext{}) == sc {
			return b3.ErrEmptyContext
		}
		Metadata, ok := ctx.Value(share.ReqMetaDataKey).(map[string]string)
		if !ok {
			return errors.New("empty metadata")
		}

		if sc.Debug {
			Metadata[b3.Flags] = "1"
		} else if sc.Sampled != nil {
			// Debug is encoded as X-B3-Flags: 1. Since Debug implies Sampled,
			// so don't also send "X-B3-Sampled: 1".
			if *sc.Sampled {
				Metadata[b3.Sampled] = "1"
			} else {
				Metadata[b3.Sampled] = "0"
			}
		}

		if !sc.TraceID.Empty() && sc.ID > 0 {
			Metadata[b3.TraceID] = sc.TraceID.String()
			Metadata[b3.SpanID] = sc.ID.String()
			if sc.ParentID != nil {
				Metadata[b3.ParentSpanID] = sc.ParentID.String()
			}
		}

		return nil
	}
}

func Extract(extractor propagation.Extractor) (sc model.SpanContext) {
	psc, err := extractor()
	if psc != nil {
		sc = *psc
	}
	sc.Err = err
	return
}

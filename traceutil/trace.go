package traceutil

import (
	"context"
	"net"
	"time"

	"github.com/tal-tech/xtools/confutil"
	logger "github.com/tal-tech/loggerX"
	zipkin "github.com/openzipkin/zipkin-go"
	kafkareporter "github.com/openzipkin/zipkin-go/reporter/kafka"
	"github.com/spf13/cast"
)

var tracer *zipkin.Tracer

func init() {
	var err error
	confs := confutil.GetConfArrayMap("Trace")
	kafkaAddrs := confs["kafka"]
	if len(kafkaAddrs) <= 0 {
		return
	}
	defaultSampler := 0.001
	samplerRates := confs["sample"]
	if len(samplerRates) > 0 {
		defaultSampler = cast.ToFloat64(samplerRates[0])
	}
	serverName := "localService"
	serverNames := confs["servername"]
	if len(serverNames) > 0 {
		serverName = cast.ToString(serverNames[0])
	}
	reporter, err := kafkareporter.NewReporter(kafkaAddrs)
	if err != nil {
		logger.W("TRACEINIT", "unable to create reporter: %+v\n", err)
		return
	}
	sampler, err := zipkin.NewBoundarySampler(defaultSampler, time.Now().UnixNano())
	if err != nil {
		logger.W("TRACEINIT", "unable to create sampler: %+v\n", err)
		return
	}
	endpoint, err := zipkin.NewEndpoint(serverName, getLocalIp())
	if err != nil {
		logger.W("TRACEINIT", "unable to create local endpoint: %+v\n", err)
		return
	}
	tracer, err = zipkin.NewTracer(reporter, zipkin.WithSampler(sampler), zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		logger.W("TRACEINIT", "unable to create tracer: %+v\n", err)
		return
	}
}

func getLocalIp() string {
	addrSlice, err := net.InterfaceAddrs()
	if nil != err {
		return "localhost"
	}
	for _, addr := range addrSlice {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if nil != ipnet.IP.To4() {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

func Trace(ctx context.Context, name string) (zipkin.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}
	return StartSpanFromContext(ctx, name)
}

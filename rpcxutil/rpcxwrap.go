package rpcxutil

import (
	"context"
	"time"

	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/share"

	"github.com/spf13/cast"
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/loggerX/logtrace"
	"github.com/tal-tech/xtools/traceutil"
)

const (
	DefaultWrapType = iota

	//------上下文key标识
	//超时时间, 值为time.duration 字符串，如：1s
	// 配置文件：rpcCallTimeout 为全局， 为单独请求设置
	WRAPCLIENT_CTX_KEY_CALLTIMEOUT = "WRAPCLIENT_CTX_KEY_CALLTIMEOUT"
)

type RpcxWrap interface {
	WrapCall(client.XClient, context.Context, string, interface{}, interface{}) error
}

type DefaultWrap struct {
	serviceName string
}

func NewDefaultWrap(serviceName string) RpcxWrap {
	w := new(DefaultWrap)
	w.serviceName = serviceName
	return w
}

func (d *DefaultWrap) WrapCall(c client.XClient, ctx context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {
	tag := d.serviceName + "." + serviceMethod
	systemErrTag := "systemErr." + d.serviceName

	logger.Tx(ctx, tag, "Trace Rpc Call [destinationAddr:%s]", d.getServerAddr(ctx))
	if skip := ctx.Value("RPCXSKIPLOG"); skip == nil {
		defer func() {
			logger.Ix(ctx, tag, "[destinationAddr:%s], WrapCall args:[%+v],reply:[%+v]", d.getServerAddr(ctx), args, reply)
		}()
	}
	metadata := map[string]string{
		"logid":        cast.ToString(ctx.Value("logid")),
		"hostname":     cast.ToString(ctx.Value("hostname")),
		"IS_PLAYBACK":  cast.ToString(ctx.Value("IS_PLAYBACK")),
		"IS_BENCHMARK": cast.ToString(ctx.Value("IS_BENCHMARK")),
		"url":          cast.ToString(ctx.Value("url")),
	}

	if cast.ToString(ctx.Value("extra")) != "" {
		metadata["extra"] = cast.ToString(ctx.Value("extra"))
	}

	if appId != "" {
		timestamp, sign := genRpcAuth()
		metadata["X-Auth-TimeStamp"] = timestamp
		metadata["X-Auth-Sign"] = sign
	}
	if cast.ToString(ctx.Value("RPCX_APPID")) != "" {
		metadata["X-Auth-AppId"] = cast.ToString(ctx.Value("RPCX_APPID"))
		metadata["X-Auth-TimeStamp"] = cast.ToString(ctx.Value("RPCX_TIMESTAMP"))
		metadata["X-Auth-Sign"] = cast.ToString(ctx.Value("RPCX_SIGN"))
	}
	ctx = context.WithValue(ctx, share.ReqMetaDataKey, metadata)
	span, ctx := traceutil.TraceRPCXInject(ctx, tag)
	if span != nil {
		defer span.Finish()
	}
	logtrace.InjectTraceNodeToRpcx(ctx)

	if rpcxOptCallTimeout > 0 {
		err = d.callWithTimeout(c, ctx, serviceMethod, args, reply)
	} else {
		err = c.Call(ctx, serviceMethod, args, reply)
	}

	return
}

//超时控制
func (d *DefaultWrap) callWithTimeout(c client.XClient, ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	timeout := rpcxOptCallTimeout
	if t := cast.ToString(ctx.Value(WRAPCLIENT_CTX_KEY_CALLTIMEOUT)); t != "" {
		if td, err := time.ParseDuration(t); err == nil && td > 0 {
			timeout = td
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Call(ctx, serviceMethod, args, reply)
}

func (d *DefaultWrap) getServerAddr(ctx context.Context) (serverAddr string) {
	if metaData := ctx.Value(share.ReqMetaDataKey); metaData != nil {
		m, ok := metaData.(map[string]string)
		if !ok {
			return
		}
		if addr, ok := m["DESTINATION_ADDR"]; ok {
			return addr
		}
	}

	return
}

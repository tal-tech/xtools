package rpcxutil

import (
	"context"

	"github.com/smallnest/rpcx/client"
)

type WrapClient struct {
	xclient client.XClient
	wrap    RpcxWrap
}

type InitXClientFunc func(c client.XClient) error

func NewWrapClient(basePath, servicePath string, failMode client.FailMode, selectMode client.SelectMode, option client.Option, fns ...InitXClientFunc) *WrapClient {
	wClient := new(WrapClient)
	discovery := getClientDiscovery(basePath, servicePath)
	wClient.xclient = client.NewXClient(servicePath, failMode, selectMode, discovery, option)
	for _, fn := range fns {
		fn(wClient.xclient)
	}
	if appId != "" {
		wClient.xclient.Auth(appId)
	}
	wClient.wrap = NewDefaultWrap(servicePath)
	return wClient
}

func (w *WrapClient) WrapCall(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	return w.wrap.WrapCall(w.xclient, ctx, serviceMethod, args, reply)
}

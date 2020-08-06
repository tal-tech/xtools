//+build etcd

package rpcxutil

import (
	"github.com/docker/libkv/store"
	"github.com/smallnest/rpcx/client"
)

var regcenter = string(store.ETCD)

func initClientDiscovery(basePath string) client.ServiceDiscovery {
	return client.NewEtcdDiscoveryTemplate(basePath, GetSdAddrs(), nil)
}

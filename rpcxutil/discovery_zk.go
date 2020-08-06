//+build zookeeper

package rpcxutil

import (
	"github.com/docker/libkv/store"
	"github.com/smallnest/rpcx/client"
)

var regcenter = string(store.ZK)

func initClientDiscovery(basePath string) client.ServiceDiscovery {
	return client.NewZookeeperDiscoveryTemplate(basePath, GetSdAddrs(), nil)
}

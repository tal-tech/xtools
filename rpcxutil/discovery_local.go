package rpcxutil

import (
	"sync"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/rpcxutil/store"
	"github.com/smallnest/rpcx/client"
)

type localDiscovery struct {
	basePath   string
	serverName string
	kv         store.Store
	pairs      []*client.KVPair
	chans      []chan []*client.KVPair
	mu         sync.Mutex

	stopCh chan struct{}
}

func NewLocalDiscovery(basePath string, serverName string) client.ServiceDiscovery {

	kv, err := store.NewStore(store.LOCAL_FILE, storeConfig)
	if err != nil {
		panic(err)
	}
	return NewLocalDiscoveryStore(trimPath(basePath), serverName, kv)
}

func NewLocalDiscoveryStore(basePath string, serverName string, kv store.Store) client.ServiceDiscovery {
	ls := localDiscovery{
		kv:         kv,
		basePath:   basePath,
		serverName: serverName,
	}

	ls.stopCh = make(chan struct{})
	ls.pairs = ls.GetServices()

	go ls.watch()
	return &ls
}

func (ls *localDiscovery) getStoreKey() string {
	return ls.basePath + "/" + ls.serverName
}

func (ls *localDiscovery) GetServices() (ret []*client.KVPair) {
	servers, err := ls.kv.GetServers(ls.getStoreKey())
	if err != nil {
		logger.E("LocalDiscoery.GetServices", "cannot get services of from registry: %v, err: %v", ls.getStoreKey(), err)
		ret = make([]*client.KVPair, 0)
		return
	}

	ret = make([]*client.KVPair, 0, len(servers))
	for _, s := range servers {
		ret = append(ret, &client.KVPair{Key: s.Key, Value: s.Value})
	}

	return
}

func (ls *localDiscovery) WatchService() (ret chan []*client.KVPair) {
	ch := make(chan []*client.KVPair, 10)
	ls.chans = append(ls.chans, ch)
	return ch
}

func (ls *localDiscovery) RemoveWatcher(ch chan []*client.KVPair) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	var chans []chan []*client.KVPair
	for _, c := range ls.chans {
		if c == ch {
			continue
		}

		chans = append(chans, c)
	}

	ls.chans = chans
}

func (ls *localDiscovery) Clone(servicePath string) (dis client.ServiceDiscovery) {
	return NewLocalDiscoveryStore(ls.basePath, servicePath, ls.kv)
}

func (ls *localDiscovery) Close() {
	close(ls.stopCh)
	return
}

func (ls *localDiscovery) watch() {
	logTag := "LocalDiscoery.watch"
	c, _ := ls.kv.WatchServers(ls.getStoreKey())
	for {
		select {
		case <-ls.stopCh:
			logger.I(logTag, "discovery has been closed")
			return
		case ps := <-c:
			if ps == nil {
				time.Sleep(time.Second)
				continue
			}

			if len(ps) == 0 {
				logger.W(logTag, "本地储存的服务列表为空！ serverPath:%v", ls.getStoreKey())
				continue
			}

			var pairs []*client.KVPair // latest servers
			for _, p := range ps {
				pairs = append(pairs, &client.KVPair{Key: p.Key, Value: p.Value})
			}
			ls.pairs = pairs

			for _, ch := range ls.chans {
				ch := ch
				go func() {
					defer func() {
						if r := recover(); r != nil {

						}
					}()

					select {
					case ch <- pairs:
					case <-time.After(time.Minute):
						logger.E(logTag, "chan is full and new change has been dropped")
					}
				}()
			}
		}
	}
}

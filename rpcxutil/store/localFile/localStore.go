package localFile

import (
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/jsutil"
	"github.com/tal-tech/xtools/rpcxutil/store"
	"github.com/peterbourgon/diskv"
	"strings"
	"sync"
	"time"
)

var (
	defaultOption = &store.Options{
		BasePath: "localServerStorage",
	}
)

type LocalStore struct {
	storage        *diskv.Diskv
	watchedServers map[string][]chan []store.KVPair

	//监控server变化的通道
	changeCh chan StorageUnit

	lock sync.RWMutex
}

type StorageUnit struct {
	//存储server列表的缓存key
	ServerPath string
	Value      []store.KVPair
}

func Register() {
	store.AddStore(store.LOCAL_FILE, NewLocalStore)
}

func NewLocalStore(opt *store.Options) (store.Store, error) {

	if opt == nil {
		opt = defaultOption
	}

	ls := &LocalStore{
		storage: diskv.New(diskv.Options{
			BasePath:     opt.BasePath,
			CacheSizeMax: 1024 * 1024, //1MB
		}),
		watchedServers: make(map[string][]chan []store.KVPair),
		changeCh:       make(chan StorageUnit, 20),
	}

	go ls.watchServerChange()

	return ls, nil
}

func (ls *LocalStore) GetServers(serverPath string) (kv []store.KVPair, err error) {
	v, err := ls.get(ls.convertKey(serverPath))
	if err != nil {
		logger.E("LocalStore.List", "ls.Get err:%+v, key:%v", err, serverPath)
		return
	}

	kv = make([]store.KVPair, 0)
	_ = jsutil.Json.UnmarshalFromString(v, &kv)

	return kv, nil
}

func (ls *LocalStore) StoreServers(serverPath string, kv []store.KVPair) (err error) {
	str, err := jsutil.Json.MarshalToString(kv)
	if err != nil {
		logger.E("LocalStore.StoreServers", "jsutil.Json.MarshalToString err:%+v, key:%v", err, serverPath)
		return
	}

	logger.D("LocalStore.StoreServers", "本地保存服务列表 serverPath:%v", serverPath)

	err = ls.set(serverPath, str)
	if err == nil && len(kv) > 0 {
		s := StorageUnit{
			ServerPath: serverPath,
			Value:      kv,
		}
		ls.changeCh <- s
	}

	return err
}

func (ls *LocalStore) WatchServers(serverPath string) (kvChan chan []store.KVPair, err error) {
	//预留一定的buffer 以免后续处理速度不匹配导致阻塞
	kvChan = make(chan []store.KVPair, 10)

	ls.lock.Lock()
	defer ls.lock.Unlock()

	if _, ok := ls.watchedServers[serverPath]; !ok {
		ls.watchedServers[serverPath] = make([]chan []store.KVPair, 0)
	}

	ws := ls.watchedServers[serverPath]
	ws = append(ws, kvChan)
	ls.watchedServers[serverPath] = ws

	return
}

func (ls *LocalStore) watchServerChange() {
	for su := range ls.changeCh {
		logger.D("LocalStore.watchServerChange", "监控到服务列表刷新， severpath:%v, servers:%+v", su.ServerPath, su.Value)
		ls.lock.RLock()
		wsChan, ok := ls.watchedServers[su.ServerPath]
		ls.lock.RUnlock()
		if !ok {
			continue
		}
		logger.D("LocalStore.watchServerChange", "传递服务列表至所有监听者-start")
		for _, ch := range wsChan {
			ch := ch
			go func() {
				defer func() {
					if r := recover(); r != nil {

					}
				}()

				select {
				case ch <- su.Value:
				case <-time.After(time.Second * 10):
					logger.E("LocalStore.watchServerChange", "chan is full and new change has been dropped")
				}
			}()
		}
		logger.D("LocalStore.watchServerChange", "传递服务列表至所有监听者-end")
	}
}

func (ls *LocalStore) set(key string, value string) (err error) {
	filekey := ls.convertKey(key)
	err = ls.storage.Write(filekey, []byte(value))
	if err != nil {
		logger.E("LocalStore.Set", "err:%+v, key:%v, value:%v", err, key, value)
	}

	return err
}

func (ls *LocalStore) get(key string) (v string, err error) {
	b, err := ls.storage.Read(ls.convertKey(key))
	if err != nil {
		logger.E("LocalStore.Get", "ls.storage.Read err:%+v, key:%v", err, key)
		return
	}
	return string(b), err
}

func (ls *LocalStore) Del(key string) (err error) {
	err = ls.storage.Erase(ls.convertKey(key))
	if err != nil {
		logger.E("LocalStore.Del", "err:%+v, key:%v", err, key)
	}

	return
}

func (ls *LocalStore) convertKey(key string) string {
	return strings.Replace(key, "/", "-", -1)
}

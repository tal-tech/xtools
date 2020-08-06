package store

import (
	"fmt"
	"sort"
	"strings"
)

type Backend string

const (
	//本地文件存储
	LOCAL_FILE Backend = "localFile"
)

var (
	ErrBackendNotSupported = "Backend storage not supported yet, please choose one of"
)

type Store interface {
	GetServers(serverPath string) (kv []KVPair, err error)
	StoreServers(serverPath string, kv []KVPair) (err error)
	WatchServers(serverPath string) (kvChan chan []KVPair, err error)
}

type KVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Initialize func(options *Options) (Store, error)

var (
	// Backend initializers
	initializers = make(map[Backend]Initialize)

	supportedBackend = func() string {
		keys := make([]string, 0, len(initializers))
		for k := range initializers {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		return strings.Join(keys, ", ")
	}()
)

func NewStore(backend Backend, options *Options) (Store, error) {
	if init, exists := initializers[backend]; exists {
		return init(options)
	}

	return nil, fmt.Errorf("%s %s", ErrBackendNotSupported, supportedBackend)
}

func AddStore(store Backend, init Initialize) {
	initializers[store] = init
}

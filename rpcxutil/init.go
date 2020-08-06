package rpcxutil

import (
	"sync"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/spf13/cast"

	"github.com/tal-tech/xtools/addrutil"
	"github.com/tal-tech/xtools/confutil"
	"github.com/smallnest/rpcx/client"
	"github.com/smallnest/rpcx/protocol"
	"github.com/smallnest/rpcx/share"
)

var (
	sdaddrs            []string //注册中心地址
	testingAddrs       []string //灰度测试机器ip
	defaultBasePath    string
	group              string
	templates          map[string]client.ServiceDiscovery
	appId              string
	appKey             string
	eventHooks         *discoveryEventHooks //服务发现实例化事件同步回调
	mu                 sync.RWMutex
	rpcxOptCallTimeout time.Duration //调用超时
	rpcxConnectTimeout time.Duration //连接超时
	rpcxRetry          int           //重试次数
)

//实例化服务发现实例事件
type discoveryEvent struct {
	//是否成功初始化
	SuccessInit bool
	//服务发现实例
	Discovery client.ServiceDiscovery

	ServerPath string
}

func NewEeventHooks() *discoveryEventHooks {
	e := &discoveryEventHooks{
		callbacks: make([]eventCallBackFunc, 0),
	}
	return e
}

type eventCallBackFunc func(e discoveryEvent)
type discoveryEventHooks struct {
	callbacks []eventCallBackFunc
}

func (h *discoveryEventHooks) AddFunc(fn eventCallBackFunc) {
	h.callbacks = append(h.callbacks, fn)
}

func (h *discoveryEventHooks) syncCall(e discoveryEvent) {
	for _, fn := range h.callbacks {
		fn(e)
	}
}

func init() {
	templates = make(map[string]client.ServiceDiscovery, 0)

	sdmap := confutil.GetConfArrayMap("Registration")
	sdaddrs = sdmap["addrs"]
	testingAddrs = sdmap["testingAddrs"]
	if callTimeout := confutil.GetConf("Registration", "rpcCallTimeout"); callTimeout != "" {
		if d, err := time.ParseDuration(callTimeout); err == nil {
			rpcxOptCallTimeout = d
		}
	}
	rpcxConnectTimeout = time.Second //default 1s
	if connectTimeout := confutil.GetConf("Registration", "rpcConnectTimeout"); connectTimeout != "" {
		if d, err := time.ParseDuration(connectTimeout); err == nil {
			rpcxConnectTimeout = d
		}
	}
	rpcxRetry = 2 //default 2
	if retry := confutil.GetConf("Registration", "rpcRetry"); retry != "" {
		if cast.ToInt(retry) > 0 {
			rpcxRetry = cast.ToInt(retry)
		}
	}

	defaultBasePath = confutil.GetConf("Registration", "basePath")
	group = confutil.GetConf("Registration", "group")
	if len(sdaddrs) <= 0 {
		sdmap = confutil.GetConfArrayMap("Registry")
		sdaddrs = sdmap["addrs"]
		testingAddrs = sdmap["testingAddrs"]
		defaultBasePath = confutil.GetConf("Registry", "basePath")
		group = confutil.GetConf("Registry", "group")
	}
	rpcAuth := confutil.GetConfStringMap("RpcxAuth")
	appId = rpcAuth["appId"]
	appKey = rpcAuth["appKey"]
	eventHooks = NewEeventHooks()
}

func GetSdAddrs() []string {
	return sdaddrs
}

func GetServiceBasePath() string {
	return defaultBasePath
}

func GetFailMode() client.FailMode {
	return client.Failover
}

func GetRegCenter() string {
	return regcenter
}

func GetSelectMode() client.SelectMode {
	return client.WeightedRoundRobin
}

func GetClientOption() client.Option {
	option := client.Option{
		Retries:           rpcxRetry,
		RPCPath:           share.DefaultRPCPath,
		ConnectTimeout:    rpcxConnectTimeout,
		SerializeType:     protocol.MsgPack,
		CompressType:      protocol.None,
		BackupLatency:     10 * time.Millisecond,
		Heartbeat:         true,
		HeartbeatInterval: 1 * time.Second,
		Group:             group,
		WriteTimeout:      2 * time.Second,
		ReadTimeout:       2 * time.Second,
	}
	//判断当前客户端机器是否为灰度机
	ip, err := addrutil.Extract("")
	if ip != "" && err == nil && len(testingAddrs) > 0 {
		for _, scale := range testingAddrs {
			if scale == ip {
				option.Group = "testing"
				break
			}
		}
	}
	//if failed 5 times, return error immediately, and will try to connect after 10 seconds
	option.GenBreaker = func() client.Breaker {
		return client.NewConsecCircuitBreaker(5, 10*time.Second)
	}
	return option
}

func getClientDiscovery(basePath, servicePath string) (discovery client.ServiceDiscovery) {
	mu.Lock()
	defer mu.Unlock()
	if basePath == "" {
		basePath = defaultBasePath
	}

	defer func() {
		successInit := true
		if e := recover(); e != nil {
			successInit = false
			logger.E("rpcxutil.init.GetClientDiscovery", "初始化服务发现对象失败:err:%+v", e)
		}

		eventHooks.syncCall(discoveryEvent{
			SuccessInit: successInit,
			Discovery:   discovery, //初始化失败时，接口为nil
			ServerPath:  trimPath(basePath) + "/" + servicePath,
		})
	}()

	if template, ok := templates[basePath]; !ok {
		template = initClientDiscovery(basePath)
		templates[basePath] = template
		discovery = template.Clone(servicePath)
	} else {
		discovery = template.Clone(servicePath)
	}

	return discovery
}

func GetAppId() string {
	return appId
}

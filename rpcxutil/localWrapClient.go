package rpcxutil

import (
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/confutil"
	"github.com/tal-tech/xtools/rpcxutil/store"
	"github.com/tal-tech/xtools/rpcxutil/store/localFile"
	"github.com/smallnest/rpcx/client"
	"os"
	"strings"
)

var (
	REGMonitor *RegistrationMonitor
	//本地服务发现
	localTemplates map[string]client.ServiceDiscovery
	localStorage   store.Store
	storeConfig    *store.Options
)

func init() {
	localTemplates = make(map[string]client.ServiceDiscovery)
	//注册中心健康状态监控
	REGMonitor = newRegistrationMonitor()
	sdmap := confutil.GetConfStringMap("Registration")
	if v, ok := sdmap["enableLocalDiscovery"]; ok && v == "true" {
		REGMonitor.EnableLocalDiscovery()
		localFile.Register()

		storeConfig = &store.Options{
			BasePath: "/home/logs/localServerStorage/" + getServername(),
		}

		if localPath, ok := sdmap["localStorage"]; ok {
			storeConfig.BasePath = localPath
		}

		localStorage, _ = store.NewStore(store.LOCAL_FILE, storeConfig)
		eventHooks.AddFunc(watchService)
	}
}

func getServername() string {
	arg0 := os.Args[0]
	items := strings.Split(arg0, "/")
	return items[len(items)-1]
}

type LocalWrapClient struct {
	//wrapClietn 目前不是一个接口，修改的话，向上不兼容
	WrapClient
}

func NewLocalWrapClient(basePath, servicePath string, failMode client.FailMode, selectMode client.SelectMode, option client.Option, fns ...InitXClientFunc) *WrapClient {
	wClient := new(LocalWrapClient)
	discovery := wClient.getLocalClientDiscovery(basePath, servicePath)
	wClient.xclient = client.NewXClient(servicePath, failMode, selectMode, discovery, option)
	for _, fn := range fns {
		fn(wClient.xclient)
	}
	if appId != "" {
		wClient.xclient.Auth(appId)
	}
	wClient.wrap = NewDefaultWrap(servicePath)

	logger.D("NewLocalWrapClient", "注入本地服务发现")
	return &wClient.WrapClient
}

func (this *LocalWrapClient) getLocalClientDiscovery(basePath, servicePath string) (discovery client.ServiceDiscovery) {
	mu.Lock()
	defer mu.Unlock()
	if basePath == "" {
		basePath = defaultBasePath
	}

	if template, ok := localTemplates[basePath]; !ok {
		template = NewLocalDiscovery(basePath, servicePath)
		localTemplates[basePath] = template
		return template
	} else {
		return template.Clone(servicePath)
	}
}

//监控注册中心server变化，并存储到本地
func watchService(e discoveryEvent) {
	if !REGMonitor.localDiscoveryIsEnabled {
		return
	}

	if !e.SuccessInit {
		REGMonitor.SetStatusInFault()
		return
	}

	go func() {
		for pairs := range e.Discovery.WatchService() {
			servers := make([]store.KVPair, 0, len(pairs))
			for _, p := range pairs {
				servers = append(servers, store.KVPair{p.Key, p.Value})
			}

			if len(pairs) > 0 {
				if err := localStorage.StoreServers(e.ServerPath, servers); err != nil {
					logger.E("watchService", "DefaultStore.StoreServers err")
				}
			}
		}
	}()
}

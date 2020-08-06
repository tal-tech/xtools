//监控注册中心健康状态

package rpcxutil

import (
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
)

type serverDiscoveryStatus int

const (
	//注册中心正常
	SERVER_DISCOVERY_STATUS_NORMAL serverDiscoveryStatus = 1

	//注册中心异常
	SERVER_DISCOVERY_STATUS_INFALT serverDiscoveryStatus = 2
)

type RegistrationMonitor struct {
	//是否开启本地服务发现模式，仅当注册中心故障时，才启用，待故障恢复后，后续的请求会切换回注册中心服务发现模式
	localDiscoveryIsEnabled bool
	//注册中心健康状态
	discoveryStatus serverDiscoveryStatus
}

func newRegistrationMonitor() *RegistrationMonitor {
	m := RegistrationMonitor{
		discoveryStatus: SERVER_DISCOVERY_STATUS_NORMAL,
	}

	go m.checkDiscoveryStatus()

	return &m
}

func (m *RegistrationMonitor) EnableLocalDiscovery() {
	m.localDiscoveryIsEnabled = true
}

func (m *RegistrationMonitor) SetStatusInFault() {
	if !m.localDiscoveryIsEnabled {
		return
	}

	m.discoveryStatus = SERVER_DISCOVERY_STATUS_INFALT
}

func (m *RegistrationMonitor) SetStatusNormal() {
	m.discoveryStatus = SERVER_DISCOVERY_STATUS_NORMAL
}

func (m *RegistrationMonitor) IsInFault() bool {
	return m.discoveryStatus == SERVER_DISCOVERY_STATUS_INFALT
}

func (m *RegistrationMonitor) IsNormal() bool {
	return m.discoveryStatus == SERVER_DISCOVERY_STATUS_NORMAL
}

func (m *RegistrationMonitor) checkDiscoveryStatus() {
	for {
		time.Sleep(time.Second)

		if !m.localDiscoveryIsEnabled || m.IsNormal() {
			continue
		}

		if err := checkStoreStateIsNormal(); err == nil {
			m.SetStatusNormal()
			logger.I("RegistrationMonitor.checkDiscoveryStatus", "注册中心故障已恢复")
		} else {
			logger.E("RegistrationMonitor.checkDiscoveryStatus", "注册中心仍处于故障中, err:%+v", err)
		}
	}
}

func checkStoreStateIsNormal() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = logger.NewError(r)
			return
		}
	}()

	kv, err := libkv.NewStore(store.Backend(GetRegCenter()), GetSdAddrs(), nil)
	if err != nil {
		return logger.NewError(err)
	}

	defer kv.Close()

	//不同的注册中心，路径方式有所不同
	testKey := "rpcxutilfoo"
	if GetRegCenter() == string(store.ETCD) {
		testKey = "/" + testKey
	}

	if err = kv.Put(testKey, []byte("bar"), &store.WriteOptions{TTL: time.Second}); err != nil {
		return logger.NewError(err)
	}

	return nil
}

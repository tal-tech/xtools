package expvarutil

import (
	"expvar"
	"log"
	"net"
	"net/http"
	"runtime"
	"runtime/debug"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/confutil"
	"github.com/spf13/cast"
)

func init() {
	//此路由已经和数据收集组约定好了，请不要更改
	expvar.Publish("customMetrics", expvar.Func(NewCustomMetrics().GetAllMetrics))
}

func Expvar() {
	grace := confutil.GetConf("Server", "grace")
	if grace == "true" {
		InitPort()
	} else {
		expvarStartWithoutGrace()
	}
}

func expvarStartWithoutGrace() {
	expvarEnabled := confutil.GetConf("Expvar", "enable")
	if expvarEnabled != "true" {
		log.Printf("[expvarutil] Expvar not enabled\n")
		return
	}

	expvarPort := confutil.GetConf("Expvar", "port")
	if len(expvarPort) == 0 {
		log.Printf("[expvarutil]", "Expvar enabled but no port found in Expvar section\n")
		return
	}

	go func() {
		defer func() {
			if x := recover(); x != nil {
				log.Printf("[expvarutil]", "panic captued\n")
			}
		}()

		log.Printf("[expvarutil] listening on port:%+v\n", expvarPort)
		if err := http.ListenAndServe(":"+expvarPort, nil); err != nil {
			log.Printf("[expvarutil] ListenAndServe failed with err: %+v\n", err)
		}
	}()

}

var ExpvarPort string
var Eserver = &http.Server{Addr: ExpvarPort}

func InitPort() {
	expvarEnabled := confutil.GetConf("Expvar", "enable")
	if expvarEnabled != "true" {
		log.Printf("[expvarutil] Expvar not enabled\n")
		return
	}

	ExpvarPort = ":" + confutil.GetConf("Expvar", "port")
	if len(ExpvarPort) == 0 {
		log.Printf("[expvarutil]", "Expvar enabled but no port found in Expvar section\n")
		return
	}
	logger.I("expvar", "open expvar on port:%s", ExpvarPort)

}

//自定义指标项。 通过扩充字段来新增指标
type CustomMetrics struct {
	//协程数量
	NumGoroutines int
}

func NewCustomMetrics() *CustomMetrics {
	return &CustomMetrics{}
}

func (cm *CustomMetrics) SetNumGoroutine() {
	cm.NumGoroutines = runtime.NumGoroutine()
	return
}

func (cm *CustomMetrics) GetAllMetrics() interface{} {
	cm.SetNumGoroutine()

	return cm
}

func Start(l net.Listener) {
	err := Eserver.Serve(l)
	if err != nil {
		logger.W("ServerError", "Unhandled error: %v\n stack:%v", err.Error(), cast.ToString(debug.Stack()))
	}
}

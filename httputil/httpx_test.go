package httputil

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/loggerX/builders"
	"github.com/tal-tech/loggerX/logtrace"
)

func TestMain(m *testing.M) {
	config := logger.NewLogConfig()
	config.LogPath = "/home/logs/xeslog/httputil/httputil.log"
	logger.InitLogWithConfig(config)
	//或使用xml配置 logger.InitLogger("conf/log.xml")
	defer logger.Close()
	builder := new(builders.TraceBuilder)
	builder.SetTraceDepartment("HS-Golang")
	builder.SetTraceVersion("0.1")
	logger.SetBuilder(builder)
	m.Run()
}

func TestGetX(t *testing.T) {
	//初始化trace信息 一次完整调用前执行
	ctx := context.WithValue(context.Background(), logtrace.GetMetadataKey(), logtrace.GenLogTraceMetadata())
	header := make(map[string]string, 0)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	header["Httputil-Example"] = "get"
	ret, e := GetX(ctx, "http://127.0.0.1:9898/demo/test", header)
	fmt.Printf("ret:%s,e:%v", ret, e)
}

func TestPostX(t *testing.T) {
	//初始化trace信息 一次完整调用前执行
	ctx := context.WithValue(context.Background(), logtrace.GetMetadataKey(), logtrace.GenLogTraceMetadata())
	u := url.Values{}
	u.Set("param", "fefe")
	header := make(map[string]string, 0)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	header["Httputil-Example"] = "post"
	ret, e := PostX(ctx, "http://127.0.0.1:9898/demo/test", u, header)
	fmt.Printf("ret:%s,e:%v", ret, e)
}

package headerutil

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/loggerX/logtrace"
	"github.com/tal-tech/xtools/confutil"
	"github.com/spf13/cast"
)

var (
	appId      string
	appKey     string
	ircId      string
	ircKey     string
	ircVersion string
	ircBizId   string
)

func RegisterAppSecret(id, key string) {
	appId = id
	appKey = key
}

func RegisterIrcSecret(id, key, version, bizId string) {
	ircId = id
	ircKey = key
	ircVersion = version
	ircBizId = bizId
}

func init() {
	appId = confutil.GetConf("GatewayAuth", "appId")
	if len(appId) == 0 {
		logger.W("SDK init", "SDK.init() GatewayAuth.appId not found in conf file")
	}

	appKey = confutil.GetConf("GatewayAuth", "appKey")
	if len(appKey) == 0 {
		logger.W("SDK init", "SDK.init() GatewayAuth.appKey not found in conf file")
	}
	ircId = confutil.GetConf("Irc", "appId")
	if len(ircId) == 0 {
		logger.W("SDK init", "SDK.init() Irc.appId not found in conf file")
	}
	ircKey = confutil.GetConf("Irc", "appKey")
	if len(ircKey) == 0 {
		logger.W("SDK init", "SDK.init() Irc.appKey not found in conf file")
	}
	ircVersion = confutil.GetConf("Irc", "version")
	if len(ircVersion) == 0 {
		logger.W("SDK init", "SDK.init() Irc.version not found in conf file")
		ircVersion = "1"
	}
	ircBizId = confutil.GetConf("Irc", "bizId")
	if len(ircBizId) == 0 {
		logger.W("SDK init", "SDK.init() Irc.bizId not found in conf file")
	}
}

func GenGatewayAuthHeader(ctx context.Context, header map[string]string) {
	header["X-Auth-Appid"] = appId
	now := strconv.Itoa(int(time.Now().Unix()))
	md5Ctx := md5.New()
	header["X-Auth-TimeStamp"] = now
	md5Ctx.Write([]byte(appId + "&" + now + appKey))
	cipherStr := md5Ctx.Sum(nil)
	signstr := hex.EncodeToString(cipherStr)
	header["X-Auth-Sign"] = signstr
}

func GenIRCAuthHeader(ctx context.Context, header map[string]string) {
	now := strconv.Itoa(int(time.Now().Add(900 * time.Second).Unix())) //valid for 15 minutes
	header["X-PS-AppId"] = ircId
	header["X-PS-Timestamp"] = now
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(ircId + "\n" + ircKey + "\n" + now + "\n"))
	cipherStr := md5Ctx.Sum(nil)
	signstr := hex.EncodeToString(cipherStr)
	header["X-PS-Signature"] = signstr
	header["X-PS-Version"] = ircVersion
	header["X-PS-BizId"] = ircBizId
}

func GenTraceHeader(ctx context.Context, header map[string]string) {
	if ctx == nil {
		return
	}
	logtrace := logtrace.ExtractTraceNodeFromXesContext(ctx)

	if traceId := logtrace.Get("x_trace_id"); traceId != "" {
		header["traceid"] = strings.Replace(traceId, "\"", "", -1)
	}

	if rpcId := logtrace.Get("x_rpcid"); rpcId != "" {
		header["rpcid"] = strings.Replace(rpcId, "\"", "", -1)
	}

	bench := ctx.Value("IS_BENCHMARK")
	if cast.ToString(bench) == "1" {
		header["Xes-Request-Type"] = "performance-testing"
	}
}

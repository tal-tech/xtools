package rpcxutil

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func trimPath(basePath string) string {
	if basePath[0] == '/' {
		basePath = basePath[1:]
	}

	if len(basePath) > 1 && strings.HasSuffix(basePath, "/") {
		basePath = basePath[:len(basePath)-1]
	}

	return basePath
}

func genRpcAuth() (string, string) {
	now := strconv.Itoa(int(time.Now().Unix()))
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(appId + "&" + now + appKey))
	cipherStr := md5Ctx.Sum(nil)
	signstr := hex.EncodeToString(cipherStr)
	return now, signstr
}

func genRpcAuthCtx(appId, appKey string) (string, string) {
	now := strconv.Itoa(int(time.Now().Unix()))
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(appId + "&" + now + appKey))
	cipherStr := md5Ctx.Sum(nil)
	signstr := hex.EncodeToString(cipherStr)
	return now, signstr
}

package httputil

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/headerutil"
)

func PostX(ctx context.Context, target string, values url.Values, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	strval := values.Encode()
	return PostRawX(ctx, target, strval, headers, client...)
}

func GetX(ctx context.Context, target string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	return DoRawX(ctx, "GET", target, "", headers, client...)
}

func PostRawX(ctx context.Context, target string, strval string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	return DoRawX(ctx, "POST", target, strval, headers, client...)
}

func DoRawX(ctx context.Context, method, target string, strval string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	initConf()
	logger.Tx(ctx, "httputil", "Trace Http Call [Url:%s]", target)
	headerutil.GenTraceHeader(ctx, headers)
	if enableLog {
		defer logger.Ix(ctx, "httputil", "DoRawX method:%s,target:%s,param:%s,header:%v,ret:%s", method, target, strval, headers, ret)
	}
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Transport = transport
		httpclient.Timeout = time.Millisecond * time.Duration(timeout)
	}

	for i := 0; i < retry+1; i++ {
		var resp *http.Response
		var req *http.Request
		err = nil
		body := bytes.NewBufferString(strval)
		req, err = http.NewRequest(method, target, body)
		if err != nil {
			err = logger.NewError(err)
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if headers != nil && len(headers) > 0 {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		resp, err = httpclient.Do(req)
		if err != nil {
			err = logger.NewError(err)
			time.Sleep(time.Millisecond * 10)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			ret, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			break
		} else {
			ret, _ = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			err = logger.NewError(resp.Status)
			time.Sleep(time.Millisecond * 10)
			continue
		}
	}
	return
}

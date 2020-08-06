package httputil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"sync"
	"time"

	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/confutil"
	"github.com/spf13/cast"
)

var (
	once                  sync.Once
	timeout               int64 = 10000 //client总超时时间 默认10s
	disableKeepAlives     bool          //是否禁用长连接，默认开启
	tLSHandshakeTimeout   int64 = 10    //限制TLS握手使用的时间
	maxIdleConns          int   = 100   //最大空闲连接数
	maxConnsPerHost       int           //每个host的最大连接数
	maxIdleConnsPerHost   int   = 2     //每个host的最大空闲连接数
	idleConnTimeout       int64 = 90    //空闲连接在连接池中的保留时间
	retry                 int   = 2     //请求失败重试次数
	expectContinueTimeout int64 = 1
	enableLog             bool  = false
	transport             http.RoundTripper
)

func Post(target string, values url.Values, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	strval := values.Encode()
	return PostRaw(target, strval, headers, client...)
}

func GetRaw(target string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	return DoRaw("GET", target, "", headers, client...)
}

func PostRaw(target string, strval string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	return DoRaw("POST", target, strval, headers, client...)
}

func DoRaw(method, target string, strval string, headers map[string]string, client ...interface{}) (ret []byte, err error) {
	initConf()
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

func initConf() {
	once.Do(func() {
		cfg := confutil.GetConfStringMap("HttpClient")
		if v, ok := cfg["timeout"]; ok {
			timeout = cast.ToInt64(v)
		}
		if v, ok := cfg["disableKeepAlives"]; ok {
			disableKeepAlives = cast.ToBool(v)
		}
		if v, ok := cfg["tLSHandshakeTimeout"]; ok {
			tLSHandshakeTimeout = cast.ToInt64(v)
		}
		if v, ok := cfg["expectContinueTimeout"]; ok {
			expectContinueTimeout = cast.ToInt64(v)
		}
		if v, ok := cfg["maxIdleConns"]; ok {
			maxIdleConns = cast.ToInt(v)
		}
		if v, ok := cfg["maxConnsPerHost"]; ok {
			maxConnsPerHost = cast.ToInt(v)
		}
		if v, ok := cfg["maxIdleConnsPerHost"]; ok {
			maxIdleConnsPerHost = cast.ToInt(v)
		}
		if v, ok := cfg["idleConnTimeout"]; ok {
			idleConnTimeout = cast.ToInt64(v)
		}
		if v, ok := cfg["retry"]; ok {
			retry = cast.ToInt(v)
		}
		if v, ok := cfg["enableLog"]; ok {
			enableLog = cast.ToBool(v)
		}
		transport = &http.Transport{
			DisableKeepAlives: disableKeepAlives,
			Proxy:             http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          cast.ToInt(maxIdleConns),
			IdleConnTimeout:       time.Duration(idleConnTimeout) * time.Second,
			TLSHandshakeTimeout:   time.Duration(tLSHandshakeTimeout) * time.Second,
			ExpectContinueTimeout: time.Duration(expectContinueTimeout) * time.Second,
			MaxConnsPerHost:       maxConnsPerHost,
			MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		}
	})
}

func Get(url string, client ...interface{}) (ret []byte, err error) {
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Timeout = time.Millisecond * 10000
	}
	for i := 0; i < 3; i++ {
		var resp *http.Response
		resp, err = httpclient.Get(url)
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

func GetReader(url string, client ...interface{}) (ret io.ReadCloser, err error) {
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Timeout = time.Millisecond * 10000
	}
	var resp *http.Response
	resp, err = httpclient.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("ResponseError:%s", resp.Status)
		return
	}
	ret = resp.Body
	return
}

func GetV(url string, client ...interface{}) (ret []byte, header http.Header, err error) {
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Timeout = time.Millisecond * 10000
	}
	var resp *http.Response
	resp, err = httpclient.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("ResponseError:%s", resp.Status)
		return

	}
	header = resp.Header
	ret, err = ioutil.ReadAll(resp.Body)
	return
}

func HeaderFilter(header http.Header) http.Header {
	filtered := make(http.Header)
	for k, _ := range header {
		if k == "Content-Length" || k == "Cache-Control" || k == "Content-Md5" || k == "Content-Type" || k == "Last-Modified" {
			filtered.Set(k, header.Get(k))
		}
	}
	return filtered
}

func PostFile(target, name string, data []byte, headers http.Header, client ...interface{}) (ret []byte, err error) {
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Timeout = time.Millisecond * 10000
	}
	for i := 0; i < 3; i++ {
		var resp *http.Response
		var req *http.Request
		err = nil
		//multipart
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		//if fw, e := CreateFormFile("filename", name, http.DetectContentType(data), w); e != nil {
		if fw, e := w.CreateFormFile("filename", name); e != nil {
			err = logger.NewError(e)
			return
		} else if _, err = io.Copy(fw, bytes.NewReader(data)); err != nil {
			return
		} else {
			w.Close()
			req, err = http.NewRequest("POST", target, &b)
			if err != nil {
				err = logger.NewError(err)
				time.Sleep(time.Millisecond * 10)
				continue
			}
			if headers != nil && len(headers) > 0 {
				req.Header = headers
			}
			req.Header.Set("Content-Type", w.FormDataContentType())
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

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func CreateFormFile(fieldname, filename, contentType string, w *multipart.Writer) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			quoteEscaper.Replace(fieldname), quoteEscaper.Replace(filename)))
	h.Set("Content-Type", contentType)
	return w.CreatePart(h)
}

func Head(url string, client ...interface{}) (header http.Header, err error) {
	httpclient := http.DefaultClient
	if client != nil && len(client) == 1 {
		if c, ok := client[0].(*http.Client); ok && c != nil {
			httpclient = c
		}
	} else {
		httpclient.Timeout = time.Millisecond * 10000
	}
	var resp *http.Response
	resp, err = httpclient.Head(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("ResponseError:%s", resp.Status)
		return

	}
	header = resp.Header
	return
}

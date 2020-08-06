package httputil

import (
	"fmt"
	"net/url"
	"testing"
)

func TestDoRaw(t *testing.T) {
	header := make(map[string]string, 0)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	ret, e := DoRaw("POST", "http://127.0.0.1:9898/demo/test", "param=efefe", header)
	fmt.Printf("ret:%s,e:%v", ret, e)
}

func TestPost(t *testing.T) {
	u := url.Values{}
	u.Set("param", "fefe")
	header := make(map[string]string, 0)
	header["Content-Type"] = "application/x-www-form-urlencoded"
	ret, e := Post("http://127.0.0.1:9898/demo/test", u, header)
	fmt.Printf("ret:%s,e:%v", ret, e)
}

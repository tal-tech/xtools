// +build noflag

/*===============================================================
*   Copyright (C) 2019 All rights reserved.
*
*   FileName：config_test.go
*   Author：WuGuoFu
*   Date： 2019-12-05
*   Description：
*
================================================================*/
package confutil

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type IniStruct struct {
	Max     int           `ini:"max"`
	Port    string        `ini:"port"`
	Rate    float32       `ini:"rate"`
	Hosts   []string      `ini:"hosts"`
	Timeout time.Duration `ini:"timeout"`
}

func TestGoConfig(t *testing.T) {
	InitConfig(os.Getenv("GOPATH") + "/src/github.com/tal-tech/xtools/confutil/conf/conf.ini")
	assert.Equal(t, "goconfig", GetConf("goconfig", "name"))
	assert.Equal(t, GetConfDefault("goconfig", "name", ""), "goconfig")
	assert.Equal(t, GetConfDefault("goconfig", "default", ""), "")
	assert.Equal(t, strings.Join(GetConfs("goconfig", "hosts"), ","), "127.0.0.1,127.0.0.2,127.0.0.3")
	val := GetConfStringMap("goconfigStringMap")
	assert.Equal(t, val["name"], "goconfig")
	assert.Equal(t, val["host"], "127.0.0.1")
	mval := GetConfArrayMap("goconfigArrayMap")
	assert.Equal(t, mval["name"][0], "goconfig1")
	assert.Equal(t, mval["name"][1], "goconfig2")

	//struct
	var s IniStruct
	ConfMapToStruct("goconfigObject", &s)
	assert.Equal(t, s.Max, 101)
	assert.Equal(t, s.Port, "9099")
	assert.Equal(t, s.Rate, float32(1.01))
	assert.Equal(t, s.Hosts[0], "127.0.0.1")
	assert.Equal(t, s.Hosts[1], "127.0.0.2")
	assert.Equal(t, s.Timeout, time.Second*5)

}

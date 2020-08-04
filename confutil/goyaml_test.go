/*===============================================================
*   Copyright (C) 2019 All rights reserved.
*
*   FileName：goyaml_test.go
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

	"git.100tal.com/wangxiao_go_lib/xesTools/flagutil"
	"github.com/stretchr/testify/assert"
)

type YamlStruct struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func TestGoYaml(t *testing.T) {
	flagutil.SetConfig("conf/conf.yaml")
	SetConfPathPrefix(os.Getenv("GOPATH") + "/src/git.100tal.com/wangxiao_go_lib/xesTools/confutil")
	assert.Equal(t, GetConf("goyaml", "name"), "goyaml")
	assert.Equal(t, GetConfDefault("goyaml", "name", ""), "goyaml")
	assert.Equal(t, GetConfDefault("goyaml", "default", ""), "")
	assert.Equal(t, strings.Join(GetConfs("goyaml", "hosts"), ","), "127.0.0.1,127.0.0.2,127.0.0.3")
	val := GetConfStringMap("goyamlStringMap")
	assert.Equal(t, val["name"], "goyaml")
	assert.Equal(t, val["host"], "127.0.0.1")
	mval := GetConfArrayMap("goyamlArrayMap")
	assert.Equal(t, mval["name"][0], "goyaml1")
	assert.Equal(t, mval["name"][1], "goyaml2")

	//struct
	ob := []YamlStruct{}
	err := ConfMapToStruct("goyamlObject", &ob)
	if err != nil {
		t.Errorf("GoYaml failed,err:%v", err)
		return
	}
	assert.Equal(t, ob[0].Host, "public1")
	assert.Equal(t, ob[0].Port, "9092")
	assert.Equal(t, ob[1].Host, "public2")
	assert.Equal(t, ob[1].Port, "9093")
	assert.Equal(t, ob[2].Host, "public3")
	assert.Equal(t, ob[2].Port, "9094")
}

type YamlObjectTest struct {
	Object []YamlStruct `yaml:"goyamlObject"`
}

func TestGoYamlObject(t *testing.T) {
	SetConfPathPrefix(os.Getenv("GOPATH") + "/src/git.100tal.com/wangxiao_go_lib/xesTools/confutil")
	cfg, err := Load("conf/object.yaml")
	if err != nil {
		t.Errorf("GoYaml Load failed,err:%v", err)
	}
	ob := YamlObjectTest{}
	err = cfg.GetSectionObject("", &ob)
	if err != nil {
		t.Errorf("GoYaml failed,err:%v", err)
		return
	}
	assert.Equal(t, ob.Object[0].Host, "public1")
	assert.Equal(t, ob.Object[0].Port, "9092")
	assert.Equal(t, ob.Object[1].Host, "public2")
	assert.Equal(t, ob.Object[1].Port, "9093")
	assert.Equal(t, ob.Object[2].Host, "public3")
	assert.Equal(t, ob.Object[2].Port, "9094")

}

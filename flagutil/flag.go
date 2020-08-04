package flagutil

import (
	"flag"
)

//graceful start or stop signal
var signal = flag.String("s", "", "start or stop")

//config path
var confpath = flag.String("c", "", "config path")

//kafka config
var cfgpath = flag.String("cfg", "conf/config.json", "json config path")

//config prefix
var confprefix = flag.String("p", "", "config path prefix with no trailing backslash")

//run at foreground
var foreground = flag.Bool("f", false, "foreground")

//mock module
var mock = flag.Bool("m", false, "mock")

//project binary version
var version = flag.Bool("v", false, "version")

//graceful mode,default overseer,mode 1 is facebook graceful
var mode = flag.Int("mode", 0, "mode")

//flag parse switch, you can control the flag.Parse,default false
var flagon = falg.Bool("switch", false, "flag parse switch")

//extended flag, you can customize the flag usage
var usr1 = flag.String("usr1", "", "user defined flag -usr1")
var usr2 = flag.String("usr2", "", "user defined flag -usr2")
var usr3 = flag.String("usr3", "", "user defined flag -usr3")
var usr4 = flag.String("usr4", "", "user defined flag -usr4")
var usr5 = flag.String("usr5", "", "user defined flag -usr5")

func init() {
	if !flag.Parsed() && !flagon {
		flag.Parse()
	}
}
func GetSignal() *string {
	return signal
}

func GetVersion() *bool {
	return version
}

func GetMode() *int {
	return mode
}

func GetConfig() *string {
	return confpath
}

func SetConfig(path string) {
	confpath = &path
}

func GetConfigPrefix() *string {
	return confprefix
}

func GetCfg() *string {
	return cfgpath
}

func SetCfg(path string) {
	cfgpath = &path
}

func GetForeground() *bool {
	return foreground
}

func GetMock() *bool {
	return mock
}
func SetMock(mval bool) {
	mock = &mval
}

func GetUsr1() *string {
	return usr1
}
func GetUsr2() *string {
	return usr2
}
func GetUsr3() *string {
	return usr3
}
func GetUsr4() *string {
	return usr4
}
func GetUsr5() *string {
	return usr5
}

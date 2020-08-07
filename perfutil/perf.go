package perfutil

import (
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
	logger "github.com/tal-tech/loggerX"
)

const FALCON_SUM_POINT_TYPE = "1"
const FALCON_PERCENTILE_POINT_TYPE = "2"

var hostname string

var servername string

var level string

var timedebug string

var lock sync.Mutex

var dataStore []string

var toWrite bool = false

var levelmap map[string]int = map[string]int{
	"D": 1,
	"I": 2,
	"W": 3,
	"E": 4,
	"F": 5,
	"T": 6,
}

func CountD(metric string) {
	counterX(FALCON_SUM_POINT_TYPE, "D", metric)
}

func CountI(metric string) {
	counterX(FALCON_SUM_POINT_TYPE, "I", metric)
}

func CountIx(metric string, v ...int64) {
	counterX(FALCON_SUM_POINT_TYPE, "I", metric, v...)
}

func CountW(metric string) {
	counterX(FALCON_SUM_POINT_TYPE, "W", metric)
}

func CountE(metric string) {
	counterX(FALCON_SUM_POINT_TYPE, "E", metric)
}

func CountC(metric string) {
	counterX(FALCON_SUM_POINT_TYPE, "C", metric)
}

func AutoElapsed(metric string, start time.Time) {
	cost := time.Now().UnixNano() - start.UnixNano()
	counterX(FALCON_PERCENTILE_POINT_TYPE, "T", metric, cost)
}

func AutoElapsedDebug(metric string) func() {
	if timedebug != "true" {
		return func() {}
	}
	start := time.Now()
	return func() {
		cost := time.Now().UnixNano() - start.UnixNano()
		counterX(FALCON_PERCENTILE_POINT_TYPE, "T", metric, cost)
	}
}

func counterX(typ, lvl, metric string, v ...int64) {
	if levelmap[lvl] < levelmap[level] {
		return
	}
	if !toWrite {
		return
	}
	if metric == "" {
		metric = "NometricError"
	} else if strings.Contains(metric, " ") {
		metric = strings.Replace(metric, " ", "", -1)
	}
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	if servername == "" {
		servername = getServername()
	}
	var value string
	if len(v) > 0 {
		value = cast.ToString(v[0])
	} else {
		value = "1"
	}
	var tags string
	tags = "s=" + servername + ",l=" + lvl

	outSlice := make([]string, 0, 5)
	outSlice = append(outSlice, typ)
	outSlice = append(outSlice, value)
	outSlice = append(outSlice, metric)
	outSlice = append(outSlice, hostname)
	outSlice = append(outSlice, tags)
	output := strings.Join(outSlice, " ")
	output = Filter(output)
	store(output)
	return
}

var defaultReplacer *strings.Replacer

var conn net.Conn

var perfhost string

type perfConfig struct {
	Host      string
	Level     string
	TimeDebug string
}

func NewperfConfig() *perfConfig {
	config := new(perfConfig)
	config.Host = "127.0.0.1:13333"
	config.Level = "I"
	config.TimeDebug = "false"
	return config
}

func InitPerfWithConfig(config *perfConfig) {
	level = config.Level
	perfhost = config.Host
	timedebug = config.TimeDebug
	toWrite = true
	dataStore = make([]string, 0, 1024)
	servername = getServername()
	hostname, _ = os.Hostname()
	defaultReplacer = strings.NewReplacer("\t", "", "\r", "", "\n", "")
	conn, _ = net.Dial("udp", perfhost)
	go deal()
	//RegegisterPerfPlugin 注册之后,perf包内不可打ERROR级log
	logger.RegisterPerfPlugin(CountE)
}

func getServername() string {
	arg0 := os.Args[0]
	items := strings.Split(arg0, "/")
	return items[len(items)-1]
}

func Filter(msg string, r ...string) string {
	replacer := defaultReplacer
	if len(r) > 0 {
		replacer = strings.NewReplacer("\t", r[0], "\r", r[0], "\n", r[0])
	}
	return replacer.Replace(msg)
}

func store(data string) {
	lock.Lock()
	defer lock.Unlock()
	dataStore = append(dataStore, data)
	return
}

func deal() {
	timer := time.NewTicker(time.Millisecond * 10)
	for {
		select {
		case <-timer.C:
			lock.Lock()
			replica := dataStore[:]
			dataStore = make([]string, 0, len(replica)*2)
			lock.Unlock()
			if len(replica) > 0 {
				go toPerfProxy(replica)
			}
		}
	}
	return
}

func toPerfProxy(dataslice []string) {
	if conn == nil {
		var err error
		conn, err = net.Dial("udp", perfhost)
		if err != nil {
			logger.W("PerfInitServer", "Init error:%v", err)
			return
		}
	}
	var data string
	length := len(dataslice)
	for i := 0; ; i++ {
		if (i+1)*100 >= length {
			data = strings.Join(dataslice[i*100:length], "\t")
			break
		} else {
			data = strings.Join(dataslice[i*100:i*100+100], "\t")
		}
		conn.Write([]byte(data + "\n"))
	}
	conn.Write([]byte(data + "\n"))
	return
}

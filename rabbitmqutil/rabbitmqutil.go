/**
 *Created by He Haitao at 2019/10/25 10:41 上午
 */
package rabbitmqutil

import (
	"errors"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/tal-tech/connPool"
	logger "github.com/tal-tech/loggerX"
	"github.com/tal-tech/xtools/confutil"

	"github.com/henrylee2cn/teleport/socket"
)

var uSocket string
var tcpSockets []string
var pool *connPool.ConnPool

var once sync.Once

var readerPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return &buf
	},
}

func InitRabbitmqutil(confMap map[string][]string) {
	if unixs, ok := confMap["unix"]; ok {
		if len(unixs) > 0 {
			uSocket = confMap["unix"][0]
		}
	}
	if hosts, ok := confMap["host"]; ok {
		tcpSockets = hosts
	}
	return
}

func _init() {
	confMap := confutil.GetConfArrayMap("RabbitmqProxy")
	if unixs, ok := confMap["unix"]; ok {
		if len(unixs) > 0 {
			uSocket = confMap["unix"][0]
		}
	}
	if hosts, ok := confMap["host"]; ok {
		tcpSockets = hosts
	}
	pool = connPool.NewConnPool(&connPool.Options{
		Dialer:             dial,
		PoolSize:           200,
		PoolTimeout:        time.Second * 250,
		IdleTimeout:        time.Second * 100,
		IdleCheckFrequency: time.Millisecond * 500,
	})
}

func selectSocket() (conn *connPool.Conn, err error) {
	if pool == nil {
		once.Do(_init)
	}
	conn, _, err = connPool.GetConn(pool)
	return
}

func dial() (conn net.Conn, err error) {
	conn, err = net.DialTimeout("unix", uSocket, time.Millisecond*200)
	for cnt := 3; err != nil && cnt > 0; cnt-- {
		if len(tcpSockets) > 0 {
			index := rand.Intn(len(tcpSockets))
			conn, err = net.DialTimeout("tcp", tcpSockets[index], time.Millisecond*200)
		} else {
			err = logger.NewError("No unix or hosts found")
		}
	}
	if err != nil {
		logger.E("Socket", "Socket No Available,err %v", err)
		return nil, err
	}
	return
}
func isBadConn(err error) bool {
	if _, ok := err.(*net.OpError); ok {
		return true
	}
	return false
}
func sendBySocket(msg []byte) (back string, err error) {
	var conn *connPool.Conn
	conn, err = selectSocket()
	if err != nil {
		return "", err
	}
	defer func() {
		_, e := conn.ReleaseConn(pool, err, isBadConn)
		if e != nil {
			logger.E("ReleaseConn", "Failed,err=%v", e)
		}
	}()
	cn := conn.GetConn()
	s := socket.GetSocket(cn)
	defer func() {
		s.Reset(nil)
		s.Close()
	}()
	message := socket.GetMessage(socket.WithSetMeta("X-MQTYPE", "rabbitmq"))
	defer socket.PutMessage(message)
	//message.Reset()
	/*
		message.SetMtype(0)
		message.SetBodyCodec('r')
		message.SetSeq("1")
		message.SetUri("proxy")
	*/
	message.SetBody(msg)
	err = s.WriteMessage(message)
	if err != nil {
		return "", err
	}
	message.Reset(socket.WithBody(readerPool.Get().(*[]byte)))
	err = s.ReadMessage(message)
	if err != nil {
		return "", err
	}
	body, _ := message.MarshalBody()
	back = string(body)
	bb := message.Body().(*[]byte)
	*bb = (*bb)[:0]
	readerPool.Put(bb)
	logger.D("SendBySocket", "msg %s,back %s", msg, back)
	return back, nil
}

var bytePool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, 4096)
		return buf
	},
}

func Send2Proxy(exchangeName, routingKey string, msg []byte, expire string) error {
	var err error
	if exchangeName == "" || strings.Contains(exchangeName, " ") {
		err = errors.New("exchange name can not be empty or contain empty")
		logger.E("ExchangeNameError", "ExchangeName err:%v", err)
		return err
	}
	if routingKey == "" || strings.Contains(routingKey, " ") {
		err = errors.New("routing key can not be empty or contain empty")
		logger.E("RoutingKeyError", "RoutingKey err:%v", err)
		return err
	}

	if expire == "" {
		expire = "-1"
	}

	line := bytePool.Get().([]byte)
	line = append(line, []byte(exchangeName)...)
	line = append(line, ' ')
	line = append(line, []byte(routingKey)...)
	line = append(line, ' ')
	line = append(line, []byte(expire)...)
	line = append(line, ' ')
	line = append(line, msg...)

	back, err := sendBySocket(line)
	if !strings.Contains(back, "OK") {
		logger.E("SocketError", "InvalidSend back:%s msg:%v", back, msg)
	}
	for cnt := 3; !strings.Contains(back, "OK") && cnt > 0; cnt-- {
		back, err = sendBySocket(line)
		if !strings.Contains(back, "OK") {
			logger.E("SocketErrorRetry", "SendMsg err:%v msg:%v back:%s", err, msg, back)
		}
	}
	line = line[:0]
	bytePool.Put(line)

	return err
}

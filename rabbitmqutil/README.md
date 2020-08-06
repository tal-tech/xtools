### rabbitmqutil
-----
#### 使用方法
```go
package main
 
import (
    "fmt"
    "time"
 
    "github.com/tal-tech/xtools/rabbitmqutil"
)

func main() {
    t := time.Tick(5 * time.Second)
    count := 0
    for {
        select {
        case <-t:
            count++
            s := fmt.Sprintf("rabbitmq %d", count)
            err := rabbitmqutil.Send2Proxy("test", "test", []byte(s))
            if err != nil {
                fmt.Println(err)
            }
            err = rabbitmqutil.Send2Proxy("test", "test1", []byte(s))
            if err != nil {
                fmt.Println(err)
            }
            continue
        }
    }
}
```
#### 使用配置
```shell
[RabbitmqProxy]
unix=/home/www/pan.xesv5.com/pan.sock   //pan的sock文件地址
host=localhost:9999  //pan的ip:port地址
```

#### 注意事项
注意go.mod文件中替换包
```shell
replace github.com/henrylee2cn/teleport v5.0.0+incompatible => github.com/hhtlxhhxy/github.com_henrylee2cn_teleport v1.0.0

或

replace github.com/henrylee2cn/teleport v0.0.0 => github.com/hhtlxhhxy/github.com_henrylee2cn_teleport v1.0.0
```


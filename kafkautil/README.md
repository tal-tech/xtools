### kafkautil
-----
#### 使用方法
```go
package main
 
import (
    "fmt"
    "time"
 
    "github.com/tal-tech/xtools/kafkautil"

    "github.com/spf13/cast"
)
 
func main() {
    t := time.Tick(5 * time.Second)
    count := 0
    for {
        select {
        case <-t:
            count++
            err := kafkautil.Send2Proxy("test", []byte("kafka "+cast.ToString(count)))
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
[KafkaProxy]
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



## traceutil
### 组件说明
traceutil是对接开源链路跟踪组件zipkin的工具类库,可以控制采样粒度对服务内外进行链路跟踪,定位链路问题。
### 使用方法
```go
    //生成span对象 
    //参数1 context为请求上下文,链路共用,存储链路信息
    //参数2 submitCourseWareTest为节点名称
	span, newCtx := traceutil.Trace(context, "submitCourseWareTest")     
	if span != nil {
        //节点参数注入 可以从链路跟踪界面查看节点数据
		span.Tag("stuId", cast.ToString(stuId))
		span.Tag("liveId", cast.ToString(liveId))
		span.Tag("testPlan", cast.ToString(testPlan))
		span.Tag("packageId", cast.ToString(packageId))
        //切记要回收span
		defer span.Finish()
	}
```
API(gaea)\RPC(odin)框架已集成trace context传递，业务代码只需参考以上使用方法

### 组件配置
```golang
[Trace]
//kafka集群节点信息
kafka=public-log-kafka-1:9092 public-log-kafka-2:9092 public-log-kafka-3:9092 public-log-kafka-4:9092 public-log-kafka-5:9092 public-log-kafka-6:9092
//采样比例配置
sample=0.001
//服务名称
servername=xesApi
```

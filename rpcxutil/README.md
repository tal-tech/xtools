
conf.ini 配置

[SD]
addrs=10.99.2.162:12181  //配置中心地址列表
basePath=/xes_science //根目录
testingAddrs=10.10.135.111 //灰度测试机ip列表

example：

baseid := serviceUtils.BaseId{
		StuID:1,
		StuCouID:1,
		ClassID:1,
		TeamID:1,
		LiveID:1,
	}
	saveEnergyArgs := &protocol.SaveH5EnergyArgs{
		BaseId:       baseid,
		TestId:       1,
		SrcType:      1,
		Gold:         1,
		IsSubmit:     1,
		WrongTestNum: 1,
		RightTestNum: 1,
	}
	saveEnergyReply := &protocol.CommonReply{}

	err := rpcxutil.Call(context.Background(),"energygold","SaveH5EnergyAndGold",saveEnergyArgs,saveEnergyReply)
	if err != nil{
		fmt.Println("call err:",err)
		return
	}
	fmt.Println(saveEnergyReply)
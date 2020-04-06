package main

import (
	"github.com/zgfzgf/rabbitmq/mqengine"
	"os"
)

func main() {
	mqengine.GetConfig("./conf.json")
	//byte, _ := json.Marshal(conf)
	logger := mqengine.GetLog()
	logger.Info("log 初始化成功")
	//logger.Info("see:",
	//	zap.ByteString("conf", byte))
	StartEngine()

	c := make(chan os.Signal)
	<-c
}

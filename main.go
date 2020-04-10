package main

import (
	"context"
	"github.com/zgfzgf/rabbitmq/mqengine"
	"os"
	"os/signal"
	"sync"
)

func main() {
	mqengine.GetConfig("./conf.json")
	logger := mqengine.GetLog()
	logger.Info("log 初始化成功")
	mqengine.GetRabbitMqConn()
	defer mqengine.CloseRabbitMqConn()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	StartEngine(ctx, &wg)

	c := make(chan os.Signal)
	signal.Notify(c)
	<-c
	logger.Info("begin cancel")
	cancel()
	wg.Wait()

	logger.Info("Success End Main Func")
}

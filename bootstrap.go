package main

import "github.com/zgfzgf/rabbitmq/mqengine"

func StartEngine() {
	//productIds := [2]string{"aaa", "bbb"}
	productIds := [1]string{"aaa"}
	for _, productId := range productIds {
		process := NewProcess(productId)
		readerMq := mqengine.NewReaderMQ(productId)
		storeMq := mqengine.NewStoreMQ(productId)
		infoMq := mqengine.NewInfoMQ(productId)
		engine := mqengine.NewEngine(productId, process, readerMq, storeMq, infoMq)
		engine.Start()
	}
}

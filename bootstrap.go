package main

import (
	"context"
	"github.com/zgfzgf/rabbitmq/mqengine"
	"sync"
)

func StartEngine(ctx context.Context, wg *sync.WaitGroup) {
	//productIds := [2]string{"aaa", "bbb"}
	productIds := [1]string{"aaa"}
	for _, productId := range productIds {
		process := NewProcess(productId)
		readerMq := mqengine.NewReaderMQ(productId)
		storeMq := mqengine.NewStoreMQ(productId)
		infoMq := mqengine.NewInfoMQ(productId)
		snapshotStore := mqengine.NewRedisSnapshotStore(productId)
		engine := mqengine.NewEngine(productId, process, readerMq, storeMq, infoMq, snapshotStore)
		engine.Start(ctx, wg)
	}
}

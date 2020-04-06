package main

import (
	"github.com/zgfzgf/rabbitmq/mqengine"
	"time"
)

type Process struct {
	productId string
}

func NewProcess(productId string) *Process {
	process := &Process{
		productId: productId,
	}
	return process
}

func (p *Process) OnStart(message *mqengine.Message, num int) mqengine.Log {
	return &LogStart{
		Base: Base{Status: mqengine.MessageStatusStart,
			ProductId:     p.productId,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		Num: num,
	}
}
func (p *Process) OnProccess(message *mqengine.Message) (stores []mqengine.Log, infos []mqengine.Log) {
	log := NewLogProccessStore(message, p.productId, "store1")
	stores = append(stores, log)
	log = NewLogProccessStore(message, p.productId, "store2")
	stores = append(stores, log)
	log = NewLogProccessStore(message, p.productId, "store3")
	stores = append(stores, log)

	info := NewLogProccessInfo(message, p.productId, "info1")
	infos = append(infos, info)
	info = NewLogProccessInfo(message, p.productId, "info2")
	infos = append(infos, info)
	return stores, infos
}

func (p *Process) OnEnd(message *mqengine.Message, num int) mqengine.Log {
	return &LogEnd{
		Base: Base{Status: mqengine.MessageStatusEnd,
			ProductId:     p.productId,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		Num:                 num,
		CorrelationDelivery: message.CorrelationDelivery,
	}
}

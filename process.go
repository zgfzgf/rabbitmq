package main

import (
	"encoding/json"
	"github.com/zgfzgf/rabbitmq/mqengine"
	"go.uber.org/zap"
	"time"
)

type Process struct {
	productId string
	processId int64
}

type Snapshot struct {
	ProductId string
	ProcessId int64
}

func NewProcess(productId string) *Process {
	process := &Process{
		productId: productId,
	}
	return process
}

func (p *Process) Work(message *mqengine.Message) (stores []mqengine.Log, infos []mqengine.Log, id interface{}) {
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
	return stores, infos, p.nextId()
}
func (p *Process) Start(message *mqengine.Message, workId interface{}, num int) mqengine.Log {
	return &LogStart{
		Base: Base{Status: mqengine.MessageStatusStart,
			ProductId:     p.productId,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		TransId: workId,
		Num:     num,
	}
}
func (p *Process) End(message *mqengine.Message, workId interface{}, num int) mqengine.Log {
	return &LogEnd{
		Base: Base{Status: mqengine.MessageStatusEnd,
			ProductId: p.productId,

			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		TransId:             workId,
		Num:                 num,
		CorrelationDelivery: message.CorrelationDelivery,
	}
}

func (p *Process) Snapshot() []byte {
	store := &Snapshot{ProductId: p.productId, ProcessId: p.processId}
	buf, err := json.Marshal(store)
	if err != nil {
		return nil
	}
	return buf
}

func (p *Process) Restore(snapshot []byte) {
	var store Snapshot
	json.Unmarshal(snapshot, &store)
	p.processId = store.ProcessId
	logger.Info("Restore", zap.Int64("processId", p.processId))
}

func (p *Process) CmpTransId(newId interface{}, oldId interface{}) bool {
	if newId.(int64) >= oldId.(int64) {
		return true
	}
	return false
}

func (p *Process) NotSnapShot(newId interface{}, oldId interface{}, delta int) bool {
	return !p.CmpTransId(newId, oldId.(int64)+int64(delta))
}

func (p *Process) nextId() interface{} {
	p.processId++
	return p.processId
}
func (p *Process) GetTransId() interface{} {
	return p.processId
}

package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"github.com/zgfzgf/rabbitmq/mqengine"
	"time"
)

type Base struct {
	Status        mqengine.MessageStatus
	ProductId     string
	CorrelationId string
	ReplyTo       string
	MessageId     string
	Time          time.Time
}

type LogStart struct {
	Base
	Num int
}

func (l *LogStart) GetMessage() *mqengine.Message {
	byte, _ := json.Marshal(l)
	return &mqengine.Message{
		Status:        l.Status,
		Body:          byte,
		CorrelationId: l.CorrelationId,
		ReplyTo:       l.ReplyTo,
		MessageId:     l.MessageId,
		Timestamp:     l.Time,
	}
}

type LogEnd struct {
	Base
	Num                 int
	CorrelationDelivery *amqp.Delivery
}

func (l *LogEnd) GetMessage() *mqengine.Message {
	byte, _ := json.Marshal(l)
	return &mqengine.Message{
		Status:              l.Status,
		Body:                byte,
		CorrelationId:       l.CorrelationId,
		ReplyTo:             l.ReplyTo,
		MessageId:           l.MessageId,
		Timestamp:           l.Time,
		CorrelationDelivery: l.CorrelationDelivery,
	}
}

type LogProccessStore struct {
	Base
	Store string
}

func (l *LogProccessStore) GetMessage() *mqengine.Message {
	byte, _ := json.Marshal(l)
	return &mqengine.Message{
		Status:        l.Status,
		Body:          byte,
		CorrelationId: l.CorrelationId,
		ReplyTo:       l.ReplyTo,
		MessageId:     l.MessageId,
		Timestamp:     l.Time,
	}
}
func NewLogProccessStore(message *mqengine.Message, productId, store string) *LogProccessStore {
	return &LogProccessStore{
		Base: Base{Status: mqengine.MessageStatusProcess,
			ProductId:     productId,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		Store: store,
	}
}

type LogProccessInfo struct {
	Base
	Info string
}

func (l *LogProccessInfo) GetMessage() *mqengine.Message {
	byte, _ := json.Marshal(l)
	return &mqengine.Message{
		Status:        l.Status,
		Body:          byte,
		CorrelationId: l.CorrelationId,
		ReplyTo:       l.ReplyTo,
		MessageId:     l.MessageId,
		Timestamp:     l.Time,
	}
}
func NewLogProccessInfo(message *mqengine.Message, productId, info string) *LogProccessInfo {
	return &LogProccessInfo{
		Base: Base{Status: mqengine.MessageStatusProcess,
			ProductId:     productId,
			CorrelationId: message.CorrelationId,
			ReplyTo:       message.ReplyTo,
			MessageId:     message.MessageId,
			Time:          time.Now()},
		Info: info,
	}
}

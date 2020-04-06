package mqengine

import (
	"github.com/streadway/amqp"
)

func NewMessage(delivery *amqp.Delivery) *Message {
	return &Message{
		Body:                delivery.Body,
		Priority:            delivery.Priority,
		CorrelationId:       delivery.CorrelationId,
		ReplyTo:             delivery.ReplyTo,
		MessageId:           delivery.MessageId,
		Timestamp:           delivery.Timestamp,
		Type:                delivery.Type,
		UserId:              delivery.UserId,
		AppId:               delivery.AppId,
		CorrelationDelivery: delivery,
	}
}

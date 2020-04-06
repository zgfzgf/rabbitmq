package mqengine

import (
	"github.com/streadway/amqp"
	"time"
)

type Message struct {
	Body                []byte
	Status              MessageStatus
	Commit              bool
	Priority            uint8     // queue implementation use - 0 to 9
	CorrelationId       string    // application use - correlation identifier
	ReplyTo             string    // application use - address to to reply to (ex: RPC)
	MessageId           string    // application use - message identifier
	Timestamp           time.Time // application use - message timestamp
	Type                string    // application use - message type name
	UserId              string    // application use - creating user - should be authenticated user
	AppId               string    // application use - creating application id
	CorrelationDelivery *amqp.Delivery
}

type MessageStatus string

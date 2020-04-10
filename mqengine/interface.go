package mqengine

//  生产者接口
type Producer interface {
	Store() (<-chan *Message, chan<- *Message)
}

//  消费者接口
type Receiver interface {
	Reader() chan<- *Message
}

//  消费者处理结果接口
type Log interface {
	GetMessage() *Message
}

//  消息处理接口
type Proccess interface {
	OnStart(*Message, int) Log
	OnProccess(*Message) ([]Log, []Log)
	OnEnd(*Message, int) Log
}

//  发送消息处理接口
type Send interface {
	SendChan() <-chan *Message
	OnProccess(*Message)
}

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
	Start(*Message, interface{}, int) Log
	Work(*Message) ([]Log, []Log, interface{})
	End(*Message, interface{}, int) Log
	NotSnapShot(interface{}, interface{}, int) bool
	CmpTransId(interface{}, interface{}) bool
	GetTransId() interface{}
	Snapshot() []byte
	Restore([]byte)
}

//  发送消息处理接口
type Send interface {
	SendChan() <-chan *Message
	OnProccess(*Message)
}

type SnapshotStore interface {
	// 保存快照
	Store([]byte) error
	// 获取最后一次快照
	GetLatest() ([]byte, error)
}

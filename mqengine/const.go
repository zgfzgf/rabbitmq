package mqengine

const (
	LevelDebug int = 1
	LevelInfo  int = 2
	// LevelWarn int = 3 default
)

const (
	//  开始状态
	MessageStatusStart = MessageStatus("start")
	//  处理状态
	MessageStatusProcess = MessageStatus("process")
	//  结束状态
	MessageStatusEnd = MessageStatus("end")
)

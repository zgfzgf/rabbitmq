package mqengine

import "go.uber.org/zap"

type Engine struct {
	// productId是一个engine的唯一标识，每个product都会对应一个engine
	productId string
	// engine持有的处理
	proccess Proccess
	// 用于读取消息
	readerHandle *RabbitMq
	readerChan   chan *Message
	// 用于保存回复消息
	storeHandle *RabbitMq
	storeChan   chan *Message
	ackChan     chan *Message
	// 用于保存日志清息
	infoHandle *RabbitMq
	infoChan   chan *Message

	logStoreChan chan Log
	logInfoChan  chan Log
}

func NewEngine(productId string, proccess Proccess, reader, store, info *RabbitMq) *Engine {
	e := &Engine{
		productId:    productId,
		proccess:     proccess,
		readerHandle: reader,
		readerChan:   make(chan *Message, config.ChanNum.Reader),
		storeHandle:  store,
		storeChan:    make(chan *Message, config.ChanNum.Store),
		ackChan:      make(chan *Message, config.ChanNum.Ack),
		infoHandle:   info,
		infoChan:    make(chan *Message, config.ChanNum.Info),
		logStoreChan: make(chan Log, config.ChanNum.LogStore),
		logInfoChan:  make(chan Log, config.ChanNum.LogInfo),
	}
	return e
}

func (e *Engine) Start() {
	e.readerHandle.RegisterReceiver(NewReader(e))
	go e.readerHandle.Start()
	e.storeHandle.RegisterProducer(NewStore(e))
	go e.storeHandle.Start()
	e.infoHandle.RegisterProducer(NewInfo(e))
	go e.infoHandle.Start()
	go e.runApplier()
	go e.runCommitter()
}

// 从本地队列获取消息
func (e *Engine) runApplier() {
	for {
		select {
		case message := <-e.readerChan:
			storeLog, infoLog := e.proccess.OnProccess(message)
			if len(storeLog) > 0 {
				e.logStoreChan <- e.proccess.OnStart(message, len(storeLog))
				for _, log := range storeLog {
					logger.Debug("storelog",
						zap.String("correlationId",log.GetMessage().CorrelationId),
						zap.ByteString("body",log.GetMessage().Body))
					e.logStoreChan <- log
				}
				e.logStoreChan <- e.proccess.OnEnd(message, len(storeLog))
			}
			for _, log := range infoLog {
				logger.Debug("infoLog",
					zap.String("correlationId",log.GetMessage().CorrelationId),
					zap.ByteString("body",log.GetMessage().Body))
				e.logInfoChan <- log
			}
		}
	}
}

//  将消费者产生的log进行持久化
func (e *Engine) runCommitter() {
	for {
		select {
		case logStore := <-e.logStoreChan:
			logger.Debug("logStore",
				zap.String("correlationId",logStore.GetMessage().CorrelationId),
				)
			e.storeChan <- logStore.GetMessage()

		case logInfo := <-e.logInfoChan:
			logger.Debug("logInfo",
				zap.String("correlationId",logInfo.GetMessage().CorrelationId))
			e.infoChan <- logInfo.GetMessage()
		case message := <-e.ackChan:
			if message.Commit {
				message.CorrelationDelivery.Ack(false)
			} else {
				logger.Info("to do")
			}
		}
	}
}

type Reader struct {
	engine *Engine
}

func NewReader(e *Engine) *Reader {
	return &Reader{
		engine: e}
}
func (r *Reader) Consumer() chan<- *Message {
	return r.engine.readerChan
}

type Store struct {
	engine *Engine
}

func NewStore(e *Engine) *Store {
	return &Store{
		engine: e}
}
func (s *Store) Store() (<-chan *Message, chan<- *Message) {
	return s.engine.storeChan, s.engine.ackChan
}

type Info struct {
	engine *Engine
}

func NewInfo(e *Engine) *Info {
	return &Info{
		engine: e}
}
func (i *Info) Store() (<-chan *Message, chan<- *Message) {
	return i.engine.infoChan, i.engine.ackChan
}

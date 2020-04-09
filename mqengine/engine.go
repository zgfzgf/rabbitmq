package mqengine

import (
	"context"
	"go.uber.org/zap"
	"sync"
)

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
		infoChan:     make(chan *Message, config.ChanNum.Info),
		logStoreChan: make(chan Log, config.ChanNum.LogStore),
		logInfoChan:  make(chan Log, config.ChanNum.LogInfo),
	}
	return e
}

func (e *Engine) Start(ctx context.Context, wg *sync.WaitGroup) {
	e.readerHandle.RegisterReceiver(NewReader(e))
	go e.readerHandle.Start()
	e.storeHandle.RegisterProducer(NewStore(e))
	go e.storeHandle.Start()
	e.infoHandle.RegisterProducer(NewInfo(e))
	go e.infoHandle.Start()
	go e.runApplier(ctx)
	go e.runLogStore()
	go e.runLogInfo()
	go e.runCommitter(wg)
}

// 从本地队列获取消息
func (e *Engine) runApplier(ctx context.Context) {
	for {
		select {
		case message := <-e.readerChan:
			storeLog, infoLog := e.proccess.OnProccess(message)
			if len(storeLog) > 0 {
				e.logStoreChan <- e.proccess.OnStart(message, len(storeLog))
				for _, log := range storeLog {
					logger.Debug("send storelog",
						zap.String("correlationId", log.GetMessage().CorrelationId),
						zap.ByteString("body", log.GetMessage().Body))
					e.logStoreChan <- log
				}
				e.logStoreChan <- e.proccess.OnEnd(message, len(storeLog))
			}
			for _, log := range infoLog {
				logger.Debug("send infoLog",
					zap.String("correlationId", log.GetMessage().CorrelationId),
					zap.ByteString("body", log.GetMessage().Body))
				e.logInfoChan <- log
			}
		case <-ctx.Done():
			close(e.logStoreChan)
			close(e.logInfoChan)
			logger.Info("close logStoreChan and logInfoChan")
			return
		}
	}
}

//  将消费者产生的log进行持久化
func (e *Engine) runLogStore() {
	for {
		select {
		case logStore, ok := <-e.logStoreChan:
			if ok {
				logger.Debug("receive logStore",
					zap.String("correlationId", logStore.GetMessage().CorrelationId))
				e.storeChan <- logStore.GetMessage()
			} else {
				close(e.storeChan)
				logger.Info("close storeChan")
				return
			}
		}
	}
}

func (e *Engine) runLogInfo() {
	for {
		select {
		case logInfo, ok := <-e.logInfoChan:
			if ok {
				logger.Debug("receive logInfo",
					zap.String("correlationId", logInfo.GetMessage().CorrelationId))
				e.infoChan <- logInfo.GetMessage()
			} else {
				close(e.infoChan)
				logger.Info("close infoChan")
				return
			}
		}
	}
}

//  将消费者产生的log进行持久化
func (e *Engine) runCommitter(wg *sync.WaitGroup) {
	defer wg.Done()
	wg.Add(1)
	for {
		select {
		case message, ok := <-e.ackChan:
			if ok {
				if message.Commit {
					message.CorrelationDelivery.Ack(false)
				} else {
					logger.Info("to do")
				}
			} else {
				logger.Info("end commit")
				return
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
	return i.engine.infoChan, nil
}

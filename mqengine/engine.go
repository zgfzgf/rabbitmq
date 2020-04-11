package mqengine

import (
	"context"
	"go.uber.org/zap"
	"sync"
	"time"
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

	closeChan chan struct{}

	// 发起snapshot请求，需要携带最后一次snapshot的offset
	snapshotReqCh chan *Snapshot

	// snapshot已经完全准备好，需要确保snapshot之前的所有数据都已经提交
	snapshotApproveReqCh chan *Snapshot

	// snapshot数据准备好并且snapshot之前的所有数据都已经提交
	snapshotCh chan *Snapshot

	// 持久化snapshot的存储方式，应该支持多种方式，如本地磁盘，redis等
	snapshotStore SnapshotStore

	transId interface{}
}

// 快照是engine在某一时候的一致性内存状态
type Snapshot struct {
	StoreSnapshot []byte
	TransId       interface{}
}

func NewEngine(productId string, proccess Proccess, reader, store, info *RabbitMq, snapshotStore SnapshotStore) *Engine {
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
		closeChan:    make(chan struct{}),

		snapshotReqCh:        make(chan *Snapshot, config.ChanNum.Snapshot),
		snapshotApproveReqCh: make(chan *Snapshot, config.ChanNum.Snapshot),
		snapshotCh:           make(chan *Snapshot, config.ChanNum.Snapshot),
		snapshotStore:        snapshotStore,
	}
	snapshot, err := snapshotStore.GetLatest()
	if err != nil {
		logger.Fatal("get latest snapshot error.", zap.Error(err))
	}
	if snapshot != nil {
		e.restore(snapshot)
	}
	return e
}

func (e *Engine) Start(ctx context.Context, wg *sync.WaitGroup) {
	go e.readerHandle.Reader(ctx, NewReader(e))
	go e.storeHandle.Store(NewStore(e))
	go e.infoHandle.Store(NewInfo(e))
	go e.runApplier()
	go e.runLogStore()
	go e.runLogInfo()
	go e.runCommitter()
	go e.runSnapshots()
	go e.close(wg)
}

func (e *Engine) close(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	defer e.store()
	<-e.closeChan
}

// 从本地队列获取消息
func (e *Engine) runApplier() {
	transId := e.transId
	for {
		select {
		case message, ok := <-e.readerChan:
			if ok {
				storeLog, infoLog, workId := e.proccess.Work(message)
				if len(storeLog) > 0 {
					e.logStoreChan <- e.proccess.Start(message, workId, len(storeLog))
					for _, log := range storeLog {
						logger.Debug("send storeLog",
							zap.String("correlationId", log.GetMessage().CorrelationId),
							zap.ByteString("body", log.GetMessage().Body))
						e.logStoreChan <- log
					}
					e.logStoreChan <- e.proccess.End(message, workId, len(storeLog))
				}
				for _, log := range infoLog {
					logger.Debug("send infoLog",
						zap.String("correlationId", log.GetMessage().CorrelationId),
						zap.ByteString("body", log.GetMessage().Body))
					e.logInfoChan <- log
				}
				transId = workId
			} else {
				close(e.logStoreChan)
				close(e.logInfoChan)
				logger.Info("close logStoreChan and logInfoChan")
				return
			}

		case snapshot := <-e.snapshotReqCh:
			// 接收到快照请求，判断是否真的需要执行快照
			if e.proccess.NotSnapShot(transId, snapshot.TransId, SnapshotDelta) {
				continue
			}

			logger.Info("should take snapshot",
				zap.String("productId", e.productId),
				zap.Int("delta", SnapshotDelta),
			)

			// 执行快照，并将快照数据写入批准chan
			snapshot.StoreSnapshot = e.proccess.Snapshot()
			snapshot.TransId = transId
			e.snapshotApproveReqCh <- snapshot
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
func (e *Engine) runCommitter() {
	var pending *Snapshot = nil
	for {
		select {
		case message, ok := <-e.ackChan:
			if ok {
				if message.Commit {
					message.CorrelationDelivery.Ack(false)
				} else {
					logger.Info("to do")
				}
				e.transId = message.TransId
				if pending != nil && e.proccess.CmpTransId(e.transId, pending.TransId) {
					e.snapshotCh <- pending
					pending = nil
				}
			} else {
				logger.Info("end commit func")
				e.closeChan <- struct{}{}
				return
			}
		case snapshot := <-e.snapshotApproveReqCh:
			// 写入的seq已经达到或者超过snapshot的seq，批准snapshot请求
			if e.proccess.CmpTransId(e.transId, snapshot.TransId) {
				e.snapshotCh <- snapshot
				pending = nil
				continue
			}
			// 当前还有未批准的snapshot，但是又有新的snapshot请求，丢弃旧的请求
			if pending != nil {
				logger.Info("discard snapshot request")
			}
			pending = snapshot

		}
	}
}

// 定时发起快照请求，同时负责持久化通过审批的快照
func (e *Engine) runSnapshots() {
	// 最后一次快照时的order orderOffset
	tranId := e.transId

	for {
		select {
		case <-time.After(SnapshotHour * time.Second):
			// make a new snapshot request
			e.snapshotReqCh <- &Snapshot{
				TransId: tranId,
			}

		case snapshot := <-e.snapshotCh:
			// store snapshot
			err := e.snapshotStore.Store(snapshot.StoreSnapshot)
			if err != nil {
				logger.Warn("store snapshot failed",
					zap.Error(err))
				continue
			}
			logger.Info("Success new snapshot stored")

			// update offset for next snapshot request
			tranId = snapshot.TransId
		}
	}
}

func (e *Engine) store() {
	logger.Info("storing")
	storeSnapshot := e.proccess.Snapshot()
	e.snapshotStore.Store(storeSnapshot)
}

func (e *Engine) restore(storeSnapshot []byte) {
	logger.Info("restoring")
	e.proccess.Restore(storeSnapshot)
	e.transId = e.proccess.GetTransId()
}

type Reader struct {
	engine *Engine
}

func NewReader(e *Engine) *Reader {
	return &Reader{engine: e}
}
func (r *Reader) Reader() chan<- *Message {
	return r.engine.readerChan
}

type Store struct {
	engine *Engine
}

func NewStore(e *Engine) *Store {
	return &Store{engine: e}
}
func (s *Store) Store() (<-chan *Message, chan<- *Message) {
	return s.engine.storeChan, s.engine.ackChan
}

type Info struct {
	engine *Engine
}

func NewInfo(e *Engine) *Info {
	return &Info{engine: e}
}
func (i *Info) Store() (<-chan *Message, chan<- *Message) {
	return i.engine.infoChan, nil
}

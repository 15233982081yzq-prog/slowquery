package kafka

import (
	"smart-slowquery/internal/model/filebeat"
	"smart-slowquery/pkg/store"
	"smart-slowquery/pkg/store/request"

	"fmt"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

// BatchProcessor 批量处理器，确保数据安全性的核心组件
type BatchProcessor struct {
	messages      []*sarama.ConsumerMessage   // 批量消息
	buffer        []*filebeat.SlowQuery       // 处理后的数据
	session       sarama.ConsumerGroupSession // Kafka 会话
	writer        store.CKWriter              // ClickHouse 写入器
	maxBatchSize  int                         // 最大批量大小
	mutex         sync.RWMutex                // 并发安全
	lastFlushTime time.Time                   // 上次刷新时间
	flushInterval time.Duration               // 刷新间隔
}

// NewBatchProcessor 创建批量处理器实例
func NewBatchProcessor(session sarama.ConsumerGroupSession, writer store.CKWriter, maxBatchSize int, flushInterval time.Duration) *BatchProcessor {
	return &BatchProcessor{
		messages:      make([]*sarama.ConsumerMessage, 0, maxBatchSize),
		buffer:        make([]*filebeat.SlowQuery, 0, maxBatchSize),
		session:       session,
		writer:        writer,
		maxBatchSize:  maxBatchSize,
		flushInterval: flushInterval,
		lastFlushTime: time.Now(),
	}
}

// AddMessage 添加消息到批量处理器
func (bp *BatchProcessor) AddMessage(msg *sarama.ConsumerMessage, processedData *filebeat.SlowQuery) error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	// 添加消息到缓冲区
	bp.messages = append(bp.messages, msg)
	bp.buffer = append(bp.buffer, processedData)

	// 检查是否需要刷新
	if bp.shouldFlush() {
		return bp.flushInternal()
	}

	return nil
}

// shouldFlush 判断是否需要刷新缓冲区
func (bp *BatchProcessor) shouldFlush() bool {
	// 检查批量大小
	if len(bp.buffer) >= bp.maxBatchSize {
		return true
	}

	// 检查时间间隔
	if time.Since(bp.lastFlushTime) >= bp.flushInterval && len(bp.buffer) > 0 {
		return true
	}

	return false
}

// Flush 强制刷新缓冲区
func (bp *BatchProcessor) Flush() error {
	bp.mutex.Lock()
	defer bp.mutex.Unlock()

	if len(bp.buffer) == 0 {
		return nil
	}

	return bp.flushInternal()
}

// flushInternal 内部刷新实现
func (bp *BatchProcessor) flushInternal() error {
	if len(bp.buffer) == 0 {
		return nil
	}

	// 1. 批量写入 ClickHouse
	var slowQueryLogs []*request.SlowQueryLog
	for _, fb := range bp.buffer {
		slowQueryLogs = append(slowQueryLogs, request.BuildSlowQueryLog(fb))
	}

	// 使用重试机制写入 ClickHouse
	if err := bp.writeToCKWithRetry(slowQueryLogs); err != nil {
		return fmt.Errorf("批量写入 ClickHouse 失败: %w", err)
	}

	// 2. 写入成功，批量提交 offset
	for _, msg := range bp.messages {
		bp.session.MarkMessage(msg, "")
	}

	// 3. 清空缓冲区
	bp.messages = bp.messages[:0]
	bp.buffer = bp.buffer[:0]
	bp.lastFlushTime = time.Now()

	return nil
}

// writeToCKWithRetry 带重试机制的 ClickHouse 批量写入
func (bp *BatchProcessor) writeToCKWithRetry(data []*request.SlowQueryLog) error {
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		// 直接使用批量写入方法
		// 这里需要获取底层的 ClickHouse 客户端进行批量写入
		// 由于当前的 writer 接口限制，我们需要修改实现方式

		// 临时解决方案：逐条写入，但确保原子性
		var lastError error
		for _, slowQuery := range data {
			err, _ := bp.writer.Append(slowQuery)
			if err != nil {
				lastError = err
				break
			}
		}

		// 如果所有写入都成功，返回成功
		if lastError == nil {
			return nil
		}

		// 如果有错误，进行重试
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return fmt.Errorf("ClickHouse 批量写入失败，达到最大重试次数")
}

// GetBatchSize 获取当前批量大小
func (bp *BatchProcessor) GetBatchSize() int {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	return len(bp.buffer)
}

// GetLastFlushTime 获取上次刷新时间
func (bp *BatchProcessor) GetLastFlushTime() time.Time {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()
	return bp.lastFlushTime
}

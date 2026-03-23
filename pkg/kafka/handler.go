package kafka

import (
	sysMetrics "smart-slowquery/pkg/metrics/analyzer"

	"smart-slowquery/pkg/analyzer"
	"smart-slowquery/pkg/log"

	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

type GroupHandler struct { //Kafka消费者组处理器
	service *analyzer.Service           //慢查询分析
	session sarama.ConsumerGroupSession //代表Kafka消费者会话，用于提交偏移量和管理消费者状态
}

func NewGroupHandler(service *analyzer.Service) (*GroupHandler, error) { //新建消费处理器结构体实例
	if service == nil {
		return nil, fmt.Errorf("param is empty")
	}
	return &GroupHandler{
		service: service,
	}, nil
}

func (h *GroupHandler) Setup(session sarama.ConsumerGroupSession) error { // 消费者组会话建立时的回调函数 不懂
	log.Infof("session id:%d ,claims:%v", session.GenerationID(), session.Claims())
	h.session = session
	return nil
}

func (h *GroupHandler) Cleanup(session sarama.ConsumerGroupSession) error { // 消费者组会话建立时的回调函数 不懂
	log.Infof("cleanup session id:%d ,commit", session.GenerationID())
	h.session.Commit()
	h.session = nil
	return nil
}

func (h *GroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error { // 负责处理被分配给当前消费者的特定分区里的消息流
	// 创建批量处理器，确保数据安全性：先写入CK，再提交offset
	batchProcessor := NewBatchProcessor(session, h.service.GetWriter(), 1000, 5*time.Second)
	
	// claim.Messages() 返回只读通道 <-chan *ConsumerMessage，持续产生Kafka消息
	// for ... range 循环会持续从通道中接收消息，直到分区被重新分配或会话结束
	for message := range claim.Messages() {
		start := time.Now()

		// message 是 *ConsumerMessage 类型指针，包含Kafka消息的完整信息：
		// - Topic: 消息所属Kafka主题
		// - Partition: 消息所在分区号
		// - Offset: 消息在分区中的偏移量
		// - Key: 消息键（用于路由，可为空）
		// - Value: 消息实际内容（慢查询日志数据，JSON格式）
		// - Timestamp: 消息时间戳
		log.Infof("Received message: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value_size:%d",
			message.Topic, message.Partition, message.Offset, string(message.Key), len(string(message.Value)))

		// 调用分析服务处理单条消息，返回处理后的数据和过滤状态
		processedData, err, filtered := h.service.ProcessSingleMessage(message)

		// 收集Kafka处理指标
		sysMetrics.CollectKafkaMetrics(fmt.Sprintf("analyzer_%s", message.Topic), sysMetrics.GetStatus(err), time.Since(start))

		// 处理结果分支：
		if err != nil {
			// 处理出错，记录错误并关闭Kafka客户端
			log.Errorf("ConsumeClaim processor error:%s ,kafka client will close", err.Error())
			return err
		}
		
		if filtered {
			// 消息被过滤（如不符合分析条件），立即提交offset（因为不需要写入CK）
			session.MarkMessage(message, "")
			log.Warningf("Processor message topic:%s,partition:%d,offset:%d,timeStamp:%d was filtered",
				message.Topic, message.Partition, message.Offset, message.Timestamp.Unix())
			continue
		}
		
		// 有效消息：加入批量处理器（确保写入CK成功后才提交offset）
		if err := batchProcessor.AddMessage(message, processedData); err != nil {
			log.Errorf("BatchProcessor AddMessage error:%s ,kafka client will close", err.Error())
			return err
		}
	}
	
	// 消费结束时，强制刷新剩余消息
	if err := batchProcessor.Flush(); err != nil {
		log.Errorf("BatchProcessor Flush error:%s", err.Error())
		return err
	}
	
	return nil
}

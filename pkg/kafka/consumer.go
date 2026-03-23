package kafka

import (
	"smart-slowquery/conf"
	"smart-slowquery/pkg/analyzer"
	"smart-slowquery/pkg/log"
	"time"

	"context"
	"fmt"

	"github.com/Shopify/sarama"
)

type Client struct {
	kfkCli     sarama.Client
	kfkCG      sarama.ConsumerGroup
	kfkHandler *GroupHandler
	cfg        *sarama.Config
	groupID    string
	output     chan string
	brokers    []string
	topics     []string
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewClient(cfg *conf.Kafka, service *analyzer.Service) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("kafka config or output channel is empty")
	}

	var (
		client  sarama.Client
		cg      sarama.ConsumerGroup
		handler *GroupHandler
		err     error
	)

	if client, err = sarama.NewClient(cfg.Brokers, GetDefaultConfig()); err != nil {
		return nil, err
	}

	if cg, err = sarama.NewConsumerGroupFromClient(cfg.GroupID, client); err != nil {
		return nil, err
	}

	if handler, err = NewGroupHandler(service); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		kfkCli:     client,
		kfkCG:      cg,
		kfkHandler: handler,
		cfg:        GetDefaultConfig(),
		brokers:    cfg.Brokers,
		groupID:    cfg.GroupID,
		topics:     cfg.Topics,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

func (c *Client) Consume() (err error) {
	// watch message
	go func() {
		for {
			if err = c.kfkCG.Consume(c.ctx, c.topics, c.kfkHandler); err != nil { // 这里是阻塞的，会使用到sarama框架，并且是循环调用
				log.Errorf("consumer error:%s", err.Error())
				// 在此处添加重试逻辑
				retryCount := 0
				maxRetries := 5
				for retryCount < maxRetries {
					log.Infof("Retrying to consume... attempt %d", retryCount+1)
					time.Sleep(time.Second * 5) // 等待一段时间再重试
					if err = c.kfkCG.Consume(c.ctx, c.topics, c.kfkHandler); err == nil {
						break
					}
					retryCount++
				}
				if retryCount == maxRetries {
					log.Errorf("Max retry attempts reached, exiting consumer loop")
					return
				}
			}
			if c.ctx.Err() != nil {
				log.Infof("ctx error:%s", c.ctx.Err().Error())
				return
			}
		}
	}()

	// watch error message
	go func() {
		for err := range c.kfkCG.Errors() {
			log.Errorf("broker:%v consumerGroup Error: %s", c.brokers, err.Error())
		}
	}()

	return nil
}

func (c *Client) Close() {
	log.Infof("close kafka consumerGroup broker:%v", c.brokers)
	c.cancel()
	c.kfkCG.Close()
	c.kfkCli.Close()
}

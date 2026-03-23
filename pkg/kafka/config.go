package kafka

import (
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

var (
	once   sync.Once
	config *sarama.Config
)

func GetDefaultConfig() *sarama.Config {
	once.Do(func() {
		config = sarama.NewConfig()
		config.Version = sarama.V2_5_0_0
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
		config.Consumer.Offsets.Initial = sarama.OffsetNewest

		// 添加心跳和会话超时配置 30s没有心跳会被踢出group
		config.Consumer.Group.Session.Timeout = 30 * time.Second
	})
	return config
}

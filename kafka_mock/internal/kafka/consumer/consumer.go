package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go-helloworld/kafka_mock/internal/processor"
	"log"
	"log/slog"
	"sync"
	"time"
)

type ConsumerGroupManager struct {
	logger         *slog.Logger
	consumerGroups map[string][]*ConsumerInstance
	wg             sync.WaitGroup
}

type ConsumerInstance struct {
	ID         string
	logger     *slog.Logger
	ctx        context.Context
	cancelFunc context.CancelFunc
	consumer   *kafka.Consumer
	processor  processor.Processor
	retryCh    chan RetryMsg
	semCh      chan struct{}
}

type RetryMsg struct {
	attempt int
	msg     kafka.Message
}

func NewConsumerGroupManager(
	logger *slog.Logger,
	brokers, groupPrefix string,
	topics []string,
	processorFactory func(topic string) processor.Processor,
) (*ConsumerGroupManager, error) {
	ctx := context.Background()

	partitionsPerTopic, partitionsErr := GetTopicPartitionCounts(brokers, topics)
	if partitionsErr != nil {
		return nil, fmt.Errorf("error getting partitions: topics: %v: error: %w", topics, partitionsErr)
	}

	manager := &ConsumerGroupManager{
		logger:         logger,
		consumerGroups: make(map[string][]*ConsumerInstance),
	}

	for topic, partitionsCount := range partitionsPerTopic {
		groupId := fmt.Sprintf("%s-%s", groupPrefix, topic)

		processorFunc := processorFactory(topic)

		if partitionsCount > 0 {
			consumers, consumerErr := CreateConsumers(logger, ctx, brokers, groupId, topic, processorFunc, partitionsCount)
			if consumerErr != nil {
				return nil, fmt.Errorf("failde to create service for topic %s: %w", topic, consumerErr)
			}

			manager.consumerGroups[topic] = consumers
		}
	}

	return manager, nil
}

func (m *ConsumerGroupManager) StartAllGroups() {
	for topic, group := range m.consumerGroups {
		m.wg.Add(1)
		m.logger.Info("Starting consumer group for topic", "topic", topic)
		go func() {
			for i, consumer := range group {
				m.logger.Info("Started consumer", "id", i, "topic", topic)
				defer m.wg.Done()
				consumer.Run()
			}
		}()
	}
}

func (m *ConsumerGroupManager) CloseAllGroups() {
	for _, group := range m.consumerGroups {
		for _, inst := range group {
			inst.cancelFunc()
		}
	}
	m.wg.Wait()
}

func (ci *ConsumerInstance) Run() {
	for {
		select {
		case <-ci.ctx.Done():
			log.Println("Context cancelled: kafka consumer shutting down")
			return
		default:
			ev := ci.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch msg := ev.(type) {
			case *kafka.Message:
				ci.semCh <- struct{}{}
				go func() {
					defer func() {
						<-ci.semCh
					}()
					ci.processMessage(msg)
				}()
			case kafka.Error:
				log.Printf("Kafka error: %v", msg)
			}
		}
	}
}

func (ci *ConsumerInstance) processMessage(msg *kafka.Message) {
	timeoutCtx, timeoutCancel := context.WithTimeout(ci.ctx, 30*time.Second)
	defer timeoutCancel()
	err := ci.processor.Process(timeoutCtx, msg)
	if err != nil {
		log.Printf("Processor error: %v, scheduling retry...", err)
		ci.scheduleRetry(*msg, 0)
		return
	}

	_, err = ci.consumer.CommitMessage(msg)
	if err != nil {
		log.Printf("Failed to commit message: %v", err)
	}
}

func (ci *ConsumerInstance) scheduleRetry(msg kafka.Message, attempt int) {
	if attempt > 5 {
		ci.logger.Error("Max retries exceeded", "key", string(msg.Key))
		return
	}

	go func() {
		time.Sleep(5 * time.Second)
		ci.retryCh <- RetryMsg{attempt + 1, msg}
	}()
}

func (ci *ConsumerInstance) retryLoop() {
	for {
		select {
		case <-ci.ctx.Done():
			log.Println("Context cancelled: stop retry loop")
			return
		case retryMsg := <-ci.retryCh:
			err := ci.processor.Process(ci.ctx, &retryMsg.msg)
			if err != nil {
				log.Printf("Scheduled processor error: %v", err)
				return
			}
		}
	}
}

func CreateConsumers(logger *slog.Logger, ctx context.Context, brokers, groupID, topic string, processor processor.Processor, partitionsCount int) ([]*ConsumerInstance, error) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"group.id":                 groupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.partition.eof":     true,
		"go.application.rebalance": true,
	}

	consumers := make([]*ConsumerInstance, 0, partitionsCount)

	for i := 0; i < partitionsCount; i++ {
		ctxWithCancel, cancel := context.WithCancel(ctx)
		c, err := kafka.NewConsumer(config)
		if err != nil {
			cancel()
			return nil, err
		}

		err = c.Subscribe(topic, nil)
		if err != nil {
			cancel()
			closeErr := c.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}

		consumers = append(consumers, &ConsumerInstance{
			ID:         fmt.Sprintf("%s-%s-%d", topic, groupID, i),
			logger:     logger,
			consumer:   c,
			ctx:        ctxWithCancel,
			cancelFunc: cancel,
			processor:  processor,
			retryCh:    make(chan RetryMsg, 1000),
			semCh:      make(chan struct{}, 10),
		})
	}

	return consumers, nil
}

func GetTopicPartitionCounts(brokers string, topics []string) (map[string]int, error) {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}
	defer admin.Close()

	metadata, err := admin.GetMetadata(nil, true, 10_000)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, topic := range topics {
		topicMeta, ok := metadata.Topics[topic]
		if !ok {
			continue
		}
		counts[topic] = len(topicMeta.Partitions)
	}

	return counts, nil
}

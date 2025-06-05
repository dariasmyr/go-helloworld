package kafka

import (
	"slog"
	"kafka"
	"context"
)

type ConsumerGroupManager struct {
	logger *slog.logger
	consumerGroups map[string][]*ConsumerInstance
	wg sync.WaitGroup
}

type ConsumerInstance struct {
	ID string
	ctx context.Context
	consumer *kafka.Consumer
	cancelFunc context.cancelFunc
	processor processor.Processor
	retryCh chan kafka.Message
}

func NewConsumerGroupManager (
	logger *slog.logger, 
	brokers, groupPrefix string, 
	topics []string,
	processorFactory func(topic string) processor.Processor,
	) (*ConsumerGroupManager, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	partitionsPerTopic, partitionsErr := GetTopicPartitionCounts(brokers, topic)
		if partitionsErr != nil {
			return nil, fmt.Error("error getting partitions: %s: %w", topic, partitionsErr)
	}

	manager := &ConsumerGroupManager{
		consumerGroups: make(map[string][]*ConsumerInstance),
	}

	for topic, partitionsCount := range partitionsPerTopic {
		groupId := fmt.Sprintf("%s-%s", groupPrefix, topic)

		processor := processor.GetProcessor(topic)

		if partitionCount > 0 {
			consumers, consumerErr := CreateConsumers(ctx, brokers, groupId, topic, processor, partitionsCount) 
		    if consumerErr != nil {
			    return nil, fmt.Error("failde to create service for topic %s: %w", topic, err)
		    }

		    manager.consumerGroups[topic] = consumerGroup
		}
	}

	return manager, nil
}

func (m *ConsumerGroupManager) StartAllGroups() {
	for topic, group := range m.consumerGroups {
		m.wg.Add(1)
		m.logger.Info("Starting consumer group for topic", "topic", topic)
		go func(){
			for i, consumer := range group {
				m.logger.Info("Started consumer", "id", i, "topic", topic)
				defer m.wd.Done()
			    consumer.Run()
			}
		}()
	}
}

func (m *ConsumerGroupManager) CloseAllGroups() {
	for topic, group := range m.consumerGroups {
		for _, inst := range group.consumers {
			inst.Cancel()
		}
	}
	m.wg.Wait()
}

func (ci *ConsumerInstance) Run() {
	for {
		select {
		case <-ci.Context.Done():
			log.Println("Context cancelled: kafka consumer shutting down")
			return
		default:
			ev := ci.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch msg := ev.(type) {
			case *kafka.Message:
				go ci.consumer.processMessage(msg)
			case kafka.Error:
				log.Printf("Kafka error: %v", msg)
			}
		}
	}
}

func (ci *ConsumerInstance) processMessage(msg *kafka.Message) {
	err := ci.processor.Process(msg.Value)
	if err != nil {
		log.Printf("Processor error: %v, scheduling retry...", err)
		ci.scheduleRetry(*msg)
		return
	}

	_, err = ci.consumer.CommitMessage(msg)
	if err != nil {
		log.Printf("Failed to commit message: %v", err)
	}
}

func (ci *ConsumerInstance) scheduleRetry(msg kafka.Message) {
	go func(){
		time.Sleep(5 * time.Second)
		ci.retryCh <- msg
	}()
}

func (ci *ConsumerInstance) retryLoop() {
	for {
		select {
		case <- ci.Context.Done():
			log.Println("Context cancelled: stop retry loop")
			return
		case msg:= <- ci.retryCh:
			cs.processor.Process(&msg)
		}
	}
}


func CreateConsumers(ctx context.Context, brokers, groupId, topic string, processor processor.Processor, partitionsCount int) []*ConsumerInstance {
	config := &kafka.ConfigMap{
		"bootstrap.servers":        brokers,
		"group.id":                 groupID,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.partition.eof":     true,
		"go.application.rebalance": true,
	}

	consumers := make([]*ConsumerInstance, 0, partitionsCount)

	for i:=0; i<partitionCount; i++ {
		ctxWithCancel, cancel := context.WithCancel(ctx)
		c, err := kafka.NewConsumer(config)
		if err != nil {
			return nil, err
		}
		
		err = c.SubscribeTopic(topic, nil)
		if err != nil {
			return nil, err
		}

		consumers[i] = &ConsumerInstance{
			ID: fmt.Sprintf("%s-%s-%d", topic, groupId, i),
			consumer: c,
			Context: ctxWithCancel,
			Cancel: cancel,
			processor: processor,
			make(chan kafka.Message, 1000),
		}
	}

	return consumers
} 


func GetTopicPartitionCounts(brokers, topic string) (map[string]int error) {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}
	defer admin.Close()

	metadata, err := admin.GetMetadata(nil, true, 10_000)
	if err !- nil {
		return nil, err
	}

	couns := make(map[string]int)
	for _, topic := range topics {
		topicMeta, ok := metadata.Topics[topic]
		if !ok {
			continue
		}
		counts[topic] = len(topicMeta.Partitions)
	}

	return counts, nil
}
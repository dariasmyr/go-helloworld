package processor

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Processor interface {
	Process(ctx context.Context, msg *kafka.Message) error
}

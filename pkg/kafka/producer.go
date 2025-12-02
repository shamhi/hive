package kafka

import "github.com/segmentio/kafka-go"

type Producer struct {
	Writer *kafka.Writer
}

func NewProducer(cfg Config) *Producer {
	return &Producer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.Brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}

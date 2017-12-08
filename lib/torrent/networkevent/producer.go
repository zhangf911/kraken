package networkevent

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"code.uber.internal/go-common.git/x/kafka"
	"code.uber.internal/infra/kraken/utils/log"
)

// Producer emits events.
type Producer interface {
	Produce(e *Event) error
}

type producer struct {
	config Config
	rest   *kafka.RestProducer
	logger *zap.SugaredLogger
}

func newLogger(config Config) (*zap.SugaredLogger, error) {
	eventConfig := zap.NewProductionConfig()

	if config.Enabled {
		if config.LogPath == "" {
			return nil, errors.New("no network event path defined")
		}

		eventConfig.OutputPaths = []string{config.LogPath}
		eventConfig.ErrorOutputPaths = []string{config.LogPath}
	}

	logger, err := eventConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// NewProducer creates a new Kafka producer.
func NewProducer(config Config) (Producer, error) {
	rest, err := kafka.NewRestProducer()
	if err != nil {
		return nil, err
	}
	if config.Enabled && config.KafkaTopic == "" {
		return nil, errors.New("no kafka topic supplied")
	}
	if !config.Enabled {
		log.Warn("Kafka network events not enabled")
	}
	logger, err := newLogger(config)
	if err != nil {
		return nil, fmt.Errorf("event logger: %s", err)
	}
	return &producer{config, rest, logger}, nil
}

// Produce publishes e on the configured Kafka topic.
func (p *producer) Produce(e *Event) error {
	if !p.config.Enabled {
		return nil
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("json marshal: %s", err)
	}
	p.logger.Info(string(b))
	return p.rest.Produce(p.config.KafkaTopic, b)
}
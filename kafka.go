package main

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
)

type KafkaConfig struct {
	Enabled bool
	Brokers []string
	Topic   string
}

type KafkaOutput struct {
	Config   KafkaConfig
	Producer sarama.SyncProducer
}

func InitKafka(conf KafkaConfig) (*KafkaOutput, error) {
	var k = &KafkaOutput{Config: conf}
	var err error
	log.WithField("brokers", conf.Brokers).Info("initializing Kafka producer")
	if k.Producer, err = sarama.NewSyncProducer(conf.Brokers, nil); err != nil {
		return nil, err
	}
	return k, nil
}

func (k *KafkaOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
		log.Error("error converting JobUsageRecord to json")
		log.Debugf("%v", jur)
		return err
	} else {
		msg := &sarama.ProducerMessage{Topic: k.Config.Topic, Value: sarama.ByteEncoder(j)}
		_, _, err := k.Producer.SendMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

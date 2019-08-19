package main

import (
	"encoding/xml"
	"strings"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/opensciencegrid/gracc-collector/gracc"
)

type KafkaConfig struct {
	Enable  bool   `env:"ENABLE"`
	Brokers string `env:"HOST"`
	Topic   string `env:"PORT"`
	Format  string `env:"FORMAT"`
}

type KafkaOutput struct {
	Config   KafkaConfig
	producer sarama.SyncProducer
}

func InitKafka(conf KafkaConfig) (*KafkaOutput, error) {
	log.Info("initializing Kafka")
	brokers := strings.Split(conf.Brokers, ",")
	p, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		return nil, err
	}
	var k = &KafkaOutput{
		Config:   conf,
		producer: p,
	}
	return k, nil
}

func (k *KafkaOutput) PublishRecord(rec gracc.Record) error {
	ll := log.WithFields(log.Fields{
		"where": "KafkaOutput.PublishRecord",
	})
	msg := k.makeMessage(rec)
	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		ll.WithFields(log.Fields{
			"topic":     k.Config.Topic,
			"partition": partition,
			"offset":    offset,
		}).Errorf("error sending record: %s", err)
	} else {
		ll.WithFields(log.Fields{
			"topic":     k.Config.Topic,
			"partition": partition,
			"offset":    offset,
		}).Debug("record sent")

	}
	return nil
}

func (k *KafkaOutput) makeMessage(jur gracc.Record) *sarama.ProducerMessage {
	ll := log.WithFields(log.Fields{
		"where": "KafkaOuput.makePublishing",
	})
	msg := sarama.ProducerMessage{Topic: k.Config.Topic}

	switch k.Config.Format {
	case "raw":
		msg.Value = sarama.ByteEncoder(jur.Raw())
	case "xml":
		if j, err := xml.Marshal(jur); err != nil {
			ll.Error("error converting JobUsageRecord to xml")
			ll.Debugf("%v", jur)
			return nil
		} else {
			msg.Value = sarama.ByteEncoder(j)
		}
	default:
		if j, err := jur.ToJSON("    "); err != nil {
			ll.Error("error converting JobUsageRecord to json")
			ll.Debugf("%v", jur)
			return nil
		} else {
			msg.Value = sarama.ByteEncoder(j)
		}
	}
	return &msg
}

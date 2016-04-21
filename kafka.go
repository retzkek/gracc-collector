package main

import (
	"encoding/json"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/opensciencegrid/gracc-collector/gracc"
)

type KafkaConfig struct {
	Enabled bool
	Brokers []string
	Topic   string
}

type KafkaOutput struct {
	Config     KafkaConfig
	Producer   sarama.SyncProducer
	outputChan chan gracc.Record
}

func InitKafka(conf KafkaConfig) (*KafkaOutput, error) {
	var k = &KafkaOutput{Config: conf}
	var err error
	log.WithField("brokers", conf.Brokers).Info("initializing Kafka producer")
	if k.Producer, err = sarama.NewSyncProducer(conf.Brokers, nil); err != nil {
		return nil, err
	}
	k.outputChan = make(chan gracc.Record, 10)
	go k.OutputRecords()
	return k, nil
}

func (k *KafkaOutput) Type() string {
	return "kafka"
}

func (k *KafkaOutput) OutputChan() chan gracc.Record {
	return k.outputChan
}

func (k *KafkaOutput) OutputRecords() {
	for jur := range k.outputChan {
		if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
			log.Error("error converting JobUsageRecord to json")
			log.Debugf("%v", jur)
			//return err
		} else {
			msg := &sarama.ProducerMessage{Topic: k.Config.Topic, Value: sarama.ByteEncoder(j)}
			_, _, err := k.Producer.SendMessage(msg)
			if err != nil {
				log.Error(err)
				//return err
			}
		}
	}
	//return nil
}

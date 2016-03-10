package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Vhost    string
	User     string
	Password string
	Format   string
}

type AMQPOutput struct {
	Config     AMQPConfig
	URI        string
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func InitAMQP(conf AMQPConfig) (*AMQPOutput, error) {
	var a = &AMQPOutput{
		Config: conf,
		URI: "amqp://" + conf.User + ":" + conf.Password + "@" +
			conf.Host + ":" + conf.Port + "/" + conf.Vhost,
	}
	log.WithFields(log.Fields{
		"user":  conf.User,
		"host":  conf.Host,
		"vhost": conf.Vhost,
		"port":  conf.Port,
	}).Info("connecting to RabbitMQ")
	var err error
	if a.Connection, err = amqp.Dial(a.URI); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *AMQPOutput) Type() string {
	return "AMQP"
}

func (a *AMQPOutput) StartBatch() error {
	var err error
	a.Channel, err = a.Connection.Channel()
	return err
}

func (a *AMQPOutput) EndBatch() error {
	return a.Channel.Close()
}

// OutputJUR sends a JobUsageRecord.
func (a *AMQPOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	return fmt.Errorf("not implemented")
}

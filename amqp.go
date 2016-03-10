package main

import (
	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Enabled    bool
	Host       string
	Port       string
	Vhost      string
	Queue      string
	Exchange   string
	User       string
	Password   string
	Format     string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
}

type AMQPOutput struct {
	Config     AMQPConfig
	URI        string
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      amqp.Queue
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
	// declare our queue
	log.WithFields(log.Fields{
		"queue":      a.Config.Queue,
		"durable":    a.Config.Durable,
		"autoDelete": a.Config.AutoDelete,
		"exclusive":  a.Config.Exclusive,
	}).Debug("AMQP: declaring queue")
	if a.Queue, err = a.Channel.QueueDeclare(a.Config.Queue, a.Config.Durable,
		a.Config.AutoDelete, a.Config.Exclusive, false, nil); err != nil {
		return err
	}
	// Bind the queue to the exchange, matching all routing keys
	log.WithFields(log.Fields{
		"exchange": a.Config.Exchange,
		"queue":    a.Queue.Name,
	}).Debug("AMQP: binding exchange to queue")
	if err = a.Channel.QueueBind(a.Queue.Name, "#", a.Config.Exchange, false, nil); err != nil {
		return err
	}
	return err
}

func (a *AMQPOutput) EndBatch() error {
	return a.Channel.Close()
}

// OutputJUR sends a JobUsageRecord.
func (a *AMQPOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	return nil
}

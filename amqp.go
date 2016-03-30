package main

import (
	"encoding/json"
	"encoding/xml"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	"github.com/streadway/amqp"
)

type AMQPConfig struct {
	Enabled      bool
	Host         string
	Port         string
	Vhost        string
	User         string
	Password     string
	Format       string
	Exchange     string
	ExchangeType string
	Durable      bool
	AutoDelete   bool
	Internal     bool
	RoutingKey   string
}

type AMQPOutput struct {
	Config      AMQPConfig
	URI         string
	Connection  *amqp.Connection
	Channel     *amqp.Channel
	ChannelOpen bool
	outputChan  chan gracc.Record
}

func InitAMQP(conf AMQPConfig) (*AMQPOutput, error) {
	var a = &AMQPOutput{
		Config: conf,
		URI: "amqp://" + conf.User + ":" + conf.Password + "@" +
			conf.Host + ":" + conf.Port + "/" + conf.Vhost,
		ChannelOpen: false,
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
	if err = a.refreshChannel(); err != nil {
		return nil, err
	}
	a.outputChan = make(chan gracc.Record, 10)
	go a.OutputRecords()
	return a, nil
}

func (a *AMQPOutput) Type() string {
	return "AMQP"
}

func (a *AMQPOutput) OutputChan() chan gracc.Record {
	return a.outputChan
}

func (a *AMQPOutput) refreshChannel() error {
	var err error
	if a.ChannelOpen {
		a.Channel.Close()
	}
	a.ChannelOpen = false
	a.Channel, err = a.Connection.Channel()
	// declare our queue
	log.WithFields(log.Fields{
		"name":       a.Config.Exchange,
		"type":       a.Config.ExchangeType,
		"durable":    a.Config.Durable,
		"autoDelete": a.Config.AutoDelete,
		"internal":   a.Config.Internal,
	}).Debug("AMQP: declaring exchange")
	if err = a.Channel.ExchangeDeclare(a.Config.Exchange, a.Config.ExchangeType,
		a.Config.Durable, a.Config.AutoDelete, a.Config.Internal, false, nil); err != nil {
		a.Channel.Close()
		return err
	}
	a.ChannelOpen = true
	return err
}

func (a *AMQPOutput) OutputRecords() {
	for jur := range a.outputChan {
		if !a.ChannelOpen {
			a.refreshChannel()
		}
		var pub amqp.Publishing
		switch a.Config.Format {
		case "raw":
			pub.ContentType = "text/xml"
			pub.Body = jur.Raw()
		case "xml":
			if j, err := xml.Marshal(jur); err != nil {
				log.Error("error converting JobUsageRecord to xml")
				log.Debugf("%v", jur)
				//return err
			} else {
				pub.ContentType = "text/xml"
				pub.Body = j
			}
		case "json":
			if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
				log.Error("error converting JobUsageRecord to json")
				log.Debugf("%v", jur)
				//return err
			} else {
				pub.ContentType = "application/json"
				pub.Body = j
			}
		}
		log.WithFields(log.Fields{
			"exchange":   a.Config.Exchange,
			"routingKey": a.Config.RoutingKey,
		}).Debug("AMQP: publishing record")
		if err := a.Channel.Publish(
			a.Config.Exchange, // exchange
			"",                // routing key
			false,             // mandatory
			false,             // immediate
			pub); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warning("AMQP: error publishing to channel; refreshing channel")
			a.refreshChannel()
			// put record back in queue to resend
			a.outputChan <- jur
		}
	}
	//return err
}

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
	Workers      int
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
	if err := a.dial(); err != nil {
		return nil, err
	}
	a.outputChan = make(chan gracc.Record)
	for i := 0; i < conf.Workers; i++ {
		go a.OutputRecords()
	}
	return a, nil
}

func (a *AMQPOutput) dial() error {
	if a.Connection != nil {
		a.Connection.Close()
	}
	log.WithFields(log.Fields{
		"user":  a.Config.User,
		"host":  a.Config.Host,
		"vhost": a.Config.Vhost,
		"port":  a.Config.Port,
	}).Info("AMQP: connecting to RabbitMQ")
	var err error
	if a.Connection, err = amqp.Dial(a.URI); err != nil {
		log.WithField("error", err).Error("AMQP: error connecting to RabbitMQ")
		return err
	}
	// listen for close events
	closing := a.Connection.NotifyClose(make(chan *amqp.Error))
	go func() {
		for c := range closing {
			log.WithFields(log.Fields{
				"code":             c.Code,
				"reason":           c.Reason,
				"server-initiated": c.Server,
				"can-recover":      c.Recover,
			}).Warning("AMQP: connection closed")
		}
	}()
	return nil
}

func (a *AMQPOutput) Type() string {
	return "AMQP"
}

func (a *AMQPOutput) OutputChan() chan gracc.Record {
	return a.outputChan
}

func (a *AMQPOutput) OutputRecords() {
	amqpChan, err := a.Connection.Channel()
	if err != nil {
		log.WithField("error", err).Panic("AMQP: error opening channel")
	}
	defer amqpChan.Close()
	// listen for close events
	closing := amqpChan.NotifyClose(make(chan *amqp.Error))
	go func() {
		for c := range closing {
			log.WithFields(log.Fields{
				"code":             c.Code,
				"reason":           c.Reason,
				"server-initiated": c.Server,
				"can-recover":      c.Recover,
			}).Warning("AMQP: channel closed")
			amqpChan, err = a.Connection.Channel()
			if err != nil {
				log.WithField("error", err).Panic("AMQP: error opening channel")
			}
		}
	}()
	log.WithFields(log.Fields{
		"name":       a.Config.Exchange,
		"type":       a.Config.ExchangeType,
		"durable":    a.Config.Durable,
		"autoDelete": a.Config.AutoDelete,
		"internal":   a.Config.Internal,
	}).Debug("AMQP: declaring exchange")
	if err = amqpChan.ExchangeDeclare(a.Config.Exchange, a.Config.ExchangeType,
		a.Config.Durable, a.Config.AutoDelete, a.Config.Internal, false, nil); err != nil {
		amqpChan.Close()
		log.WithField("error", err).Panic("AMQP: error declaring exchange")
	}
	for jur := range a.outputChan {
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
		if err := amqpChan.Publish(
			a.Config.Exchange, // exchange
			"",                // routing key
			false,             // mandatory
			false,             // immediate
			pub); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Warning("AMQP: error publishing to channel")
			// put record back in queue to resend, if possible
			select {
			case a.outputChan <- jur:
			default:
				log.Error("AMQP: no workers available to handle record")
			}
		}
	}
	//return err
}

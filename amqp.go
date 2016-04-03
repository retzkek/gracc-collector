package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"

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
		go a.outputWorker(i)
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
	// listen for blocking events
	blockings := a.Connection.NotifyBlocked(make(chan amqp.Blocking))
	go func() {
		for b := range blockings {
			if b.Active {
				log.WithField("reason", b.Reason).Warning("AMQP: TCP blocked")
			} else {
				log.Info("AMQP: TCP unblocked")
			}
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

func (a *AMQPOutput) outputWorker(id int) {
	wlog := log.WithFields(log.Fields{
		"output": fmt.Sprintf("AMQP:%d", id),
	})
	wlog.Info("starting worker")
	amqpChan, err := a.Connection.Channel()
	if err != nil {
		wlog.WithField("error", err).Panic("error opening channel")
	}
	defer amqpChan.Close()

	ack, nack := a.setupNotifiers(amqpChan, wlog)

	for jur := range a.outputChan {
		pub := a.makePublishing(jur, wlog)
		if pub == nil {
			continue
		}
		wlog.WithFields(log.Fields{
			"exchange":   a.Config.Exchange,
			"routingKey": a.Config.RoutingKey,
			"record":     jur.Id(),
		}).Debug("publishing record")
		if err := amqpChan.Publish(
			a.Config.Exchange, // exchange
			"",                // routing key
			true,              // mandatory
			false,             // immediate
			*pub); err != nil {
			wlog.WithFields(log.Fields{
				"error": err,
			}).Warning("error publishing to channel")
			// put record back in queue to resend, if possible
			select {
			case a.outputChan <- jur:
			default:
				wlog.Error("no workers available to handle record")
			}
		}
		wlog.WithFields(log.Fields{
			"exchange":   a.Config.Exchange,
			"routingKey": a.Config.RoutingKey,
			"record":     jur.Id(),
		}).Debug("record sent, waiting for ack")
		// wait for ACK/NACK
		select {
		case tag := <-ack:
			wlog.WithField("tag", tag).Debug("ack")
		case tag := <-nack:
			wlog.WithField("tag", tag).Warning("nack")
		}
	}
	wlog.Warning("worker exiting")
}

func (a *AMQPOutput) setupNotifiers(amqpChan *amqp.Channel, wlog *log.Entry) (ack, nack chan uint64) {
	var err error
	// listen for close events
	closing := amqpChan.NotifyClose(make(chan *amqp.Error))
	go func() {
		for c := range closing {
			wlog.WithFields(log.Fields{
				"code":             c.Code,
				"reason":           c.Reason,
				"server-initiated": c.Server,
				"can-recover":      c.Recover,
			}).Warning("channel closed")
			amqpChan, err = a.Connection.Channel()
			if err != nil {
				wlog.WithField("error", err).Panic("error opening channel")
			}
		}
	}()
	// listen for ACK/NACK
	ack, nack = amqpChan.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	amqpChan.Confirm(false)
	wlog.WithFields(log.Fields{
		"name":       a.Config.Exchange,
		"type":       a.Config.ExchangeType,
		"durable":    a.Config.Durable,
		"autoDelete": a.Config.AutoDelete,
		"internal":   a.Config.Internal,
	}).Debug("declaring exchange")
	if err = amqpChan.ExchangeDeclare(a.Config.Exchange, a.Config.ExchangeType,
		a.Config.Durable, a.Config.AutoDelete, a.Config.Internal, false, nil); err != nil {
		amqpChan.Close()
		wlog.WithField("error", err).Panic("error declaring exchange")
	}
	return ack, nack
}

func (a *AMQPOutput) makePublishing(jur gracc.Record, wlog *log.Entry) *amqp.Publishing {
	var pub amqp.Publishing
	switch a.Config.Format {
	case "raw":
		pub.ContentType = "text/xml"
		pub.Body = jur.Raw()
	case "xml":
		if j, err := xml.Marshal(jur); err != nil {
			wlog.Error("error converting JobUsageRecord to xml")
			wlog.Debugf("%v", jur)
			return nil
		} else {
			pub.ContentType = "text/xml"
			pub.Body = j
		}
	case "json":
		if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
			wlog.Error("error converting JobUsageRecord to json")
			wlog.Debugf("%v", jur)
			return nil
		} else {
			pub.ContentType = "application/json"
			pub.Body = j
		}
	}
	return &pub
}

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
	recoverChan chan gracc.Record
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
		go StartWorker(fmt.Sprintf("%d", i), a, a.outputChan)
	}
	a.recoverChan = make(chan gracc.Record)
	for i := 0; i < conf.Workers; i++ {
		go StartWorker(fmt.Sprintf("recover:%d", i), a, a.recoverChan)
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

type AMQPWorker struct {
	id         string
	manager    *AMQPOutput
	listenChan chan gracc.Record
	wlog       *log.Entry
	amqpChan   *amqp.Channel
	ack, nack  chan uint64
}

func StartWorker(id string, manager *AMQPOutput, ch chan gracc.Record) {
	var a = AMQPWorker{
		id:         id,
		manager:    manager,
		listenChan: ch,
	}
	a.wlog = log.WithFields(log.Fields{
		"output": fmt.Sprintf("AMQP:%s", id),
	})
	a.wlog.Info("starting worker")
	var err error
	a.amqpChan, err = a.manager.Connection.Channel()
	if err != nil {
		a.wlog.WithField("error", err).Panic("error opening channel")
	}
	defer a.amqpChan.Close()

	a.setupNotifiers()

	for jur := range a.listenChan {
		pub := a.makePublishing(jur)
		if pub == nil {
			continue
		}
		a.wlog.WithFields(log.Fields{
			"exchange":   a.manager.Config.Exchange,
			"routingKey": a.manager.Config.RoutingKey,
			"record":     jur.Id(),
		}).Debug("publishing record")
		if err := a.amqpChan.Publish(
			a.manager.Config.Exchange, // exchange
			"",    // routing key
			true,  // mandatory
			false, // immediate
			*pub); err != nil {
			a.wlog.WithFields(log.Fields{
				"error": err,
			}).Warning("error publishing to channel")
			// put record into recovery queue
			select {
			case a.manager.recoverChan <- jur:
			default:
				a.wlog.Error("no workers available to handle record")
			}
		}
		a.wlog.WithFields(log.Fields{
			"exchange":   a.manager.Config.Exchange,
			"routingKey": a.manager.Config.RoutingKey,
			"record":     jur.Id(),
		}).Debug("record sent, waiting for ack")
		// wait for ACK/NACK
		select {
		case tag := <-a.ack:
			a.wlog.WithField("tag", tag).Debug("ack")
		case tag := <-a.nack:
			a.wlog.WithField("tag", tag).Warning("nack")
		}
	}
	a.wlog.Warning("worker exiting")
}

func (a *AMQPWorker) setupNotifiers() {
	var err error
	// listen for close events
	closing := a.amqpChan.NotifyClose(make(chan *amqp.Error))
	go func() {
		for c := range closing {
			a.wlog.WithFields(log.Fields{
				"code":             c.Code,
				"reason":           c.Reason,
				"server-initiated": c.Server,
				"can-recover":      c.Recover,
			}).Warning("channel closed")
			a.amqpChan, err = a.manager.Connection.Channel()
			if err != nil {
				a.wlog.WithField("error", err).Panic("error opening channel")
			}
		}
	}()
	// listen for ACK/NACK
	a.ack, a.nack = a.amqpChan.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))
	a.amqpChan.Confirm(false)
	a.wlog.WithFields(log.Fields{
		"name":       a.manager.Config.Exchange,
		"type":       a.manager.Config.ExchangeType,
		"durable":    a.manager.Config.Durable,
		"autoDelete": a.manager.Config.AutoDelete,
		"internal":   a.manager.Config.Internal,
	}).Debug("declaring exchange")
	if err = a.amqpChan.ExchangeDeclare(a.manager.Config.Exchange, a.manager.Config.ExchangeType,
		a.manager.Config.Durable, a.manager.Config.AutoDelete, a.manager.Config.Internal, false, nil); err != nil {
		a.amqpChan.Close()
		a.wlog.WithField("error", err).Panic("error declaring exchange")
	}
}

func (a *AMQPWorker) makePublishing(jur gracc.Record) *amqp.Publishing {
	var pub amqp.Publishing
	switch a.manager.Config.Format {
	case "raw":
		pub.ContentType = "text/xml"
		pub.Body = jur.Raw()
	case "xml":
		if j, err := xml.Marshal(jur); err != nil {
			a.wlog.Error("error converting JobUsageRecord to xml")
			a.wlog.Debugf("%v", jur)
			return nil
		} else {
			pub.ContentType = "text/xml"
			pub.Body = j
		}
	case "json":
		if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
			a.wlog.Error("error converting JobUsageRecord to json")
			a.wlog.Debugf("%v", jur)
			return nil
		} else {
			pub.ContentType = "application/json"
			pub.Body = j
		}
	}
	return &pub
}

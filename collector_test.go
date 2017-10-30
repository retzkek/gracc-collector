package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	config = &CollectorConfig{
		Address:  "localhost",
		Port:     "8787",
		Timeout:  "10s",
		LogLevel: "DEBUG",
		AMQP: AMQPConfig{
			Host:         "localhost",
			Port:         "5672",
			Scheme:       "amqp",
			User:         "guest",
			Password:     "guest",
			Format:       "json",
			Exchange:     "gracc.test",
			ExchangeType: "fanout",
			Durable:      false,
			AutoDelete:   true,
			Internal:     false,
			RoutingKey:   "",
			Retry:        "1s",
			MaxRetry:     "10s",
		},
		StartBufferSize: 4096,
		MaxBufferSize:   512 * 1024,
	}
	collector *GraccCollector
	consumer  *AMQPOutput
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		log.SetLevel(log.DebugLevel)
	}
	// initialize collector
	var err error
	if err = config.Validate(); err != nil {
		log.Fatalf("error in collector config: %s", err)
	}
	if err = config.GetEnv(); err != nil {
		log.Fatalf("error getting env var: %s", err)
	}
	if collector, err = NewCollector(config); err != nil {
		log.Fatalf("error starting collector: %s", err)
	}

	// start HTTP server
	http.Handle("/rmi", collector)
	http.HandleFunc("/stats", collector.ServeStats)
	go http.ListenAndServe(config.Address+":"+config.Port, nil)

	// start AMQP consumer
	if err := startConsumer(); err != nil {
		log.Fatalf("error starting consumer: %s", err)
	}

	// run tests
	os.Exit(m.Run())
}

func startConsumer() error {
	var err error
	if consumer, err = InitAMQP(config.AMQP); err != nil {
		return fmt.Errorf("InitAMQP: %s", err)
	}
	cch, err := consumer.OpenChannel()
	if err != nil {
		return fmt.Errorf("OpenChannel: %s", err)
	}
	if err := cch.ExchangeDeclare(config.AMQP.Exchange,
		config.AMQP.ExchangeType,
		config.AMQP.Durable,
		config.AMQP.AutoDelete,
		config.AMQP.Internal,
		false,
		nil); err != nil {
		return fmt.Errorf("ExchangeDeclare: %s", err)
	}
	if _, err := cch.QueueDeclare("gracc.test.queue", false, true, false, false, nil); err != nil {
		return fmt.Errorf("QueueDeclare: %s", err)
	}
	if err := cch.QueueBind("gracc.test.queue", "#", config.AMQP.Exchange, false, nil); err != nil {
		return fmt.Errorf("QueueBind: %s", err)
	}
	inbox, err := cch.Consume("gracc.test.queue", "", false, false, true, false, nil)
	if err != nil {
		return fmt.Errorf("Consume: %s", err)
	}
	// consume records
	go func() {
		for r := range inbox {
			log.Infof("got record %d", r.DeliveryTag)
			log.Debugf("DeliveryMode %d", r.DeliveryMode)
			if r.DeliveryMode != amqp.Persistent {
				// If we find a non-persistent record, bail immediately!
				os.Exit(1)
			}
			log.Debugf("record body:\n---%s\n---\n", r.Body)
			r.Ack(false)
		}
	}()
	return nil
}

func TestPing(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "update")
	v.Set("from", "localhost")
	v.Set("arg1", "xxx")
	v.Set("bundlesize", "1")
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("ping got response %s", resp.Status))
		}
	}
}

func TestUpdate(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "update")
	v.Set("from", "localhost")
	v.Set("arg1", testBundle)
	v.Set("bundlesize", fmt.Sprintf("%d", testBundleSize))
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("update got response %s", resp.Status))
		}
	}
}

func TestMultiUpdate(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "multiupdate")
	v.Set("from", "localhost")
	v.Set("arg1", testBundleXML)
	//v.Set("bundlesize", fmt.Sprintf("%d", testBundleXMLSize))
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("multiupdate got response %s", resp.Status))
		}
	}
}

// Load test data
var (
	testBundleSize int
	testBundle     string
)

func init() {
	f, err := os.Open("test.bundle")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		panic(err)
	}
	testBundle = buf.String()
	testBundleSize = 15
}

var (
	testBundleXML string
)

func init() {
	f, err := os.Open("test_bundle.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(f); err != nil {
		panic(err)
	}
	testBundleXML = buf.String()
}

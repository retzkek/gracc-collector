package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
)

type GraccOutput interface {
	OutputJUR(*gracc.JobUsageRecord) error
}

type GraccCollector struct {
	Config  *CollectorConfig
	Outputs []GraccOutput
}

// NewCollector initializes and returns a new Gracc collector.
func NewCollector(conf *CollectorConfig) (*GraccCollector, error) {
	var g GraccCollector
	g.Config = conf
	g.Outputs = make([]GraccOutput, 0, 4)

	var err error
	if conf.File.Enabled {
		var f *FileOutput
		if f, err = InitFile(conf.File); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, f)
	}
	if conf.Elasticsearch.Enabled {
		var e *ElasticsearchOutput
		if e, err = InitElasticsearch(conf.Elasticsearch); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, e)
	}
	if conf.Kafka.Enabled {
		var k *KafkaOutput
		if k, err = InitKafka(conf.Kafka); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, k)
	}
	if conf.AMQP.Enabled {
		var a *AMQPOutput
		if a, err = InitAMQP(conf.AMQP); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, a)
	}

	return &g, nil
}

func (g GraccCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if g.checkRequiredKeys(w, r, []string{"command"}) != nil {
		return
	}
	command := r.FormValue("command")
	switch command {
	case "update":
		g.handleUpdate(w, r)
	default:
		g.handleError(w, r, "unknown command")
	}
}

func (g GraccCollector) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if g.checkRequiredKeys(w, r, []string{"arg1", "from"}) != nil {
		return
	}
	updateLogger := log.WithFields(log.Fields{
		"from": r.FormValue("from"),
	})
	if r.FormValue("arg1") == "xxx" {
		updateLogger.Info("received test request")
		fmt.Fprintf(w, "OK")
	} else {
		if g.checkRequiredKeys(w, r, []string{"bundlesize"}) != nil {
			return
		}
		if err := g.ProcessBundle(r.FormValue("arg1"), r.FormValue("bundlesize")); err == nil {
			updateLogger.WithField("size", r.FormValue("bundlesize")).Info("received update")
			fmt.Fprintf(w, "OK")
		} else {
			g.handleError(w, r, "error processing bundle")
			log.Debug(err)
			return
		}
	}
}

func (g GraccCollector) checkRequiredKeys(w http.ResponseWriter, r *http.Request, keys []string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			err := fmt.Sprintf("no %v", k)
			g.handleError(w, r, err)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (g GraccCollector) handleError(w http.ResponseWriter, r *http.Request, err string) {
	log.WithField("error", err).Errorf("recieved unknown or misformed request")
	log.Debug(r)
	fmt.Fprintf(w, "Error")
}

func (g *GraccCollector) ProcessBundle(bundle string, bundlesize string) error {
	//fmt.Println("---+++---")
	//fmt.Print(bundle)
	size, err := strconv.Atoi(bundlesize)
	if err != nil {
		return err
	}

	received := 0

	parts := strings.Split(bundle, "|")
	for i := 0; i < len(parts); i++ {
		//fmt.Printf("--- %d ----\n%s---\n\n", i, p)
		switch parts[i] {
		case "":
			continue
		case "replication":
			if err := g.ProcessXml(parts[i+1]); err != nil {
				return err
			}
			received++
			i += 2
		}
	}
	if received != size {
		return fmt.Errorf("actual bundle size (%d) different than expected (%d)", len(parts)-1, size)
	}
	return nil
}

func (g *GraccCollector) ProcessXml(x string) error {
	var jur gracc.JobUsageRecord
	xb := []byte(x)
	if err := xml.Unmarshal(xb, &jur); err != nil {
		return err
	}
	for _, o := range g.Outputs {
		if err := o.OutputJUR(&jur); err != nil {
			return err
		}
	}
	return nil
}

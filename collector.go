package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
)

type GraccOutput interface {
	// Type returns the type of the output.
	Type() string
	// StartBatch performs any setup needed to handle a batch of records.
	StartBatch() error
	// EndBatch performs any cleanup needed after sending a batch of records.
	EndBatch() error
	// OutputRecord sends a Record to the output
	OutputRecord(gracc.Record) error
}

type CollectorStats struct {
	Records       uint64
	RecordErrors  uint64
	Requests      uint64
	RequestErrors uint64
}

type Event int

const (
	GOT_RECORD Event = iota
	RECORD_ERROR
	GOT_REQUEST
	REQUEST_ERROR
)

type GraccCollector struct {
	Config  *CollectorConfig
	Outputs []GraccOutput
	Stats   CollectorStats

	Events chan Event
}

// NewCollector initializes and returns a new Gracc collector.
func NewCollector(conf *CollectorConfig) (*GraccCollector, error) {
	var g GraccCollector
	g.Config = conf
	g.Outputs = make([]GraccOutput, 0, 4)

	g.Events = make(chan Event)
	go g.LogEvents()

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

func (g *GraccCollector) LogEvents() {
	for {
		switch <-g.Events {
		case GOT_RECORD:
			g.Stats.Records++
		case RECORD_ERROR:
			g.Stats.RecordErrors++
		case GOT_REQUEST:
			g.Stats.Requests++
		case REQUEST_ERROR:
			g.Stats.RequestErrors++
		}
	}
}

func (g *GraccCollector) ServeStats(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	if err := enc.Encode(g.Stats); err != nil {
		http.Error(w, "error writing stats", http.StatusInternalServerError)
	}
}

func (g *GraccCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Events <- GOT_REQUEST
	log.WithFields(log.Fields{
		"address": r.RemoteAddr,
		"length":  r.ContentLength,
		"agent":   r.UserAgent(),
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
	}).Info("received request")
	r.ParseForm()
	if g.checkRequiredKeys(w, r, []string{"command"}) != nil {
		return
	}
	command := r.FormValue("command")
	switch command {
	case "update":
		g.handleUpdate(w, r)
	default:
		g.Events <- REQUEST_ERROR
		g.handleError(w, r, "unknown command")
	}
}

func (g *GraccCollector) handleUpdate(w http.ResponseWriter, r *http.Request) {
	if g.checkRequiredKeys(w, r, []string{"arg1", "from"}) != nil {
		g.Events <- REQUEST_ERROR
		return
	}
	updateLogger := log.WithFields(log.Fields{
		"from": r.FormValue("from"),
	})
	if r.FormValue("arg1") == "xxx" {
		updateLogger.Info("received ping")
		fmt.Fprintf(w, "OK")
	} else {
		if g.checkRequiredKeys(w, r, []string{"bundlesize"}) != nil {
			g.Events <- REQUEST_ERROR
			return
		}
		bundlesize, err := strconv.Atoi(r.FormValue("bundlesize"))
		if err != nil {
			g.Events <- REQUEST_ERROR
			updateLogger.WithField("error", err).Warning("error handling update")
			g.handleError(w, r, "error interpreting bundlesize")
			return
		}
		if err := g.ProcessBundle(r.FormValue("arg1"), bundlesize); err == nil {
			updateLogger.WithField("bundlesize", r.FormValue("bundlesize")).Info("received update")
			fmt.Fprintf(w, "OK")
		} else {
			g.Events <- REQUEST_ERROR
			updateLogger.WithField("error", err).Warning("error handling update")
			g.handleError(w, r, "error processing bundle")
			return
		}
	}
}

func (g *GraccCollector) checkRequiredKeys(w http.ResponseWriter, r *http.Request, keys []string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			err := fmt.Sprintf("no %v", k)
			g.handleError(w, r, err)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (g *GraccCollector) handleError(w http.ResponseWriter, r *http.Request, err string) {
	log.WithField("error", err).Warning("error handling request")
	fmt.Fprintf(w, "Error")
}

func (g *GraccCollector) ProcessBundle(bundle string, bundlesize int) error {
	//fmt.Println("---+++---")
	//fmt.Print(bundle)
	// prepare outputs
	for _, o := range g.Outputs {
		if err := o.StartBatch(); err != nil {
			log.WithFields(log.Fields{
				"output": o.Type(),
				"error":  err,
			}).Error("error starting output batch")
		}
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
				log.WithFields(log.Fields{
					"index": i,
					"error": err,
				}).Error("error processing record")
				g.Events <- RECORD_ERROR
			}
			received++
			i += 2
		}
	}

	// clean up outputs
	for _, o := range g.Outputs {
		if err := o.EndBatch(); err != nil {
			log.WithFields(log.Fields{
				"output": o.Type(),
				"error":  err,
			}).Error("error ending output batch")
		}
	}

	if received != bundlesize {
		return fmt.Errorf("actual bundle size (%d) different than expected (%d)", len(parts)-1, bundlesize)
	}
	return nil
}

func (g *GraccCollector) ProcessXml(x string) error {
	g.Events <- GOT_RECORD
	var jur gracc.JobUsageRecord
	if err := jur.ParseXml([]byte(x)); err != nil {
		return err
	}
	for _, o := range g.Outputs {
		if err := o.OutputRecord(&jur); err != nil {
			return err
		}
	}
	return nil
}

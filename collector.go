package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
)

type GraccOutput interface {
	// Type returns the type of the output.
	Type() string
	// OutputChan returns a channel to send a record to be output
	OutputChan() chan gracc.Record
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
	statsm  sync.Mutex

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
			g.statsm.Lock()
			g.Stats.Records++
			g.statsm.Unlock()
		case RECORD_ERROR:
			g.statsm.Lock()
			g.Stats.RecordErrors++
			g.statsm.Unlock()
		case GOT_REQUEST:
			g.statsm.Lock()
			g.Stats.Requests++
			g.statsm.Unlock()
		case REQUEST_ERROR:
			g.statsm.Lock()
			g.Stats.RequestErrors++
			g.statsm.Unlock()
		}
	}
}

func (g *GraccCollector) ServeStats(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	g.statsm.Lock()
	stats := g.Stats
	g.statsm.Unlock()
	if err := enc.Encode(stats); err != nil {
		http.Error(w, "error writing stats", http.StatusInternalServerError)
	}
}

func (g *GraccCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Events <- GOT_REQUEST
	rlog := log.WithFields(log.Fields{
		"address": r.RemoteAddr,
		"length":  r.ContentLength,
		"agent":   r.UserAgent(),
		"path":    r.URL.EscapedPath(),
	})
	r.ParseForm()
	if err := g.checkRequiredKeys(w, r, []string{"command"}); err != nil {
		g.Events <- REQUEST_ERROR
		g.handleError(w, r, rlog, err, http.StatusBadRequest)
		return
	}
	command := r.FormValue("command")
	switch command {
	case "update":
		g.handleUpdate(w, r, rlog)
	default:
		g.Events <- REQUEST_ERROR
		g.handleError(w, r, rlog, fmt.Errorf("unknown command"), http.StatusBadRequest)
	}
}

func (g *GraccCollector) handleUpdate(w http.ResponseWriter, r *http.Request, rlog *log.Entry) {
	if err := g.checkRequiredKeys(w, r, []string{"arg1", "from"}); err != nil {
		g.Events <- REQUEST_ERROR
		g.handleError(w, r, rlog, err, http.StatusBadRequest)
		return
	}
	updateLogger := log.WithFields(log.Fields{
		"from": r.FormValue("from"),
	})
	if r.FormValue("arg1") == "xxx" {
		updateLogger.Info("received ping")
		g.handleSuccess(w, r, rlog)
		return
	} else {
		if err := g.checkRequiredKeys(w, r, []string{"bundlesize"}); err != nil {
			g.Events <- REQUEST_ERROR
			g.handleError(w, r, rlog, err, http.StatusBadRequest)
			return
		}
		bundlesize, err := strconv.Atoi(r.FormValue("bundlesize"))
		if err != nil {
			g.Events <- REQUEST_ERROR
			updateLogger.WithField("error", err).Error("error handling update")
			g.handleError(w, r, rlog, fmt.Errorf("error interpreting bundlesize"), http.StatusBadRequest)
			return
		}
		if err := g.ProcessBundle(r.FormValue("arg1"), bundlesize); err == nil {
			updateLogger.WithField("bundlesize", r.FormValue("bundlesize")).Info("received update")
			g.handleSuccess(w, r, rlog)
			return
		} else {
			g.Events <- REQUEST_ERROR
			updateLogger.WithField("error", err).Error("error handling update")
			g.handleError(w, r, rlog, fmt.Errorf("error processing bundle"), http.StatusInternalServerError)
			return
		}
	}
}

func (g *GraccCollector) checkRequiredKeys(w http.ResponseWriter, r *http.Request, keys []string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			err := fmt.Sprintf("no %v", k)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (g *GraccCollector) handleError(w http.ResponseWriter, r *http.Request, rlog *log.Entry, err error, code int) {
	res := fmt.Sprintf("Error: %s", err)
	rlog.WithFields(log.Fields{
		"response":      res,
		"response-code": code,
		"error":         err,
	}).Info("handled request")
	w.WriteHeader(code)
	fmt.Fprint(w, res)
}

func (g *GraccCollector) handleSuccess(w http.ResponseWriter, r *http.Request, rlog *log.Entry) {
	rlog.WithFields(log.Fields{
		"response":      "OK",
		"response-code": 200,
	}).Info("handled request")
	fmt.Fprintf(w, "OK")
}

func (g *GraccCollector) ProcessBundle(bundle string, bundlesize int) error {
	received := 0
	parts := strings.Split(bundle, "|")
	for i := 0; i < len(parts); i++ {
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
				return fmt.Errorf("error processing record")
			}
			received++
			i += 2
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
		select {
		case o.OutputChan() <- &jur:
		case <-time.After(g.Config.Timeout * time.Second):
			return fmt.Errorf("timed out waiting for %s output worker", o.Type())
		}
	}
	return nil
}

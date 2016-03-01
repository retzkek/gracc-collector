package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "gratia.cfg", "config file")
}

func main() {
	flag.Parse()

	log.WithField("file", configFile).Info("reading config")
	config, err := ReadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.WithField("level", config.LogLevel).Info("setting log level")
	logLevel, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)

	log.Info("initializing collector")
	logConfig(config)
	g, err := NewCollector(config)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/gratia-servlets/rmi", g)

	log.WithFields(log.Fields{
		"address": config.Address,
		"port":    config.Port,
	}).Info("starting HTTP server")

	log.Fatal(http.ListenAndServe(config.Address+":"+config.Port, nil))
}

func logConfig(config *CollectorConfig) {
	if config.File.Enabled {
		log.WithFields(log.Fields{
			"path":   config.File.Path,
			"format": config.File.Format,
		}).Info("file output enabled")
	} else {
		log.Info("file output diabled")
	}
	if config.Elasticsearch.Enabled {
		log.WithFields(log.Fields{
			"host":  config.Elasticsearch.Host,
			"index": config.Elasticsearch.Index,
		}).Info("elasticsearch output enabled")
	} else {
		log.Info("elasticsearch output disabled")
	}
	if config.Kafka.Enabled {
		log.WithFields(log.Fields{
			"brokers": config.Kafka.Brokers,
			"topic":   config.Kafka.Topic,
		}).Info("kafka output enabled")
	} else {
		log.Info("kafka output disabled")
	}
}

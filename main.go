package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"os"
)

// build parameters
var (
	buildDate string
	commit    string
)

// flags
var (
	configFile string
	logFile    string
)

func init() {
	flag.StringVar(&configFile, "c", "gratia.cfg", "config file")
	flag.StringVar(&logFile, "l", "stderr", "log file: stdout, stderr, or file name")
}

func main() {
	flag.Parse()

	// need to set log output first, since we log everything else
	switch logFile {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stdout)
	default:
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		log.SetOutput(f)
		//log.SetFormatter(&log.JSONFormatter{})
		log.SetFormatter(&log.TextFormatter{DisableColors: true})
	}

	log.WithFields(log.Fields{
		"build":  buildDate,
		"commit": commit,
	}).Info("gratia2")

	log.WithField("file", configFile).Info("reading config")
	config, err := ReadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"level": config.LogLevel,
		"file":  logFile,
	}).Info("initializing logging")
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

	log.WithFields(log.Fields{
		"address": config.Address,
		"port":    config.Port,
	}).Info("starting HTTP server")
	http.Handle("/gratia-servlets/rmi", g)
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

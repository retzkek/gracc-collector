package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// build parameters
var (
	build_ver  = "0.02.03"
	build_date = "???"
	build_ref  = "scratch"
)

// flags
var (
	configFile string
	logFile    string
)

func init() {
	flag.StringVar(&configFile, "c", "gracc.cfg", "config file")
	flag.StringVar(&logFile, "l", "stderr", "log file: stdout, stderr, or file name")
}

func main() {
	flag.Parse()

	// need to set log output first, since we log everything else
	var flog *os.File
	var err error
	switch logFile {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stdout)
	default:
		flog, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.Fatal(err)
		}
		defer flog.Close()
		log.SetOutput(flog)
		//log.SetFormatter(&log.JSONFormatter{})
		log.SetFormatter(&log.TextFormatter{DisableColors: true})
	}

	log.WithFields(log.Fields{
		"version": build_ver,
		"build":   build_date,
		"ref":     build_ref,
	}).Info("GRÅCC")

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
	g, err := NewCollector(config)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"address": config.Address,
		"port":    config.Port,
	}).Info("starting HTTP server")
	http.Handle("/gratia-servlets/rmi", g)
	http.HandleFunc("/stats", g.ServeStats)
	go http.ListenAndServe(config.Address+":"+config.Port, nil)

	// loop to catch signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)
MainLoop:
	for {
		select {
		case s := <-signals:
			log.WithField("signal", s).Debug("got signal")
			switch s {
			case os.Interrupt, syscall.SIGTERM:
				// terminate
				log.WithField("signal", s).Info("exiting")
				break MainLoop
			case syscall.SIGUSR1, syscall.SIGHUP:
				// refresh log file
				if flog != nil {
					log.WithField("signal", s).Info("closing log")
					flog.Close()
					flog, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
					if err != nil {
						log.Fatal(err)
					}
					defer flog.Close()
					log.SetOutput(flog)
				}
			case syscall.SIGUSR2:
				// toggle log level between debug and config option
				if log.GetLevel() != logLevel {
					log.SetLevel(logLevel)
				} else {
					log.SetLevel(log.DebugLevel)
				}
				log.WithFields(log.Fields{
					"signal": s,
					"level":  log.GetLevel().String(),
				}).Info("changing log level")
			}
		}
	}
}

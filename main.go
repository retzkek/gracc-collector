package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
)

const (
	HOST     = ""
	PORT     = "8080"
	ES_HOST  = "http://fermicloud080.fnal.gov:9200"
	ES_INDEX = "gratia-test2"
)

func main() {
	g, err := NewCollector(ES_HOST, ES_INDEX)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/gratia-servlets/rmi", g)

	log.WithFields(log.Fields{
		"host": HOST,
		"port": PORT,
	}).Info("listening")

	log.Fatal(http.ListenAndServe(HOST+":"+PORT, nil))
}

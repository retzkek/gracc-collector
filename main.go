package main

import (
	log "github.com/Sirupsen/logrus"
	"net/http"
)

const (
	HOST = ""
	PORT = "8080"
)

func main() {
	var g GratiaCollector

	http.Handle("/gratia-servlets/rmi", g)

	log.WithFields(log.Fields{
		"host": HOST,
		"port": PORT,
	}).Info("listening")

	log.Fatal(http.ListenAndServe(HOST+":"+PORT, nil))
}

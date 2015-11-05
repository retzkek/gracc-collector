package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"html"
	"net/http"
	"strconv"
	"strings"
)

func ProcessBundle(bundle string, bundlesize string) error {
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
			if err := ProcessXml(parts[i+1]); err != nil {
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

func ProcessXml(xml string) error {
	//log.Info(xml)
	return nil
}

func rmiHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	command := r.Form["command"][0]
	if command == "update" {
		updateLogger := log.WithFields(log.Fields{
			"from": r.Form["from"][0],
		})
		if r.Form["arg1"][0] == "xxx" {
			updateLogger.Info("received test request")
			fmt.Fprintf(w, "OK")
		} else {
			updateLogger.Info("received update")
			if err := ProcessBundle(r.Form["arg1"][0], r.Form["bundlesize"][0]); err == nil {
				fmt.Fprintf(w, "OK")
			} else {
				updateLogger.Error("error processing bundle")
				updateLogger.Error(err)
				fmt.Fprintf(w, "Error")
			}
		}
	} else {
		log.Error("received unrecognized or misformed request")
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/gratia-servlets/rmi", rmiHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

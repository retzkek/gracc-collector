package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

type GratiaCollector struct {
}

func (g GratiaCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Errorf("recieved unknown request %v", r.Form["command"])
	if len(r.Form["command"]) == 0 {
		log.Errorf("recieved unknown request %v", r)
		return
	}
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
			if err := g.ProcessBundle(r.Form["arg1"][0], r.Form["bundlesize"][0]); err == nil {
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

func (g *GratiaCollector) ProcessBundle(bundle string, bundlesize string) error {
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

func (g *GratiaCollector) ProcessXml(x string) error {
	var v JobUsageRecord
	if err := xml.Unmarshal([]byte(x), &v); err != nil {
		return err
	}

	if j, err := json.MarshalIndent(v.Flatten(), "", "    "); err != nil {
		return err
	} else {
		log.Infof("%s", j)
	}

	return nil
}

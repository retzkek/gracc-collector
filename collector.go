package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	elastic "gopkg.in/olivere/elastic.v3"
	"net/http"
	"strconv"
	"strings"
)

type GratiaCollector struct {
	Client *elastic.Client
	Index  string
}

func NewCollector(esHost string, esIndex string) (*GratiaCollector, error) {
	var g GratiaCollector
	client, err := elastic.NewClient(elastic.SetURL(esHost))
	if err != nil {
		return nil, err
	}
	g.Client = client
	g.Index = esIndex
	return &g, nil
}

func (g GratiaCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (g GratiaCollector) handleUpdate(w http.ResponseWriter, r *http.Request) {
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
			updateLogger.Info("received update")
			fmt.Fprintf(w, "OK")
		} else {
			g.handleError(w, r, "error processing bundle")
			log.Debug(err)
			return
		}
	}
}

func (g GratiaCollector) checkRequiredKeys(w http.ResponseWriter, r *http.Request, keys []string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			err := fmt.Sprintf("no %v", k)
			g.handleError(w, r, err)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (g GratiaCollector) handleError(w http.ResponseWriter, r *http.Request, err string) {
	log.WithField("error", err).Errorf("recieved unknown or misformed request")
	log.Debug(r)
	fmt.Fprintf(w, "Error")
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
		log.Debugf("%s", j)
		_, err := g.Client.Index().Index(g.Index).Type("JobUsageRecord").BodyString(string(j[:])).Do()
		if err != nil {
			return err
		}
	}

	return nil
}

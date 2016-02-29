package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"hash/fnv"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
)

type GratiaCollector struct {
	Path         string
	PathTemplate *template.Template
	Format       string
}

func NewCollector(path string, format string) (*GratiaCollector, error) {
	var g GratiaCollector
	g.Path = path
	var err error
	g.PathTemplate, err = template.New("path").Parse(path)
	if err != nil {
		return nil, err
	}
	g.Format = format
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
			updateLogger.WithField("size", r.FormValue("bundlesize")).Info("received update")
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
	//fmt.Println("---+++---")
	//fmt.Print(bundle)
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
	xb := []byte(x)
	if err := xml.Unmarshal(xb, &v); err != nil {
		return err
	}

	var path bytes.Buffer
	// generate path for record from template
	if err := g.PathTemplate.Execute(&path, v); err != nil {
		return err
	}
	// hash record ID to create file name and append to path
	h := fnv.New32()
	h.Write([]byte(v.RecordIdentity.RecordId))
	fmt.Fprintf(&path, "%x.%s", h.Sum32(), g.Format)

	switch g.Format {
	case "xml":
		if err := dumpToFile(path.String(), xb); err != nil {
			log.Debugf("error writing xml: %s", x)
			return err
		}
	case "json":
		if j, err := json.MarshalIndent(v.Flatten(), "", "    "); err != nil {
			log.Debugf("error converting JobUsageRecord to json: %s", x)
			return err
		} else {
			if err := dumpToFile(path.String(), j); err != nil {
				log.Debugf("error writing json: %s", x)
				return err
			}
		}
	}
	return nil
}

func dumpToFile(filepath string, contents []byte) error {
	dirname := path.Dir(filepath)
	filename := path.Base(filepath)
	log.WithField("path", dirname).Debug("creating directory")
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return err
	}
	log.WithField("filename", filename).Debug("writing record to file")
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(contents)
	log.WithFields(log.Fields{
		"filename": filepath,
		"bytes":    n,
	}).Info("wrote record to file")
	return err
}

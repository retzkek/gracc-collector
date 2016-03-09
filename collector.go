package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	elastic "gopkg.in/olivere/elastic.v3"
	"hash/fnv"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
)

type GraccCollector struct {
	Config *CollectorConfig
	// file
	PathTemplate *template.Template
	// elasticsearch
	Client *elastic.Client
	// kafka
	Producer sarama.SyncProducer
}

// NewCollector initializes and returns a new Gracc collector.
func NewCollector(conf *CollectorConfig) (*GraccCollector, error) {
	var g GraccCollector
	g.Config = conf

	var err error
	if conf.File.Enabled {
		g.PathTemplate, err = template.New("path").Parse(conf.File.Path)
		if err != nil {
			return nil, err
		}
	}
	if conf.Elasticsearch.Enabled {
		// initialize elasticsearch client
		g.Client, err = elastic.NewClient(elastic.SetURL(conf.Elasticsearch.Host))
		if err != nil {
			return nil, err
		}
		// create index
		if err = g.CreateIndex(); err != nil {
			return nil, err
		}
	}
	if conf.Kafka.Enabled {
		g.Producer, err = sarama.NewSyncProducer(conf.Kafka.Brokers, nil)
		if err != nil {
			return nil, err
		}
	}

	return &g, nil
}

// CreateIndex initializes the Elasticsearch index, if it doesn't already exist.
func (g GraccCollector) CreateIndex() error {
	const createBody = `
{
    "mappings": {
        "JobUsageRecord": {
            "properties": {
                "Charge": {
                    "type": "float"
                },
                "CreateTime": {
                    "type": "date"
                },
                "StartTime": {
                    "type": "date"
                },
                "EndTime": {
                    "type": "date"
                },
                "CpuSystemDuration": {
                    "type": "float"
                },
                "CpuUserDuration": {
                    "type": "float"
                },
                "Memory": {
                    "type": "float"
                },
                "NodeCount": {
                    "type": "integer"
                },
                "Processors": {
                    "type": "integer"
                },
                "WallDuration": {
                    "type": "float"
                }
            },
            "dynamic_templates": [
                { "notanalyzed": {
                         "match":              "*", 
                         "match_mapping_type": "string",
                         "mapping": {
                             "type":        "string",
                             "index":       "not_analyzed"
                         }
                     }
                }
            ]
        }
    }
}`
	exists, err := g.Client.IndexExists(g.Config.Elasticsearch.Index).Do()
	if err != nil {
		return err
	}
	if !exists {
		_, err := g.Client.CreateIndex(g.Config.Elasticsearch.Index).Body(createBody).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g GraccCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (g GraccCollector) handleUpdate(w http.ResponseWriter, r *http.Request) {
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

func (g GraccCollector) checkRequiredKeys(w http.ResponseWriter, r *http.Request, keys []string) error {
	for _, k := range keys {
		if r.FormValue(k) == "" {
			err := fmt.Sprintf("no %v", k)
			g.handleError(w, r, err)
			return fmt.Errorf(err)
		}
	}
	return nil
}

func (g GraccCollector) handleError(w http.ResponseWriter, r *http.Request, err string) {
	log.WithField("error", err).Errorf("recieved unknown or misformed request")
	log.Debug(r)
	fmt.Fprintf(w, "Error")
}

func (g *GraccCollector) ProcessBundle(bundle string, bundlesize string) error {
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

func (g *GraccCollector) ProcessXml(x string) error {
	var jur gracc.JobUsageRecord
	xb := []byte(x)
	if err := xml.Unmarshal(xb, &jur); err != nil {
		return err
	}
	if g.Config.File.Enabled {
		if err := g.RecordToFile(&jur); err != nil {
			return err
		}
	}
	if g.Config.Elasticsearch.Enabled {
		if err := g.RecordToElasticsearch(&jur); err != nil {
			return err
		}
	}
	if g.Config.Kafka.Enabled {
		if err := g.RecordToKafka(&jur); err != nil {
			return err
		}
	}

	return nil
}

func (g *GraccCollector) RecordToFile(jur *gracc.JobUsageRecord) error {
	var basePath, filename bytes.Buffer
	var filePath string
	// generate path for record from template
	if err := g.PathTemplate.Execute(&basePath, jur); err != nil {
		return err
	}
	// hash record ID to create file name and append to path
	h := fnv.New32()
	for {
		// keep writing to hash until unique filename is obtained
		h.Write([]byte(jur.RecordIdentity.RecordId))
		fmt.Fprintf(&filename, "%x.%s", h.Sum32(), g.Config.File.Format)
		filePath = path.Join(basePath.String(), filename.String())
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// file (or directory) doesn't exist; it will be created later
			break
		}
		log.WithField("filename", filename.String()).Debug("file exists, adding to hash")
		filename.Reset()
	}

	switch g.Config.File.Format {
	case "xml":
		if j, err := xml.MarshalIndent(jur, "", "    "); err != nil {
			log.Error("error converting JobUsageRecord to xml")
			log.Debugf("%v", jur)
			return err
		} else {
			if err := dumpToFile(filePath, j); err != nil {
				log.Error("error writing xml to file")
				return err
			}
		}
	case "json":
		if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
			log.Error("error converting JobUsageRecord to json")
			log.Debugf("%v", jur)
			return err
		} else {
			if err := dumpToFile(filePath, j); err != nil {
				log.Error("error writing json to file")
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
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(contents)
	log.WithFields(log.Fields{
		"filename": filepath,
		"bytes":    n,
	}).Debug("wrote record to file")
	return err
}

func (g *GraccCollector) RecordToElasticsearch(jur *gracc.JobUsageRecord) error {
	if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
		log.Error("error converting JobUsageRecord to json")
		log.Debugf("%v", jur)
		return err
	} else {
		_, err := g.Client.Index().
			Index(g.Config.Elasticsearch.Index).
			Type("JobUsageRecord").
			BodyString(string(j[:])).
			Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GraccCollector) RecordToKafka(jur *gracc.JobUsageRecord) error {
	if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
		log.Error("error converting JobUsageRecord to json")
		log.Debugf("%v", jur)
		return err
	} else {
		msg := &sarama.ProducerMessage{Topic: g.Config.Kafka.Topic, Value: sarama.ByteEncoder(j)}
		_, _, err := g.Producer.SendMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

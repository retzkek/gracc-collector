package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	"github.com/streadway/amqp"
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
	Config  *CollectorConfig
	Outputs []GraccOutput
}

type GraccOutput interface {
	OutputJUR(*gracc.JobUsageRecord) error
}

type FileOutput struct {
	Config       fileConfig
	PathTemplate *template.Template
}

type ElasticsearchOutput struct {
	Config esConfig
	Client *elastic.Client
}

type KafkaOutput struct {
	Config   kafkaConfig
	Producer sarama.SyncProducer
}

type AMQPOutput struct {
	Config     amqpConfig
	Connection *amqp.Connection
}

// NewCollector initializes and returns a new Gracc collector.
func NewCollector(conf *CollectorConfig) (*GraccCollector, error) {
	var g GraccCollector
	g.Config = conf
	g.Outputs = make([]GraccOutput, 0, 4)

	var err error
	if conf.File.Enabled {
		var f *FileOutput
		if f, err = InitFile(conf.File); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, f)
	}
	if conf.Elasticsearch.Enabled {
		var e *ElasticsearchOutput
		if e, err = InitElasticsearch(conf.Elasticsearch); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, e)
	}
	if conf.Kafka.Enabled {
		var k *KafkaOutput
		if k, err = InitKafka(conf.Kafka); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, k)
	}
	if conf.AMQP.Enabled {
		var a *AMQPOutput
		if a, err = InitAMQP(conf.AMQP); err != nil {
			return nil, err
		}
		g.Outputs = append(g.Outputs, a)
	}

	return &g, nil
}

func InitFile(conf fileConfig) (*FileOutput, error) {
	var f = &FileOutput{Config: conf}
	var err error
	f.PathTemplate, err = template.New("path").Parse(conf.Path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func InitElasticsearch(conf esConfig) (*ElasticsearchOutput, error) {
	var e = &ElasticsearchOutput{Config: conf}
	var err error
	log.WithField("host", conf.Host).Info("initializing Elasticsearch client")
	e.Client, err = elastic.NewClient(elastic.SetURL(conf.Host))
	if err != nil {
		return nil, err
	}

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
	exists, err := e.Client.IndexExists(conf.Index).Do()
	if err != nil {
		return nil, err
	}
	if !exists {
		log.WithField("inde", conf.Index).Info("creating mapping for Elasticsearch index")
		_, err := e.Client.CreateIndex(conf.Index).Body(createBody).Do()
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}

func InitKafka(conf kafkaConfig) (*KafkaOutput, error) {
	var k = &KafkaOutput{Config: conf}
	var err error
	log.WithField("brokers", conf.Brokers).Info("initializing Kafka producer")
	if k.Producer, err = sarama.NewSyncProducer(conf.Brokers, nil); err != nil {
		return nil, err
	}
	return k, nil
}

func InitAMQP(conf amqpConfig) (*AMQPOutput, error) {
	var a = &AMQPOutput{Config: conf}
	url := "amqp://" + conf.User + ":" + conf.Password + "@" +
		conf.Host + ":" + conf.Port + "/" + conf.Vhost
	log.WithFields(log.Fields{
		"user": conf.User,
		"host": conf.Host,
		"port": conf.Port,
	}).Info("connecting to RabbitMQ")
	var err error
	if a.Connection, err = amqp.Dial(url); err != nil {
		return nil, err
	}
	return a, nil
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
	for _, o := range g.Outputs {
		if err := o.OutputJUR(&jur); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	var basePath, filename bytes.Buffer
	var filePath string
	// generate path for record from template
	if err := f.PathTemplate.Execute(&basePath, jur); err != nil {
		return err
	}
	// hash record ID to create file name and append to path
	h := fnv.New32()
	for {
		// keep writing to hash until unique filename is obtained
		h.Write([]byte(jur.RecordIdentity.RecordId))
		fmt.Fprintf(&filename, "%x.%s", h.Sum32(), f.Config.Format)
		filePath = path.Join(basePath.String(), filename.String())
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// file (or directory) doesn't exist; it will be created later
			break
		}
		log.WithField("filename", filename.String()).Debug("file exists, adding to hash")
		filename.Reset()
	}

	switch f.Config.Format {
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

func (e *ElasticsearchOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
		log.Error("error converting JobUsageRecord to json")
		log.Debugf("%v", jur)
		return err
	} else {
		_, err := e.Client.Index().
			Index(e.Config.Index).
			Type("JobUsageRecord").
			BodyString(string(j[:])).
			Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *KafkaOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
		log.Error("error converting JobUsageRecord to json")
		log.Debugf("%v", jur)
		return err
	} else {
		msg := &sarama.ProducerMessage{Topic: k.Config.Topic, Value: sarama.ByteEncoder(j)}
		_, _, err := k.Producer.SendMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AMQPOutput) OutputJUR(jur *gracc.JobUsageRecord) error {
	return fmt.Errorf("not implemented")
}

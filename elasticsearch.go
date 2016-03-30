package main

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
	elastic "gopkg.in/olivere/elastic.v3"
)

type ElasticsearchConfig struct {
	Enabled bool
	Host    string
	Index   string
}

type ElasticsearchOutput struct {
	Config     ElasticsearchConfig
	Client     *elastic.Client
	outputChan chan gracc.Record
}

func InitElasticsearch(conf ElasticsearchConfig) (*ElasticsearchOutput, error) {
	var e = &ElasticsearchOutput{Config: conf}
	var err error
	log.WithField("host", conf.Host).Info("initializing Elasticsearch client")
	e.Client, err = elastic.NewClient(elastic.SetURL(conf.Host))
	if err != nil {
		return nil, err
	}

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
	e.outputChan = make(chan gracc.Record, 10)
	go e.OutputRecords()
	return e, nil
}

func (e *ElasticsearchOutput) Type() string {
	return "elasticsearch"
}

func (e *ElasticsearchOutput) OutputChan() chan gracc.Record {
	return e.outputChan
}

func (e *ElasticsearchOutput) OutputRecords() {
	for jur := range e.outputChan {
		if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
			log.Error("error converting JobUsageRecord to json")
			log.Debugf("%v", jur)
			//return err
		} else {
			_, err := e.Client.Index().
				Index(e.Config.Index).
				Type("JobUsageRecord").
				BodyString(string(j[:])).
				Do()
			if err != nil {
				log.Error(err)
				//return err
			}
		}
	}
	//return nil
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

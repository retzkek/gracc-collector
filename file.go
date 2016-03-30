package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash/fnv"
	"os"
	"path"
	"text/template"

	log "github.com/Sirupsen/logrus"
	gracc "github.com/gracc-project/gracc-go"
)

type FileConfig struct {
	Enabled bool
	Path    string
	Format  string
}

type FileOutput struct {
	Config       FileConfig
	PathTemplate *template.Template
	outputChan   chan gracc.Record
}

func InitFile(conf FileConfig) (*FileOutput, error) {
	var f = &FileOutput{Config: conf}
	var err error
	f.PathTemplate, err = template.New("path").Parse(conf.Path)
	if err != nil {
		return nil, err
	}
	f.outputChan = make(chan gracc.Record, 10)
	go f.OutputRecords()
	return f, nil
}

func (f *FileOutput) Type() string {
	return "file"
}

func (f *FileOutput) OutputChan() chan gracc.Record {
	return f.outputChan
}

func (f *FileOutput) OutputRecords() {
	for jur := range f.outputChan {
		var basePath, filename bytes.Buffer
		var filePath string
		// generate path for record from template
		if err := f.PathTemplate.Execute(&basePath, jur); err != nil {
			log.Error(err)
			//return err
		}
		// hash record ID to create file name and append to path
		h := fnv.New32()
		for {
			// keep writing to hash until unique filename is obtained
			h.Write([]byte(jur.Id()))
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
				//return err
			} else {
				if err := dumpToFile(filePath, j); err != nil {
					log.Error("error writing xml to file")
					//return err
				}
			}
		case "json":
			if j, err := json.MarshalIndent(jur.Flatten(), "", "    "); err != nil {
				log.Error("error converting JobUsageRecord to json")
				log.Debugf("%v", jur)
				//return err
			} else {
				if err := dumpToFile(filePath, j); err != nil {
					log.Error("error writing json to file")
					//return err
				}
			}
		}
	}
	//return nil
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

package gracc

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	duration "github.com/retzkek/iso8601duration"
)

// XMLRecord is a generic structure for unmarshalling unknown XML data.
type XMLRecord struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}

// Record is an interface implemented by all record types.
type Record interface {
	Id() string
	Type() string
	ToJSON(indent string) ([]byte, error)
	Raw() []byte
}

// ParseRecordXML will attempt to unmarshall the XML in buf into one of the
// known record types.
func ParseRecordXML(buf []byte) (Record, error) {
	var rec XMLRecord
	if err := xml.Unmarshal(buf, &rec); err != nil {
		return nil, fmt.Errorf("unable to parse record XML")
	}
	switch rec.XMLName.Local {
	case "UsageRecord", "JobUsageRecord":
		var jur JobUsageRecord
		if err := jur.ParseXML(buf); err == nil {
			return &jur, nil
		}
	case "StorageElement":
		var se StorageElement
		if err := se.ParseXML(buf); err == nil {
			return &se, nil
		}
	case "StorageElementRecord":
		var ser StorageElementRecord
		if err := ser.ParseXML(buf); err == nil {
			return &ser, nil
		}
	}
	return nil, fmt.Errorf("unable to unmarshall XML into record")
}

type origin struct {
	Hop        int       `xml:"hop,attr"`
	ServerDate time.Time `xml:",omitempty"`
	SenderHost string    `xml:"Connection>SenderHost,omitempty"`
	Sender     string    `xml:"Connection>Sender,omitempty"`
	Collector  string    `xml:"Connection>Collector,omitempty"`
}

func (o *origin) flatten() map[string]interface{} {
	var r = make(map[string]interface{})
	if o.Hop > 0 {
		r["Origin_hop"] = o.Hop
	}
	if !o.ServerDate.IsZero() {
		r["OriginServerDate"] = o.ServerDate.Format(time.RFC3339)
	}
	if o.SenderHost != "" {
		r["OriginSenderHost"] = o.SenderHost
	}
	if o.Sender != "" {
		r["OriginSender"] = o.Sender
	}
	if o.Collector != "" {
		r["OriginCollector"] = o.Collector
	}

	return r
}

type field struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"description,attr,omitempty"`
	Unit        string `xml:"unit,attr,omitempty"`
	PhaseUnit   string `xml:"phaseUnit,attr,omitempty"`
	StorageUnit string `xml:"storageUnit,attr,omitempty"`
	Formula     string `xml:"formula,attr,omitempty"`
	Metric      string `xml:"metric,attr,omitempty"`
}

func (f *field) flatten() map[string]interface{} {
	var r = make(map[string]interface{})
	if f.Value != "" {
		r[f.XMLName.Local] = f.Value
	}
	if f.Description != "" {
		r[f.XMLName.Local+"_description"] = f.Description
	}
	if f.Unit != "" {
		r[f.XMLName.Local+"_unit"] = f.Unit
	}
	if f.PhaseUnit != "" {
		r[f.XMLName.Local+"_phaseUnit"] = convertDurationToSeconds(f.PhaseUnit)
	}
	if f.StorageUnit != "" {
		r[f.XMLName.Local+"_storageUnit"] = f.StorageUnit
	}
	if f.Formula != "" {
		r[f.XMLName.Local+"_formula"] = f.Formula
	}
	if f.Metric != "" {
		r[f.XMLName.Local+"_metric"] = f.Metric
	}
	return r
}

type timeDuration struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"description,attr"`
	Type        string `xml:"type,attr,omitempty"`
}

func (t *timeDuration) flatten() map[string]interface{} {
	k := "unknown"
	if t.Type != "" {
		k = strings.Map(mapForKey, t.Type)
	}
	var rr = make(map[string]interface{})
	rr[k] = convertDurationToSeconds(t.Value)
	if t.Description != "" {
		rr[k+"_description"] = t.Description
	}
	return rr
}

type timeInstant struct {
	XMLName     xml.Name
	Value       time.Time `xml:",chardata"`
	Description string    `xml:"description,attr"`
	Type        string    `xml:"type,attr,omitempty"`
}

func (t *timeInstant) flatten() map[string]interface{} {
	k := "unknown"
	if t.Type != "" {
		k = strings.Map(mapForKey, t.Type)
	}
	var rr = map[string]interface{}{
		k: t.Value.Format(time.RFC3339),
	}
	if t.Description != "" {
		rr[k+"_description"] = t.Description
	}
	return rr
}

func mapForKey(c rune) rune {
	switch c {
	case '.', ' ':
		return '-'
	}
	return c
}

func convertDurationToSeconds(dur string) float64 {
	d, err := duration.ParseString(dur)
	if err != nil {
		return 0.0
	}
	return d.ToDuration().Seconds()
}

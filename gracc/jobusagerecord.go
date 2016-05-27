package gracc

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"time"

	duration "github.com/retzkek/iso8601duration"
)

type recordIdentity struct {
	RecordId   string    `xml:"recordId,attr"`
	CreateTime time.Time `xml:"createTime,attr"`
}

type jobIdentity struct {
	GlobalJobId string
	LocalJobId  string
	ProcessId   []string
}

type userIdentity struct {
	GlobalUsername   string
	LocalUserId      string
	VOName           string
	ReportableVOName string
	CommonName       string
	DN               string
}

type field struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"description,attr,omitempty"`
	Unit        string `xml:"unit,attr,omitempty"`
	PhaseUnit   string `xml:"phaseUnit,attr,omitempty"`
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
	if f.Formula != "" {
		r[f.XMLName.Local+"_formula"] = f.Formula
	}
	if f.Metric != "" {
		r[f.XMLName.Local+"_metric"] = f.Metric
	}
	return r
}

type cpuDuration struct {
	UsageType   string `xml:"usageType,attr"`
	Description string `xml:"description,attr"`
	Value       string `xml:",chardata"`
}

type wallDuration struct {
	Description string `xml:"description,attr"`
	Value       string `xml:",chardata"`
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
	Type        string    `xml:"description,attr,omitempty"`
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

type resource struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"description,attr"`
	Unit        string `xml:"unit,attr,omitempty"`
	PhaseUnit   string `xml:"phaseUnit,attr,omitempty"`
	StorageUnit string `xml:"storageUnit,attr,omitempty"`
}

func (r *resource) flatten() map[string]interface{} {
	k := "unknown"
	if r.Description != "" {
		k = strings.Map(mapForKey, r.Description)
	}
	var rr = map[string]interface{}{
		k: r.Value,
	}
	if r.Unit != "" {
		rr[k+"_unit"] = r.Unit
	}
	if r.PhaseUnit != "" {
		rr[k+"_phaseUnit"] = convertDurationToSeconds(r.PhaseUnit)
	}
	if r.StorageUnit != "" {
		rr[k+"_storageUnit"] = r.StorageUnit
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

type JobUsageRecord struct {
	XMLName            xml.Name       `xml:"JobUsageRecord"`
	RecordIdentity     recordIdentity `xml:",omitempty"`
	JobIdentity        jobIdentity    `xml:",omitempty"`
	UserIdentity       userIdentity   `xml:",omitempty"`
	WallDuration       wallDuration   `xml:",omitempty"`
	CpuDuration        []cpuDuration  `xml:",omitempty"`
	StartTime          time.Time      `xml:",omitempty"`
	EndTime            time.Time      `xml:",omitempty"`
	TimeDuration       []timeDuration `xml:",omitempty"`
	TimeInstant        []timeInstant  `xml:",omitempty"`
	Resource           []resource     `xml:",omitempty"`
	ConsumableResource []resource     `xml:",omitempty"`
	PhaseResource      []resource     `xml:",omitempty"`
	VolumeResource     []resource     `xml:",omitempty"`
	Fields             []field        `xml:",any"`
	raw                []byte         `xml:"-"`
}

func (jur *JobUsageRecord) ParseXML(xb []byte) error {
	if err := xml.Unmarshal(xb, jur); err != nil {
		return err
	}
	jur.raw = append(jur.raw, xb...) // copy contents
	return nil
}

// Id returns an identification string for the record.
func (jur *JobUsageRecord) Id() string {
	return jur.RecordIdentity.RecordId
}

// Raw returns the unaltered source of the record.
func (jur *JobUsageRecord) Raw() []byte {
	return jur.raw
}

// ToJSON returns a JSON encoding of the Record, with certain elements
// transformed to fit the GRACC Raw Record schema.
// Indent specifies the string to use for each indentation level,
// if empty no indentation or pretty-printing is performed.
func (jur *JobUsageRecord) ToJSON(indent string) ([]byte, error) {
	var r = map[string]interface{}{
		"RecordId":         jur.RecordIdentity.RecordId,
		"CreateTime":       jur.RecordIdentity.CreateTime.Format(time.RFC3339),
		"GlobalJobId":      jur.JobIdentity.GlobalJobId,
		"LocalJobId":       jur.JobIdentity.LocalJobId,
		"GlobalUsername":   jur.UserIdentity.GlobalUsername,
		"LocalUserId":      jur.UserIdentity.LocalUserId,
		"VOName":           jur.UserIdentity.VOName,
		"ReportableVOName": jur.UserIdentity.ReportableVOName,
		"CommonName":       jur.UserIdentity.CommonName,
		"DN":               jur.UserIdentity.DN,
		"StartTime":        jur.StartTime.Format(time.RFC3339),
		"EndTime":          jur.EndTime.Format(time.RFC3339),
		"RawXML":           string(jur.Raw()),
	}

	// convert durations
	var totalCpu float64
	for _, c := range jur.CpuDuration {
		secs := convertDurationToSeconds(c.Value)
		if secs > 0 {
			totalCpu += secs
		}
		if c.UsageType != "" {
			r["CpuDuration_"+c.UsageType] = secs
		}
		if c.Description != "" {
			r["CpuDuration_"+c.UsageType+"_description"] = c.Description
		}
	}
	r["CpuDuration"] = totalCpu
	r["WallDuration"] = convertDurationToSeconds(jur.WallDuration.Value)
	if jur.WallDuration.Description != "" {
		r["WpuDuration_description"] = jur.WallDuration.Description
	}

	// flatten resources
	var res = make(map[string]interface{})
	if len(jur.Resource) > 0 {
		for _, resa := range [][]resource{jur.Resource,
			jur.ConsumableResource,
			jur.PhaseResource,
			jur.VolumeResource,
		} {
			for _, r := range resa {
				for k, v := range r.flatten() {
					res[k] = v
				}
			}
		}
	}
	if len(res) > 0 {
		r["Resource"] = res
	}

	// time durations and instants
	var timedur = make(map[string]interface{})
	for _, td := range jur.TimeDuration {
		for k, v := range td.flatten() {
			timedur[k] = v
		}
	}
	if len(timedur) > 0 {
		r["TimeDuration"] = timedur
	}
	var timeinst = make(map[string]interface{})
	for _, td := range jur.TimeInstant {
		for k, v := range td.flatten() {
			timeinst[k] = v
		}
	}
	if len(timeinst) > 0 {
		r["TimeInstant"] = timeinst
	}

	// flatten other fields
	for _, f := range jur.Fields {
		for k, v := range f.flatten() {
			r[k] = v
		}
	}

	if indent != "" {
		return json.MarshalIndent(r, "", indent)
	}
	return json.Marshal(r)
}

func convertDurationToSeconds(dur string) float64 {
	d, err := duration.ParseString(dur)
	if err != nil {
		return 0.0
	}
	return d.ToDuration().Seconds()
}

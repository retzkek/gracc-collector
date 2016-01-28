package main

import (
	"encoding/xml"
	"fmt"
	duration "github.com/channelmeter/iso8601duration"
	"time"
)

type recordIdentity struct {
	RecordId   string    `xml:"recordId,attr"`
	CreateTime time.Time `xml:"createTime,attr"`
}

type jobIdentity struct {
	GlobalJobId string
	LocalJobId  string
}

type userIdentity struct {
	GlobalUsername   string
	LocalUserId      string
	VOName           string
	ReportableVOName string
	CommonName       string
}

type field struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"http://www.gridforum.org/2003/ur-wg description,attr"`
	Unit        string `xml:"http://www.gridforum.org/2003/ur-wg unit,attr"`
	Formula     string `xml:"http://www.gridforum.org/2003/ur-wg formula,attr"`
	Metric      string `xml:"http://www.gridforum.org/2003/ur-wg metric,attr"`
}

func (f *field) flatten() map[string]string {
	var r = make(map[string]string)
	if f.Value != "" {
		r[f.XMLName.Local] = f.Value
	}
	if f.Description != "" {
		r[f.XMLName.Local+"_description"] = f.Description
	}
	if f.Unit != "" {
		r[f.XMLName.Local+"_unit"] = f.Unit
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
	UsageType string `xml:"http://www.gridforum.org/2003/ur-wg usageType,attr"`
	Value     string `xml:",chardata"`
}

type resource struct {
	XMLName     xml.Name
	Value       string `xml:",chardata"`
	Description string `xml:"http://www.gridforum.org/2003/ur-wg description,attr"`
}

func (r *resource) flatten() map[string]string {
	var rr = map[string]string{
		r.Description: r.Value,
	}
	return rr
}

type JobUsageRecord struct {
	XMLName        xml.Name `xml:"JobUsageRecord"`
	RecordIdentity recordIdentity
	JobIdentity    jobIdentity
	UserIdentity   userIdentity
	CpuDuration    []cpuDuration
	StartTime      time.Time
	EndTime        time.Time
	Resource       []resource
	Fields         []field `xml:",any"`
}

// Flatten returns a flattened map of the Record.
func (jur *JobUsageRecord) Flatten() map[string]string {
	var r = map[string]string{
		"RecordId":         jur.RecordIdentity.RecordId,
		"CreateTime":       jur.RecordIdentity.CreateTime.Format(time.RFC3339),
		"GlobalJobId":      jur.JobIdentity.GlobalJobId,
		"LocalJobId":       jur.JobIdentity.LocalJobId,
		"GlobalUsername":   jur.UserIdentity.GlobalUsername,
		"LocalUserId":      jur.UserIdentity.LocalUserId,
		"VOName":           jur.UserIdentity.VOName,
		"ReportableVOName": jur.UserIdentity.ReportableVOName,
		"CommonName":       jur.UserIdentity.CommonName,
		"StartTime":        jur.StartTime.Format(time.RFC3339),
		"EndTime":          jur.EndTime.Format(time.RFC3339),
	}

	for _, res := range jur.Resource {
		for k, v := range res.flatten() {
			r[k] = v
		}
	}

	for _, f := range jur.Fields {
		for k, v := range f.flatten() {
			r[k] = v
		}
	}

	for _, c := range jur.CpuDuration {
		if c.UsageType == "user" {
			r["CpuUserDuration"] = c.Value
		} else if c.UsageType == "system" {
			r["CpuSystemDuration"] = c.Value
		}
	}

	r["CpuUserDuration"] = convertDurationToSeconds(r["CpuUserDuration"])
	r["CpuSystemDuration"] = convertDurationToSeconds(r["CpuSystemDuration"])
	r["WallDuration"] = convertDurationToSeconds(r["WallDuration"])

	return r
}

func convertDurationToSeconds(dur string) string {
	d, err := duration.FromString(dur)
	if err != nil {
		return dur
	}
	sec := d.ToDuration().Seconds()
	return fmt.Sprintf("%.0f", sec)
}

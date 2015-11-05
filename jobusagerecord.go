package main

import (
	"encoding/xml"
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
	Charge         field
	Status         string
	WallDuration   string
	CpuDuration    []cpuDuration
	NodeCount      field
	Processors     field
	StartTime      time.Time
	EndTime        time.Time
	MachineName    field
	SiteName       field
	SubmitHost     string
	ProjectName    string
	Memory         field
	Resource       []resource
	ProbeName      string
	Grid           string
}

func (jur *JobUsageRecord) asMap() map[string]string {
	var r = map[string]string{
		"RecordId":         jur.RecordIdentity.RecordId,
		"CreateTime":       jur.RecordIdentity.CreateTime.String(),
		"GlobalJobId":      jur.JobIdentity.GlobalJobId,
		"LocalJobId":       jur.JobIdentity.LocalJobId,
		"GlobalUsername":   jur.UserIdentity.GlobalUsername,
		"LocalUserId":      jur.UserIdentity.LocalUserId,
		"VOName":           jur.UserIdentity.VOName,
		"ReportableVOName": jur.UserIdentity.ReportableVOName,
		"CommonName":       jur.UserIdentity.CommonName,
		"Status":           jur.Status,
		"WallDuration":     jur.WallDuration,
	}

	for k, v := range jur.Charge.flatten() {
		r[k] = v
	}
	for k, v := range jur.NodeCount.flatten() {
		r[k] = v
	}
	for k, v := range jur.Processors.flatten() {
		r[k] = v
	}
	for k, v := range jur.MachineName.flatten() {
		r[k] = v
	}
	for k, v := range jur.SiteName.flatten() {
		r[k] = v
	}
	for k, v := range jur.Memory.flatten() {
		r[k] = v
	}
	for _, res := range jur.Resource {
		for k, v := range res.flatten() {
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
	return r
}

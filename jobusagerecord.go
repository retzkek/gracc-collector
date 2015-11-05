package main

import (
	"time"
)

type recordIdentity struct {
	RecordId   string `xml:"recordId,attr"`
	CreateTime string `xml:"createTime,attr"`
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
	Value       string `xml:",chardata"`
	Description string `xml:"http://www.gridforum.org/2003/ur-wg description,attr"`
	Unit        string `xml:"http://www.gridforum.org/2003/ur-wg unit,attr"`
	Formula     string `xml:"http://www.gridforum.org/2003/ur-wg formula,attr"`
	Metric      string `xml:"http://www.gridforum.org/2003/ur-wg metric,attr"`
}

type cpuDuration struct {
	UsageType string `xml:"http://www.gridforum.org/2003/ur-wg usageType,attr"`
	Value     string `xml:",chardata"`
}

type JobUsageRecord struct {
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
	Resource       []field
	ProbeName      string
	Grid           string
}

func (jur *JobUsageRecord) asMap() map[string]string {
	r := make(map[string]string)
	return r
}

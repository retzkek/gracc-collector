package main

import (
	"encoding/json"
	"encoding/xml"
	"testing"
)

func TestJURUnmarshal(t *testing.T) {
	var v JobUsageRecord
	if err := xml.Unmarshal([]byte(testRecord), &v); err != nil {
		t.Error(err)
	}
	if j, err := json.MarshalIndent(v, "", "    "); err != nil {
		t.Error(err)
	} else {
		t.Logf("%s", j)
	}

	t.Logf("\n---\n")

	if j, err := json.MarshalIndent(v.Flatten(), "", "    "); err != nil {
		t.Error(err)
	} else {
		t.Logf("%s", j)
	}

	t.Logf("\n\n")
}

var testRecord = `
<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
                xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="mac-126903.dhcp.fnal.gov:13842.1" urwg:createTime="2015-11-03T20:28:33Z" />
<JobIdentity>
<GlobalJobId >i-065c9ddf#1446582511.798504</GlobalJobId>
<LocalJobId >i-065c9ddf</LocalJobId>
</JobIdentity>
<UserIdentity>
        <GlobalUsername>nova-159067897602</GlobalUsername>
        <LocalUserId>aws account user</LocalUserId>
        <VOName>nova</VOName>
        <ReportableVOName>nova</ReportableVOName>
        <CommonName>nova-159067897602</CommonName>
</UserIdentity>
<Charge urwg:description="The spot price charged in last hour corresponding to launch time" urwg:unit="$" urwg:formula="$/instance hr" >0.0</Charge>
<Status >1</Status>
<WallDuration >PT1H</WallDuration>
<CpuDuration urwg:usageType="user" >PT1M5.31S</CpuDuration>
<CpuDuration urwg:usageType="system" >PT0S</CpuDuration>
<NodeCount urwg:metric="total" >1</NodeCount>
<Processors urwg:description="m3.medium" urwg:metric="total" >1</Processors>
<StartTime >2015-11-03T19:34:32Z</StartTime>
<EndTime >2015-11-03T20:34:32Z</EndTime>
<MachineName urwg:description="ami-a3263c93" >no Public ip as instance has been stopped</MachineName>
<SiteName >fermilab</SiteName>
<SubmitHost >no Private ip as instance has been terminated</SubmitHost>
<ProjectName >aws-no project name given</ProjectName>
<Memory urwg:phaseUnit="PT0S" urwg:metric="total" >3.75</Memory>
<Resource urwg:description="Version" >1.0</Resource>
<ProbeName >awsvm:kretzke-dev</ProbeName>
<Grid >OSG</Grid>
<Resource urwg:description="ResourceType" >AWSVM</Resource>
</JobUsageRecord>
`

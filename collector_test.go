package main

import (
	log "github.com/Sirupsen/logrus"
	"testing"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func TestProcessBundle_File(t *testing.T) {
	// Reference time: Mon Jan 2 15:04:05 -0700 MST 2006
	var testPath = `/tmp/gracc.test/{{.RecordIdentity.CreateTime.Format "2006/01/02/15"}}/`
	g, err := NewCollector(&CollectorConfig{
		File: FileConfig{
			Enabled: true,
			Path:    testPath,
			Format:  "xml",
		},
	})
	if err != nil {
		t.Error(err)
	}
	if err := g.ProcessBundle(testBundle, 10); err != nil {
		t.Error(err)
	}
}

var testBundle = `replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="fermicloud121.fnal.gov:30586.1" urwg:createTime="2016-02-25T02:37:01Z" />
<JobIdentity>
<GlobalJobId >condor.fermicloud121.fnal.gov#2391.0#1456365750</GlobalJobId>
<LocalJobId >2391</LocalJobId>
</JobIdentity>
<UserIdentity>
	<GlobalUsername>osg@fermicloud121.fnal.gov</GlobalUsername>
	<LocalUserId>osg</LocalUserId>
	<DN>/DC=com/DC=DigiCert-Grid/O=Open Science Grid/OU=People/CN=Marco Mambelli</DN>	<VOName>/osg/Role=NULL/Capability=NULL</VOName>
	<ReportableVOName>osg</ReportableVOName>
	<CommonName>/CN=Marco Mambelli</CommonName>
</UserIdentity>
<JobName >fermicloud121.fnal.gov#2391.0#1456365750</JobName>
<Status urwg:description="Condor Exit Status" >0</Status>
<WallDuration urwg:description="Was entered in seconds" >PT23M15.0S</WallDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user" >PT20.0S</CpuDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system" >PT0S</CpuDuration>
<NodeCount urwg:metric="max" >1</NodeCount>
<Processors urwg:metric="max" >4</Processors>
<StartTime urwg:description="Was entered in seconds" >2016-02-25T02:02:30Z</StartTime>
<EndTime urwg:description="Was entered in seconds" >2016-02-25T02:25:45Z</EndTime>
<MachineName >fermicloud121.fnal.gov</MachineName>
<SiteName >MMTEST-FC1-CE1</SiteName>
<SubmitHost >fermicloud121.fnal.gov</SubmitHost>
<Queue urwg:description="Condor&apos;s JobUniverse field" >5</Queue>
<Host >fermicloud111.fnal.gov (primary) </Host>
<Network urwg:phaseUnit="PT23M15.0S" urwg:storageUnit="b" urwg:metric="total" >0.0</Network>
<TimeDuration urwg:type="RemoteUserCpu" >PT20.0S</TimeDuration>
<TimeDuration urwg:type="LocalUserCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="RemoteSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="LocalSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="CumulativeSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedTime" >PT23M15.0S</TimeDuration>
<Resource urwg:description="CondorMyType" >Job</Resource>
<Resource urwg:description="ExitBySignal" >false</Resource>
<Resource urwg:description="ExitCode" >0</Resource>
<Resource urwg:description="condor.JobStatus" >4</Resource>
<ProbeName >condor:fermicloud121.fnal.gov</ProbeName>
<Grid urwg:description="GratiaJobOrigin = GRAM" >OSG</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:createTime="2016-02-25T02:37:01Z" urwg:recordId="fermicloud121.fnal.gov:30586.1"/>
<JobIdentity>
	<GlobalJobId>condor.fermicloud121.fnal.gov#2391.0#1456365750</GlobalJobId>
	<LocalJobId>2391</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>osg</LocalUserId>
	<GlobalUsername>osg@fermicloud121.fnal.gov</GlobalUsername>
	<DN>/DC=com/DC=DigiCert-Grid/O=Open Science Grid/OU=People/CN=Marco Mambelli</DN>
	<VOName>/osg/Role=NULL/Capability=NULL</VOName>
	<ReportableVOName>osg</ReportableVOName>
</UserIdentity>
	<JobName>fermicloud121.fnal.gov#2391.0#1456365750</JobName>
	<MachineName>fermicloud121.fnal.gov</MachineName>
	<SubmitHost>fermicloud121.fnal.gov</SubmitHost>
	<Status urwg:description="Condor Exit Status">0</Status>
	<WallDuration urwg:description="Was entered in seconds">PT23M15.0S</WallDuration>
	<TimeDuration urwg:type="RemoteUserCpu">PT20.0S</TimeDuration>
	<TimeDuration urwg:type="LocalUserCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="RemoteSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="LocalSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="CumulativeSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedTime">PT23M15.0S</TimeDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT0S</CpuDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT20.0S</CpuDuration>
	<EndTime urwg:description="Was entered in seconds">2016-02-25T02:25:45Z</EndTime>
	<StartTime urwg:description="Was entered in seconds">2016-02-25T02:02:30Z</StartTime>
	<Host primary="true">fermicloud111.fnal.gov</Host>
	<Queue urwg:description="Condor's JobUniverse field">5</Queue>
	<NodeCount urwg:metric="max">1</NodeCount>
	<Processors urwg:metric="max">4</Processors>
	<Resource urwg:description="CondorMyType">Job</Resource>
	<Resource urwg:description="ExitBySignal">false</Resource>
	<Resource urwg:description="ExitCode">0</Resource>
	<Resource urwg:description="condor.JobStatus">4</Resource>
	<Network urwg:metric="total" urwg:phaseUnit="PT23M15.0S" urwg:storageUnit="b">0</Network>
	<ProbeName>condor:fermicloud121.fnal.gov</ProbeName>
	<SiteName>MMTEST-FC1-CE1</SiteName>
	<Grid urwg:description="GratiaJobOrigin = GRAM">OSG</Grid>
	<Njobs>1</Njobs>
	<Resource urwg:description="ResourceType">Batch</Resource>
</JobUsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="gridtest02.racf.bnl.gov:31684.1" urwg:createTime="2016-02-25T02:44:47Z" />
<JobIdentity>
<GlobalJobId >condor.gridtest02.racf.bnl.gov#2922934.0#1456367107</GlobalJobId>
<LocalJobId >2922934</LocalJobId>
</JobIdentity>
<UserIdentity>
	<GlobalUsername>uscms@bnl.gov</GlobalUsername>
	<LocalUserId>uscms</LocalUserId>
	<DN>/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba</DN>	<VOName>/cms/Role=lcgadmin/Capability=NULL</VOName>
	<ReportableVOName>cms</ReportableVOName>
	<CommonName>/CN=sciaba/CN=430796/CN=Andrea Sciaba</CommonName>
</UserIdentity>
<JobName >gridtest02.racf.bnl.gov#2922934.0#1456367107</JobName>
<Status urwg:description="Condor Exit Status" >0</Status>
<WallDuration urwg:description="Was entered in seconds" >PT24.0S</WallDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system" >PT0S</CpuDuration>
<NodeCount urwg:metric="max" >1</NodeCount>
<Processors urwg:metric="max" >1</Processors>
<StartTime urwg:description="Was entered in seconds" >2016-02-25T02:31:35Z</StartTime>
<EndTime urwg:description="Was entered in seconds" >2016-02-25T02:31:59Z</EndTime>
<MachineName >gridtest02.racf.bnl.gov</MachineName>
<SiteName >BNL_Test_2_CE_1</SiteName>
<SubmitHost >gridtest02.racf.bnl.gov</SubmitHost>
<Queue urwg:description="Condor&apos;s JobUniverse field" >5</Queue>
<Host >acas0065.usatlas.bnl.gov (primary) </Host>
<Network urwg:phaseUnit="PT24.0S" urwg:storageUnit="b" urwg:metric="total" >0.0</Network>
<TimeDuration urwg:type="RemoteUserCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="LocalUserCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="RemoteSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="LocalSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="CumulativeSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedTime" >PT24.0S</TimeDuration>
<Resource urwg:description="CondorMyType" >Job</Resource>
<Resource urwg:description="ExitBySignal" >false</Resource>
<Resource urwg:description="ExitCode" >0</Resource>
<Resource urwg:description="condor.JobStatus" >4</Resource>
<ProbeName >condor:gridtest02.racf.bnl.gov</ProbeName>
<Grid >OSG</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:createTime="2016-02-25T02:44:47Z" urwg:recordId="gridtest02.racf.bnl.gov:31684.1"/>
<JobIdentity>
	<GlobalJobId>condor.gridtest02.racf.bnl.gov#2922934.0#1456367107</GlobalJobId>
	<LocalJobId>2922934</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>uscms</LocalUserId>
	<GlobalUsername>uscms@bnl.gov</GlobalUsername>
	<DN>/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba</DN>
	<VOName>/cms/Role=lcgadmin/Capability=NULL</VOName>
	<ReportableVOName>cms</ReportableVOName>
</UserIdentity>
	<JobName>gridtest02.racf.bnl.gov#2922934.0#1456367107</JobName>
	<MachineName>gridtest02.racf.bnl.gov</MachineName>
	<SubmitHost>gridtest02.racf.bnl.gov</SubmitHost>
	<Status urwg:description="Condor Exit Status">0</Status>
	<WallDuration urwg:description="Was entered in seconds">PT24.0S</WallDuration>
	<TimeDuration urwg:type="RemoteUserCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="LocalUserCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="RemoteSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="LocalSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="CumulativeSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedTime">PT24.0S</TimeDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT0S</CpuDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT0S</CpuDuration>
	<EndTime urwg:description="Was entered in seconds">2016-02-25T02:31:59Z</EndTime>
	<StartTime urwg:description="Was entered in seconds">2016-02-25T02:31:35Z</StartTime>
	<Host primary="true">acas0065.usatlas.bnl.gov</Host>
	<Queue urwg:description="Condor's JobUniverse field">5</Queue>
	<NodeCount urwg:metric="max">1</NodeCount>
	<Processors urwg:metric="max">1</Processors>
	<Resource urwg:description="CondorMyType">Job</Resource>
	<Resource urwg:description="ExitBySignal">false</Resource>
	<Resource urwg:description="ExitCode">0</Resource>
	<Resource urwg:description="condor.JobStatus">4</Resource>
	<Network urwg:metric="total" urwg:phaseUnit="PT24.0S" urwg:storageUnit="b">0</Network>
	<ProbeName>condor:gridtest02.racf.bnl.gov</ProbeName>
	<SiteName>BNL_Test_2_CE_1</SiteName>
	<Grid>OSG</Grid>
	<Njobs>1</Njobs>
	<Resource urwg:description="ResourceType">Batch</Resource>
</JobUsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="gridtest02.racf.bnl.gov:31684.2" urwg:createTime="2016-02-25T02:44:47Z" />
<JobIdentity>
<GlobalJobId >condor.gridtest02.racf.bnl.gov#2922935.0#1456367678</GlobalJobId>
<LocalJobId >2922935</LocalJobId>
</JobIdentity>
<UserIdentity>
	<GlobalUsername>uscms@bnl.gov</GlobalUsername>
	<LocalUserId>uscms</LocalUserId>
	<DN>/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba</DN>	<VOName>/cms/Role=pilot/Capability=NULL</VOName>
	<ReportableVOName>cms</ReportableVOName>
	<CommonName>/CN=sciaba/CN=430796/CN=Andrea Sciaba</CommonName>
</UserIdentity>
<JobName >gridtest02.racf.bnl.gov#2922935.0#1456367678</JobName>
<Status urwg:description="Condor Exit Status" >0</Status>
<WallDuration urwg:description="Was entered in seconds" >PT6.0S</WallDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system" >PT0S</CpuDuration>
<NodeCount urwg:metric="max" >1</NodeCount>
<Processors urwg:metric="max" >1</Processors>
<StartTime urwg:description="Was entered in seconds" >2016-02-25T02:39:35Z</StartTime>
<EndTime urwg:description="Was entered in seconds" >2016-02-25T02:39:41Z</EndTime>
<MachineName >gridtest02.racf.bnl.gov</MachineName>
<SiteName >BNL_Test_2_CE_1</SiteName>
<SubmitHost >gridtest02.racf.bnl.gov</SubmitHost>
<Queue urwg:description="Condor&apos;s JobUniverse field" >5</Queue>
<Host >acas0065.usatlas.bnl.gov (primary) </Host>
<Network urwg:phaseUnit="PT6.0S" urwg:storageUnit="b" urwg:metric="total" >0.0</Network>
<TimeDuration urwg:type="RemoteUserCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="LocalUserCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="RemoteSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="LocalSysCpu" >PT0S</TimeDuration>
<TimeDuration urwg:type="CumulativeSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedSuspensionTime" >PT0S</TimeDuration>
<TimeDuration urwg:type="CommittedTime" >PT6.0S</TimeDuration>
<Resource urwg:description="CondorMyType" >Job</Resource>
<Resource urwg:description="ExitBySignal" >false</Resource>
<Resource urwg:description="ExitCode" >0</Resource>
<Resource urwg:description="condor.JobStatus" >4</Resource>
<ProbeName >condor:gridtest02.racf.bnl.gov</ProbeName>
<Grid >OSG</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:createTime="2016-02-25T02:44:47Z" urwg:recordId="gridtest02.racf.bnl.gov:31684.2"/>
<JobIdentity>
	<GlobalJobId>condor.gridtest02.racf.bnl.gov#2922935.0#1456367678</GlobalJobId>
	<LocalJobId>2922935</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>uscms</LocalUserId>
	<GlobalUsername>uscms@bnl.gov</GlobalUsername>
	<DN>/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba</DN>
	<VOName>/cms/Role=pilot/Capability=NULL</VOName>
	<ReportableVOName>cms</ReportableVOName>
</UserIdentity>
	<JobName>gridtest02.racf.bnl.gov#2922935.0#1456367678</JobName>
	<MachineName>gridtest02.racf.bnl.gov</MachineName>
	<SubmitHost>gridtest02.racf.bnl.gov</SubmitHost>
	<Status urwg:description="Condor Exit Status">0</Status>
	<WallDuration urwg:description="Was entered in seconds">PT6.0S</WallDuration>
	<TimeDuration urwg:type="RemoteUserCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="LocalUserCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="RemoteSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="LocalSysCpu">PT0S</TimeDuration>
	<TimeDuration urwg:type="CumulativeSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedSuspensionTime">PT0S</TimeDuration>
	<TimeDuration urwg:type="CommittedTime">PT6.0S</TimeDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT0S</CpuDuration>
	<CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT0S</CpuDuration>
	<EndTime urwg:description="Was entered in seconds">2016-02-25T02:39:41Z</EndTime>
	<StartTime urwg:description="Was entered in seconds">2016-02-25T02:39:35Z</StartTime>
	<Host primary="true">acas0065.usatlas.bnl.gov</Host>
	<Queue urwg:description="Condor's JobUniverse field">5</Queue>
	<NodeCount urwg:metric="max">1</NodeCount>
	<Processors urwg:metric="max">1</Processors>
	<Resource urwg:description="CondorMyType">Job</Resource>
	<Resource urwg:description="ExitBySignal">false</Resource>
	<Resource urwg:description="ExitCode">0</Resource>
	<Resource urwg:description="condor.JobStatus">4</Resource>
	<Network urwg:metric="total" urwg:phaseUnit="PT6.0S" urwg:storageUnit="b">0</Network>
	<ProbeName>condor:gridtest02.racf.bnl.gov</ProbeName>
	<SiteName>BNL_Test_2_CE_1</SiteName>
	<Grid>OSG</Grid>
	<Njobs>1</Njobs>
	<Resource urwg:description="ResourceType">Batch</Resource>
</JobUsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158494.itbv-pbs.uc.mwt2.org_1456366578" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158494.itbv-pbs.uc.mwt2.org_1456366578</GlobalJobId>
<LocalJobId >158494.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_ab947099e65b</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:16:18Z</StartTime>
<EndTime >2016-02-25T02:36:18Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/0</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >4468.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >35032.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:36:18Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158494.itbv-pbs.uc.mwt2.org_1456366578"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158494.itbv-pbs.uc.mwt2.org_1456366578</urwg:GlobalJobId>
<urwg:LocalJobId>158494.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_ab947099e65b</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:36:18Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:16:18Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/0</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">4468</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">35032</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:36:18Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158495.itbv-pbs.uc.mwt2.org_1456366556" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158495.itbv-pbs.uc.mwt2.org_1456366556</GlobalJobId>
<LocalJobId >158495.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_75a8e87d1c9a</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:15:56Z</StartTime>
<EndTime >2016-02-25T02:35:56Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/7</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >5316.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >46436.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:35:56Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158495.itbv-pbs.uc.mwt2.org_1456366556"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158495.itbv-pbs.uc.mwt2.org_1456366556</urwg:GlobalJobId>
<urwg:LocalJobId>158495.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_75a8e87d1c9a</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:35:56Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:15:56Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/7</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">5316</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">46436</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:35:56Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158493.itbv-pbs.uc.mwt2.org_1456366556" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158493.itbv-pbs.uc.mwt2.org_1456366556</GlobalJobId>
<LocalJobId >158493.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_db247b7d8eb2</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:15:56Z</StartTime>
<EndTime >2016-02-25T02:35:57Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/8</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >5320.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >46428.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:35:57Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158493.itbv-pbs.uc.mwt2.org_1456366556"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158493.itbv-pbs.uc.mwt2.org_1456366556</urwg:GlobalJobId>
<urwg:LocalJobId>158493.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_db247b7d8eb2</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:35:57Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:15:56Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/8</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">5320</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">46428</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:35:57Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158491.itbv-pbs.uc.mwt2.org_1456366182" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158491.itbv-pbs.uc.mwt2.org_1456366182</GlobalJobId>
<LocalJobId >158491.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_33234016190f</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:09:42Z</StartTime>
<EndTime >2016-02-25T02:29:42Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/11</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >4460.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >35032.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:29:42Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158491.itbv-pbs.uc.mwt2.org_1456366182"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158491.itbv-pbs.uc.mwt2.org_1456366182</urwg:GlobalJobId>
<urwg:LocalJobId>158491.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_33234016190f</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:29:42Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:09:42Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/11</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">4460</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">35032</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:29:42Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158498.itbv-pbs.uc.mwt2.org_1456367161" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158498.itbv-pbs.uc.mwt2.org_1456367161</GlobalJobId>
<LocalJobId >158498.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_47222ab0898f</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M1.0S</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:26:01Z</StartTime>
<EndTime >2016-02-25T02:46:02Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/9</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >4464.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >35032.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:46:02Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158498.itbv-pbs.uc.mwt2.org_1456367161"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158498.itbv-pbs.uc.mwt2.org_1456367161</urwg:GlobalJobId>
<urwg:LocalJobId>158498.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_47222ab0898f</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M1S</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:46:02Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:26:01Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/9</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">4464</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">35032</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:46:02Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158497.itbv-pbs.uc.mwt2.org_1456367161" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158497.itbv-pbs.uc.mwt2.org_1456367161</GlobalJobId>
<LocalJobId >158497.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_a6db89bd1ac0</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M1.0S</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:26:01Z</StartTime>
<EndTime >2016-02-25T02:46:02Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/1</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >4456.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >35032.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:46:02Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158497.itbv-pbs.uc.mwt2.org_1456367161"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158497.itbv-pbs.uc.mwt2.org_1456367161</urwg:GlobalJobId>
<urwg:LocalJobId>158497.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_a6db89bd1ac0</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M1S</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:46:02Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:26:01Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/1</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">4456</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">35032</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:46:02Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg"
		xmlns:urwg="http://www.gridforum.org/2003/ur-wg"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
		xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:recordId="itbv-pbs.uc.mwt2.org:158496.itbv-pbs.uc.mwt2.org_1456366666" urwg:createTime="2016-02-25T02:52:34Z" />
<JobIdentity>
<GlobalJobId >itbv-pbs.uc.mwt2.org:158496.itbv-pbs.uc.mwt2.org_1456366666</GlobalJobId>
<LocalJobId >158496.itbv-pbs.uc.mwt2.org</LocalJobId>
</JobIdentity>
<UserIdentity>
	<LocalUserId>glow</LocalUserId>
	<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" >

</ds:KeyInfo>
	<VOName>glow</VOName>
	<ReportableVOName>GLOW</ReportableVOName>
	<CommonName>Generic glow user</CommonName>
</UserIdentity>
<JobName >bl_bce5b13bf521</JobName>
<Status urwg:description="exit status" >1</Status>
<WallDuration >PT20M</WallDuration>
<CpuDuration urwg:usageType="user" >PT0S</CpuDuration>
<CpuDuration urwg:description="Default value" urwg:usageType="system" >PT0S</CpuDuration>
<Processors urwg:metric="total" urwg:consumptionRate="0.0" >1</Processors>
<StartTime >2016-02-25T02:17:46Z</StartTime>
<EndTime >2016-02-25T02:37:46Z</EndTime>
<MachineName urwg:description="Server" >itbv-pbs.uc.mwt2.org</MachineName>
<SiteName >UC_ITB_PBS</SiteName>
<Queue urwg:description="LRMS type: pbs" >short</Queue>
<Host urwg:description="executing host" >itb-c004.mwt2.org/4</Host>
<Memory urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >4468.0</Memory>
<Swap urwg:phaseUnit="PT0S" urwg:storageUnit="KB" urwg:metric="total" >35032.0</Swap>
<TimeInstant urwg:description="LRMS event timestamp" >2016-02-24T20:37:46Z</TimeInstant>
<Resource urwg:description="LocalUserGroup" >osgvo</Resource>
<ProbeName >pbs:itbv-ce-pbs.mwt2.org</ProbeName>
<Grid >OSG-ITB</Grid>
<Resource urwg:description="ResourceType" >Batch</Resource>
</JobUsageRecord>
|<UsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<urwg:RecordIdentity urwg:createTime="2016-02-25T02:52:34Z" urwg:recordId="itbv-pbs.uc.mwt2.org:158496.itbv-pbs.uc.mwt2.org_1456366666"/>
<urwg:JobIdentity>
<urwg:GlobalJobId>itbv-pbs.uc.mwt2.org:158496.itbv-pbs.uc.mwt2.org_1456366666</urwg:GlobalJobId>
<urwg:LocalJobId>158496.itbv-pbs.uc.mwt2.org</urwg:LocalJobId>
</urwg:JobIdentity>
<urwg:UserIdentity>
<urwg:LocalUserId>glow</urwg:LocalUserId>
<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"></ds:KeyInfo>
<urwg:VOName>glow</urwg:VOName><urwg:ReportableVOName>GLOW</urwg:ReportableVOName></urwg:UserIdentity>
<urwg:JobName>bl_bce5b13bf521</urwg:JobName>
<urwg:Status description="exit status">1</urwg:Status>
<urwg:WallDuration>PT20M</urwg:WallDuration>
<urwg:CpuDuration>PT0S</urwg:CpuDuration>
<urwg:EndTime>2016-02-25T02:37:46Z</urwg:EndTime>
<urwg:StartTime>2016-02-25T02:17:46Z</urwg:StartTime>
<urwg:MachineName description="Server">itbv-pbs.uc.mwt2.org</urwg:MachineName>
<urwg:Host description="executing host" primary="false">itb-c004.mwt2.org/4</urwg:Host>
<urwg:Queue description="LRMS type: pbs">short</urwg:Queue>
<urwg:Memory metric="total" storageUnit="KB">4468</urwg:Memory>
<urwg:Swap metric="total" storageUnit="KB">35032</urwg:Swap>
<urwg:Processors consumptionRate="0" metric="total">1</urwg:Processors>
<urwg:TimeInstant description="LRMS event timestamp">2016-02-24T20:37:46Z</urwg:TimeInstant>
<urwg:Resource urwg:description="LocalUserGroup">osgvo</urwg:Resource>
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||`

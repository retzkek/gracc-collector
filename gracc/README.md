# Raw Records

Raw records are in JSON format with a schema derived from the 
[OGF UsageRecord specification](https://www.ogf.org/documents/GFD.98.pdf) 
used by GRACC's predecessor Gratia.

## Requirements

To maintain flexibility and fully leverage the schemaless storage being used,
the schema requirements are kept to a minimum. Some fields are expected by 
GRACC compenents so leaving them out will result in the records not being
properly accounted:

* ResourceType
* CommonName
* VOName
* ReportableVOName
* ProjectName
* EndTime
* CpuDuration
* WallDuration
* Processors
* ...?

### Dates and Times

All times are strings in ISO8601 format. Time durations are floats 
representing seconds.

## Converting XML JobUsageRecord 

The mapping from an XML JobUsageRecord to a JSON Raw GRACC record is
outlined below; this should help inform how new records are generated as well.

The raw XML record is stored in the `RawXML` field, to allow for later reference 
and remapping.

### Identity Groups

Identity groups are flattened by moving their sub-elements to the top level:

* RecordIdentity
  * RecordId
  * CreateTime
* JobIdentity
  * GlobalJobId
  * LocalJobId
  * ProcessId (array)
* UserIdentity
  * LocalUserId
  * GlobalUsername
  * CommonName
  * DN
  * VOName
  * ReportableVOName

### Durations
  
Duration fields are converted to seconds.
CpuDuration can have usage "user" or "system", these are also moved into
the top level:

* CpuDuration (combined)
* CpuDuration_user
* CpuDuration_system
* WallDuration

### Resource

Resources are transformed into a `<description>:<value>` map.
The description is transformed to make it a valid field name 
(spaces and dots are converted to dashes) prefixed with `Resource_`. 
Other properties are flattened as `Resource_<description>_<property_name>:<property_value>`.

The special Resource `ResourceType` is moved directly to the top level. ResourceType 
`BatchPilot` is renamed `Payload` due to the former being misleading.

TimeDuration and TimeInstant elements are likewise put in `<type>:<value>` maps 
with their respective prefixes. Durations are converted to seconds, discrete times 
are ISO8601 strings.

### Other

Any other elements are directly included in the top level. Properties of those elements are moved to fields named as `<element>_<property>`, e.g. `JobName_description`.

### Example XML Record

    <JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
		<RecordIdentity urwg:createTime="2016-05-27T22:46:46Z" urwg:recordId="osg-gw-7.t2.ucsd.edu:35741.2"/>
		<JobIdentity>
		    <GlobalJobId>condor.osg-gw-7.t2.ucsd.edu#185777.0#1464388242</GlobalJobId>
		    <LocalJobId>185777</LocalJobId>
		</JobIdentity>
		<UserIdentity>
		    <LocalUserId>cmsuser</LocalUserId>
		    <GlobalUsername>cmsuser@t2.ucsd.edu</GlobalUsername>
		    <DN>/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba</DN>
		    <VOName>/cms/Role=production/Capability=NULL</VOName>
		    <ReportableVOName>cms</ReportableVOName>
		</UserIdentity>
        <JobName>osg-gw-7.t2.ucsd.edu#185777.0#1464388242</JobName>
        <MachineName>osg-gw-7.t2.ucsd.edu</MachineName>
        <SubmitHost>osg-gw-7.t2.ucsd.edu</SubmitHost>
        <Status urwg:description="Condor Exit Status">0</Status>
        <WallDuration urwg:description="Was entered in seconds">PT10M17.0S</WallDuration>
        <TimeDuration urwg:type="RemoteUserCpu">PT0S</TimeDuration>
        <TimeD
        <TimeDuration urwg:type="RemoteSysCpu">PT18.0S</TimeDuration>
        <TimeDuration urwg:type="LocalSysCpu">PT0S</TimeDuration>
        <TimeDuration urwg:type="CumulativeSuspensionTime">PT0S</TimeDuration>
        <TimeDuration urwg:type="CommittedSuspensionTime">PT0S</TimeDuration>
        <TimeDuration urwg:type="CommittedTime">PT10M17.0S</TimeDuration>
        <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT18.0S</CpuDuration>
        <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT0S</CpuDuration>
        <EndTime urwg:description="Was entered in seconds">2016-05-27T22:44:08Z</EndTime>
        <StartTime urwg:description="Was entered in seconds">2016-05-27T22:33:51Z</StartTime>
        <Host primary="true">cabinet-1-1-1.t2.ucsd.edu</Host>
        <Queue urwg:description="Condor's JobUniverse field">5</Queue>
        <NodeCount urwg:metric="max">1</NodeCount>
        <Processors urwg:metric="max">1</Processors>
        <Resource urwg:description="CondorMyType">Job</Resource>
        <Resource urwg:description="AccountingGroup">group_cmsprod.cmsuser</Resource>
        <Resource urwg:description="ExitBySignal">false</Resource>
        <Resource urwg:description="ExitCode">0</Resource>
        <Resource urwg:description="condor.JobStatus">4</Resource>
        <Network urwg:metric="total" urwg:phaseUnit="PT10M17.0S" urwg:storageUnit="b">0</Network>
        <ProbeName>condor:osg-gw-7.t2.ucsd.edu</ProbeName>
        <SiteName>UCSDT2-D</SiteName>
        <Grid>OSG</Grid>
        <Njobs>1</Njobs>
        <Resource urwg:description="ResourceType">Batch</Resource>
    </JobUsageRecord>

### Example Corresponding JSON

    {
        "RecordId": "osg-gw-7.t2.ucsd.edu:35741.2",
        "CreateTime": "2016-05-27T22:46:46Z",
        "GlobalJobId": "condor.osg-gw-7.t2.ucsd.edu#185777.0#1464388242",
        "LocalJobId": "185777",
        "LocalUserId": "cmsuser",
        "GlobalUsername": "cmsuser@t2.ucsd.edu",
	    "DN": "/DC=ch/DC=cern/OU=Organic Units/OU=Users/CN=sciaba/CN=430796/CN=Andrea Sciaba",
        "VOName": "/cms/Role=production/Capability=NULL",
        "ReportableVOName": "cms",
        "JobName": "osg-gw-7.t2.ucsd.edu#185777.0#1464388242",
        "MachineName": "osg-gw-7.t2.ucsd.edu",
        "SubmitHost": "osg-gw-7.t2.ucsd.edu",
        "Status": "0",
        "Status_description": "Condor Exit Status",
        "WallDuration": 617,
        "WallDuration_description": "Was entered in seconds"
        "TimeDuration_CommittedSuspensionTime": 0,
        "TimeDuration_CommittedTime": 617,
        "TimeDuration_CumulativeSuspensionTime": 0,
        "TimeDuration_LocalSysCpu": 0,
        "TimeDuration_LocalUserCpu": 0,
        "TimeDuration_RemoteSysCpu": 18,
        "TimeDuration_RemoteUserCpu": 0,
        "CpuDuration": 18,
        "CpuDuration_system": 18,
        "CpuDuration_system_description": "Was entered in seconds",
        "CpuDuration_user": 0,
        "CpuDuration_user_description": "Was entered in seconds",
        "EndTime": "2016-05-27T22:44:08Z",
        "StartTime": "2016-05-27T22:33:51Z",
        "Host": "cabinet-1-1-1.t2.ucsd.edu",
        "Queue": "5",
        "Queue_description": "Condor's JobUniverse field",
        "NodeCount": "1",
        "NodeCount_metric": "max",
        "Processors": "1",
        "Processors_metric": "max",
        "Resource_AccountingGroup": "group_cmsprod.cmsuser",
        "Resource_CondorMyType": "Job",
        "Resource_ExitBySignal": "false",
        "Resource_ExitCode": "0",
        "Resource_condor-JobStatus": "4",
        "ResourceType": "Batch",
        "Network": "0",
        "Network_metric": "total",
        "Network_phaseUnit": 617,
        "Network_storageUnit": "b",
        "ProbeName": "condor:osg-gw-7.t2.ucsd.edu",
        "SiteName": "UCSDT2-D",
        "Grid": "OSG",
        "Njobs": "1",
    }


package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
)

var (
	config = &CollectorConfig{
		Address:  "localhost",
		Port:     "8787",
		Timeout:  "10s",
		LogLevel: "DEBUG",
		AMQP: AMQPConfig{
			Host:         "localhost",
			Port:         "5672",
			User:         "guest",
			Password:     "guest",
			Format:       "json",
			Exchange:     "gracc.test.raw",
			ExchangeType: "fanout",
			Durable:      false,
			AutoDelete:   true,
			Internal:     false,
			RoutingKey:   "",
			Retry:        "10s",
		},
		StartBufferSize: 4096,
		MaxBufferSize:   512 * 1024,
	}
	collector *GraccCollector
	consumer  *AMQPOutput
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Verbose() {
		log.SetLevel(log.DebugLevel)
	}
	// initialize collector
	var err error
	if err = config.GetEnv(); err != nil {
		log.Fatalf("error getting env var: %s", err)
	}
	if err = config.Validate(); err != nil {
		log.Fatalf("error in collector config: %s", err)
	}
	if collector, err = NewCollector(config); err != nil {
		log.Fatalf("error starting collector: %s", err)
	}

	// start HTTP server
	http.Handle("/rmi", collector)
	http.HandleFunc("/stats", collector.ServeStats)
	go http.ListenAndServe(config.Address+":"+config.Port, nil)

	// start AMQP consumer
	if err := startConsumer(); err != nil {
		log.Fatalf("error starting consumer: %s", err)
	}

	// run tests
	os.Exit(m.Run())
}

func startConsumer() error {
	var err error
	if consumer, err = InitAMQP(config.AMQP); err != nil {
		return fmt.Errorf("InitAMQP: %s", err)
	}
	cch, err := consumer.OpenChannel()
	if err != nil {
		return fmt.Errorf("OpenChannel: %s", err)
	}
	if err := cch.ExchangeDeclare(config.AMQP.Exchange,
		config.AMQP.ExchangeType,
		config.AMQP.Durable,
		config.AMQP.AutoDelete,
		config.AMQP.Internal,
		false,
		nil); err != nil {
		return fmt.Errorf("ExchangeDeclare: %s", err)
	}
	if _, err := cch.QueueDeclare("gracc.test.queue", false, true, false, false, nil); err != nil {
		return fmt.Errorf("QueueDeclare: %s", err)
	}
	if err := cch.QueueBind("gracc.test.queue", "#", config.AMQP.Exchange, false, nil); err != nil {
		return fmt.Errorf("QueueBind: %s", err)
	}
	inbox, err := cch.Consume("gracc.test.queue", "", false, false, true, false, nil)
	if err != nil {
		return fmt.Errorf("Consume: %s", err)
	}
	// consume records
	go func() {
		for r := range inbox {
			log.Infof("got record %d", r.DeliveryTag)
			log.Debugf("record body:\n---%s\n---\n", r.Body)
			r.Ack(false)
		}
	}()
	return nil
}

func TestPing(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "update")
	v.Set("from", "localhost")
	v.Set("arg1", "xxx")
	v.Set("bundlesize", "1")
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("ping got response %s", resp.Status))
		}
	}
}

func TestUpdate(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "update")
	v.Set("from", "localhost")
	v.Set("arg1", testBundle)
	v.Set("bundlesize", fmt.Sprintf("%d", testBundleSize))
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("update got response %s", resp.Status))
		}
	}
}

func TestMultiUpdate(t *testing.T) {
	testURL := "http://" + config.Address + ":" + config.Port + "/rmi"
	v := url.Values{}
	v.Set("command", "multiupdate")
	v.Set("from", "localhost")
	v.Set("arg1", testBundleXML)
	v.Set("bundlesize", fmt.Sprintf("%d", testBundleXMLSize))
	resp, err := http.PostForm(testURL, v)
	if err != nil {
		t.Error(err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Error(fmt.Errorf("multiupdate got response %s", resp.Status))
		}
	}
}

var testBundleSize = 14
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
<urwg:ProbeName>pbs:itbv-ce-pbs.mwt2.org</urwg:ProbeName><urwg:SiteName>UC_ITB_PBS</urwg:SiteName><urwg:Grid>OSG-ITB</urwg:Grid><urwg:Resource urwg:description="ResourceType">Batch</urwg:Resource></UsageRecord>||replication|<StorageElement xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >UMN-CMS-SE:SE:UMN-CMS-SE</UniqueID>
    <SE >UMN-CMS-SE</SE>
    <Name >UMN-CMS-SE</Name>
    <SpaceType >SE</SpaceType>
    <Timestamp >2012-10-30T00:26:54Z</Timestamp>
    <Implementation >Hadoop</Implementation>
    <Version >Hadoop 0.20.2+737
        Subversion  -r 98c55c28258aa6f42250569bd7fa431ac657bdbd
        Compiled by mockbuild on Tue Nov 29 02:52:15 EST 2011
        From source with checksum 15b415650bf28785987688b54d5e292c
    </Version>
    <Status >Production</Status>
    <ProbeName >hadoop-storage:caffeine.spa.umn.edu</ProbeName>
    <SiteName >UMN-CMS-SE</SiteName>
    <Grid >OSG</Grid>
    <Origin hop="1" ><ServerDate >2012-10-30T00:26:59Z</ServerDate>
        <Connection><SenderHost>128.101.221.170</SenderHost>
            <Sender>hadoop-storage:caffeine.spa.umn.edu</Sender>
            <Collector>collector:gr12x0.fnal.gov/131.225.152.85</Collector>
        </Connection>
    </Origin>
</StorageElement>|||replication|<StorageElement xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >Generic:SE:Generic</UniqueID>
    <SE >Generic</SE>
    <Name >Generic</Name>
    <SpaceType >SE</SpaceType>
    <Timestamp >2013-03-25T18:37:59Z</Timestamp>
    <Implementation >dCache</Implementation>
    <Version >production-1.9.5-26</Version>
    <Status >Production</Status>
    <ProbeName >dCache-storage:dcache-admin.fnal.gov</ProbeName>
    <SiteName >dcache-admin.fnal.gov</SiteName>
    <Grid >OSG</Grid>
</StorageElement>|||replication|<StorageElementRecord xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >UMN-CMS-SE:SE:UMN-CMS-SE</UniqueID>
    <MeasurementType >raw</MeasurementType>
    <StorageType >disk</StorageType>
    <Timestamp >2012-10-30T00:26:54Z</Timestamp>
    <TotalSpace >158961960684544</TotalSpace>
    <UsedSpace >125757757218816</UsedSpace>
    <FreeSpace >33204203465728</FreeSpace>
    <FileCountLimit >2147483647</FileCountLimit>
    <FileCount >37032</FileCount>
    <ProbeName >hadoop-storage:caffeine.spa.umn.edu</ProbeName>
    <Origin hop="1" ><ServerDate >2012-10-30T00:26:59Z</ServerDate>
        <Connection><SenderHost>128.101.221.170</SenderHost>
            <Sender>hadoop-storage:caffeine.spa.umn.edu</Sender>
            <Collector>collector:gr12x0.fnal.gov/131.225.152.85</Collector>
        </Connection>
    </Origin>
</StorageElementRecord>|||replication|<StorageElementRecord xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >Generic:SE:Generic</UniqueID>
    <MeasurementType >raw</MeasurementType>
    <StorageType >disk</StorageType>
    <Timestamp >2013-03-25T18:37:59Z</Timestamp>
    <TotalSpace >16106127360</TotalSpace>
    <UsedSpace >103264</UsedSpace>
    <FreeSpace >16106024096</FreeSpace>
    <ProbeName >dCache-storage:dcache-admin.fnal.gov</ProbeName>
</StorageElementRecord>|||`

var testBundleXMLSize = 10
var testBundleXML = `<?xml version="1.0" encoding="UTF-8"?>
<RecordEnvelope>
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
<CpuDuration urwg:usageType="user" >PT1M5.32S</CpuDuration>
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
<JobUsageRecord xmlns="http://schema.ogf.org/urf/2003/09/urf"
xmlns:urf="http://schema.ogf.org/urf/2003/09/urf"
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
xsi:schemaLocation="http://schema.ogf.org/urf/2003/09/urf
file:/Users/bekah/Documents/GGF/URWG/urwg-schema.09.xsd">
<RecordIdentity urf:recordId="http://www.emsl.pnl.gov/mscf/colony/PBS.1234.0" urf:createTime="2003-08-13T18:56:56Z" />
<JobIdentity>
<LocalJobId>PBS.1234.0</LocalJobId>
</JobIdentity>
<UserIdentity>
<LocalUserId>scottmo</LocalUserId>
</UserIdentity>
<Charge>2870</Charge>
<Status>completed</Status>
<Memory urf:storageUnit="MB">1234</Memory>
<ServiceLevel urf:type="QOS">BottomFeeder</ServiceLevel>
<Processors>4</Processors>
<ProjectName>mscfops</ProjectName>
<MachineName>Colony</MachineName>
<WallDuration>PT1S</WallDuration>
<StartTime>2003-08-13T17:34:50Z</StartTime>
<EndTime>2003-08-13T18:37:38Z</EndTime>
<NodeCount>2</NodeCount>
<Queue>batch</Queue>
<Resource urf:description="quoteId">1435</Resource>
<Resource urf:description="application">NWChem</Resource>
<Resource urf:description="executable">nwchem_linux</Resource>
</JobUsageRecord>
<JobUsageRecord xmlns="http://schema.ogf.org/urf/2003/09/urf"
xmlns:urf="http://schema.ogf.org/urf/2003/09/urf"
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
xsi:schemaLocation="http://schema.ogf.org/urf/2003/09/urf
file:/Users/bekah/Documents/GGF/URWG/urwgschema.09.02.xsd">
<RecordIdentity urf:createTime="2003-08-15T14:25:56Z" urf:recordId="urn:nasa:arc:usage:82125.lomax.nas.nasa.gov:0"/>
<urf:JobIdentity>
<urf:LocalJobId>82125.lomax.nas.nasa.gov</urf:LocalJobId>
</urf:JobIdentity>
<urf:UserIdentity>
<urf:LocalUserId>foobar</urf:LocalUserId>
</urf:UserIdentity>
<Status urf:description="pbs exit status">0</Status>
<urf:Memory urf:metric="max" urf:storageUnit="KB" urf:type="virtual">1060991</urf:Memory>
<urf:Processors urf:metric="total">32</urf:Processors>
<urf:EndTime>2003-06-16T08:24:32Z</urf:EndTime>
<urf:ProjectName urf:description="local charge group">g13563</urf:ProjectName>
<urf:Host urf:primary="true">lomax.nas.nasa.gov</urf:Host>
<urf:Queue>lomax</urf:Queue>
<urf:WallDuration>PT45M48S</urf:WallDuration>
<urf:CpuDuration>PT15S</urf:CpuDuration>
<urf:Resource urf:description="pbs-jobname">m0.20a-7.0b0.0v</urf:Resource>
</JobUsageRecord>
<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:createTime="2016-05-02T16:36:19Z" urwg:recordId="ce01.brazos.tamu.edu:3976909.14"/>
<JobIdentity>
    <GlobalJobId>slurm:SLURM/brazos/brazos.13634504_4</GlobalJobId>
    <LocalJobId>13634504_4</LocalJobId>
</JobIdentity>
<UserIdentity>
    <LocalUserId>georgemm01</LocalUserId>
    <VOName>cms</VOName>
    <ReportableVOName>CMS</ReportableVOName>
</UserIdentity>
    <JobName>CDMSBulldozer</JobName>
    <Status>0</Status>
    <Processors urwg:metric="total">1</Processors>
    <WallDuration>PT21.0S</WallDuration>
    <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT0.55S</CpuDuration>
    <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT0.26S</CpuDuration>
    <EndTime urwg:description="Was entered in seconds">2016-05-02T16:00:46Z</EndTime>
    <StartTime urwg:description="Was entered in seconds">2016-05-02T16:00:25Z</StartTime>
    <Queue>stakeholder</Queue>
    <ProjectName>hepx</ProjectName>
    <Processors urwg:metric="total">1</Processors>
    <ProbeName>slurm:ce01.brazos.tamu.edu</ProbeName>
    <SiteName>TAMU_BRAZOS_CE</SiteName>
    <Grid>Local</Grid>
    <Njobs>1</Njobs>
    <Resource urwg:description="ResourceType">Batch</Resource>
</JobUsageRecord>
<JobUsageRecord xmlns="http://www.gridforum.org/2003/ur-wg" xmlns:urwg="http://www.gridforum.org/2003/ur-wg" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.gridforum.org/2003/ur-wg file:///u:/OSG/urwg-schema.11.xsd">
<RecordIdentity urwg:createTime="2016-05-27T22:46:49Z" urwg:recordId="submit-3.chtc.wisc.edu:1140965.8929"/>
<JobIdentity>
    <GlobalJobId>condor.submit-3.chtc.wisc.edu#60242911.0#1464388571</GlobalJobId>
    <LocalJobId>60242911</LocalJobId>
</JobIdentity>
<UserIdentity>
    <LocalUserId>kavandermeul</LocalUserId>
    <GlobalUsername>kavandermeul@chtc.wisc.edu</GlobalUsername>
<VOName>GLOW</VOName><ReportableVOName>GLOW</ReportableVOName></UserIdentity>
    <JobName>submit-3.chtc.wisc.edu#60242911.0#1464388571</JobName>
    <MachineName>submit-3.chtc.wisc.edu</MachineName>
    <SubmitHost>submit-3.chtc.wisc.edu</SubmitHost>
    <Status urwg:description="Condor Exit Status">0</Status>
    <WallDuration urwg:description="Was entered in seconds">PT5.0S</WallDuration>
    <TimeDuration urwg:type="RemoteUserCpu">PT0S</TimeDuration>
    <TimeDuration urwg:type="LocalUserCpu">PT0S</TimeDuration>
    <TimeDuration urwg:type="RemoteSysCpu">PT0S</TimeDuration>
    <TimeDuration urwg:type="LocalSysCpu">PT0S</TimeDuration>
    <TimeDuration urwg:type="CumulativeSuspensionTime">PT0S</TimeDuration>
    <TimeDuration urwg:type="CommittedSuspensionTime">PT0S</TimeDuration>
    <TimeDuration urwg:type="CommittedTime">PT5.0S</TimeDuration>
    <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="system">PT0S</CpuDuration>
    <CpuDuration urwg:description="Was entered in seconds" urwg:usageType="user">PT0S</CpuDuration>
    <EndTime urwg:description="Was entered in seconds">2016-05-27T22:36:42Z</EndTime>
    <StartTime urwg:description="Was entered in seconds">2016-05-27T22:36:37Z</StartTime>
    <Host primary="true" urwg:description="wisc.edu">c220g2-030631.wisc.cloudlab.us</Host>
    <Queue urwg:description="Condor's JobUniverse field">5</Queue>
    <NodeCount urwg:metric="max">1</NodeCount>
    <Processors urwg:metric="max">1</Processors>
    <Resource urwg:description="CondorMyType">Job</Resource>
    <Resource urwg:description="ExitBySignal">false</Resource>
    <Resource urwg:description="ExitCode">1</Resource>
    <Resource urwg:description="condor.JobStatus">4</Resource>
    <Network urwg:metric="total" urwg:phaseUnit="PT5.0S" urwg:storageUnit="b">0</Network>
    <ProbeName>condor:submit-3.chtc.wisc.edu</ProbeName>
    <SiteName>CHTC</SiteName>
    <Grid>OSG</Grid>
    <Njobs>1</Njobs>
    <Resource urwg:description="ResourceType">BatchPilot</Resource>
</JobUsageRecord>
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
    <TimeDuration urwg:type="LocalUserCpu">PT0S</TimeDuration>
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
<StorageElement xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >UMN-CMS-SE:SE:UMN-CMS-SE</UniqueID>
    <SE >UMN-CMS-SE</SE>
    <Name >UMN-CMS-SE</Name>
    <SpaceType >SE</SpaceType>
    <Timestamp >2012-10-30T00:26:54Z</Timestamp>
    <Implementation >Hadoop</Implementation>
    <Version >Hadoop 0.20.2+737
        Subversion  -r 98c55c28258aa6f42250569bd7fa431ac657bdbd
        Compiled by mockbuild on Tue Nov 29 02:52:15 EST 2011
        From source with checksum 15b415650bf28785987688b54d5e292c
    </Version>
    <Status >Production</Status>
    <ProbeName >hadoop-storage:caffeine.spa.umn.edu</ProbeName>
    <SiteName >UMN-CMS-SE</SiteName>
    <Grid >OSG</Grid>
    <Origin hop="1" ><ServerDate >2012-10-30T00:26:59Z</ServerDate>
        <Connection><SenderHost>128.101.221.170</SenderHost>
            <Sender>hadoop-storage:caffeine.spa.umn.edu</Sender>
            <Collector>collector:gr12x0.fnal.gov/131.225.152.85</Collector>
        </Connection>
    </Origin>
</StorageElement>
<StorageElement xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >Generic:SE:Generic</UniqueID>
    <SE >Generic</SE>
    <Name >Generic</Name>
    <SpaceType >SE</SpaceType>
    <Timestamp >2013-03-25T18:37:59Z</Timestamp>
    <Implementation >dCache</Implementation>
    <Version >production-1.9.5-26</Version>
    <Status >Production</Status>
    <ProbeName >dCache-storage:dcache-admin.fnal.gov</ProbeName>
    <SiteName >dcache-admin.fnal.gov</SiteName>
    <Grid >OSG</Grid>
</StorageElement>

<StorageElementRecord xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >UMN-CMS-SE:SE:UMN-CMS-SE</UniqueID>
    <MeasurementType >raw</MeasurementType>
    <StorageType >disk</StorageType>
    <Timestamp >2012-10-30T00:26:54Z</Timestamp>
    <TotalSpace >158961960684544</TotalSpace>
    <UsedSpace >125757757218816</UsedSpace>
    <FreeSpace >33204203465728</FreeSpace>
    <FileCountLimit >2147483647</FileCountLimit>
    <FileCount >37032</FileCount>
    <ProbeName >hadoop-storage:caffeine.spa.umn.edu</ProbeName>
    <Origin hop="1" ><ServerDate >2012-10-30T00:26:59Z</ServerDate>
        <Connection><SenderHost>128.101.221.170</SenderHost>
            <Sender>hadoop-storage:caffeine.spa.umn.edu</Sender>
            <Collector>collector:gr12x0.fnal.gov/131.225.152.85</Collector>
        </Connection>
    </Origin>
</StorageElementRecord>
<StorageElementRecord xmlns:urwg="http://www.gridforum.org/2003/ur-wg">
    <UniqueID >Generic:SE:Generic</UniqueID>
    <MeasurementType >raw</MeasurementType>
    <StorageType >disk</StorageType>
    <Timestamp >2013-03-25T18:37:59Z</Timestamp>
    <TotalSpace >16106127360</TotalSpace>
    <UsedSpace >103264</UsedSpace>
    <FreeSpace >16106024096</FreeSpace>
    <ProbeName >dCache-storage:dcache-admin.fnal.gov</ProbeName>
</StorageElementRecord>
</RecordEnvelope>`

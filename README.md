# gratia2 collector

This is an all-new implementation of the [Open Science Grid](http://www.opensciencegrid.org) 
[Gratia](https://sourceforge.net/projects/gratia/) accounting collector. It is intended
to collect records forwarded from existing collectors and put them into other formats/datastores,
including:

* File (XML or JSON format)
* Elasticsearch index
* Kafka topic

This is meant for testing, development, or backup purposes.

## Configuration

The config file is [TOML](https://github.com/toml-lang/toml) format. 

    address = ""         # address to listen on
    port = "8888"        # port to listen on
    loglevel = "debug"   # log level (debug, info, warn, error, fatal, panic)
    
    [file]
    enabled = true       # output records to file, one record per file
    path = '/tmp/gratia/{{.RecordIdentity.CreateTime.Format "2006/01/02/15/04"}}/'
                         # path is the directory to create files in, supports dynamic naming based on 
                         # templated attributes of the Record. 
                         # Reference time: Mon Jan 2 15:04:05 -0700 MST 2006 
    format = "xml"       # format (xml or json). filename is <recordID hash>.<format>
    
    [elasticsearch]
    enabled = true                   # output records to Elasticsearch
    host = "http://localhost:9200"   # Elasticsearch URL
    index = "gratia-test"            # index
    
    [kafka]
    enabled = true                   # output records to Kafka
    brokers = ["localhost:9092"]     # list of brokers
    topic = "gratia-osg"             # topic

## Usage

    gratia2-collector -c <config file> -l <log file>

Where `<log file>` can be "stdout", "stderr", or a file name.

See `sample/gratia.service` for a sample systemd unit configuration. Copy the file (with 
appropriate changes) to `/usr/lib/systemd/system/` then use standard systemd commands to
control the process.

* Start: `systemctl start gratia2.service`
* Stop:  `systemctl stop gratia2.service`
* Restart:  `systemctl restart gratia2.service`
* Refresh log file:  `systemctl kill --signal=SIGUSR1 gratia2.service`
* Toggle debug logging:  `systemctl kill --signal=SIGUSR2 gratia2.service`

See `sample/gratia2.logrotate` for a sample logrotate configuration. Copy the file (with
appropriate changes) to `/etc/logrotate.d/gratia2`.

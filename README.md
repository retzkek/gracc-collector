# GRÅCC - Gratia-Compatible Collector

This is an all-new implementation of the [Open Science Grid](http://www.opensciencegrid.org) 
[Gratia](https://sourceforge.net/projects/gratia/) accounting collector. It is intended
to collect records forwarded from existing collectors and put them into other formats/datastores,
including:

* File (XML or JSON format)
* Elasticsearch index
* Kafka topic
* RabbitMQ (or other AMQP 0.9.1-compatible broker)

This is meant for testing, development, or backup purposes.

## Configuration

The config file is [TOML](https://github.com/toml-lang/toml) format. 

    address = ""         # address to listen on
    port = "8888"        # port to listen on
    loglevel = "debug"   # log level (debug, info, warn, error, fatal, panic)
    
    [file]
    enabled = true       # output records to file, one record per file
    path = '/tmp/gracc/{{.RecordIdentity.CreateTime.Format "2006/01/02/15/04"}}/'
                         # path is the directory to create files in, supports dynamic naming based on 
                         # templated attributes of the Record. 
                         # Reference time: Mon Jan 2 15:04:05 -0700 MST 2006 
    format = "xml"       # format (xml or json). filename is <recordID hash>.<format>
    
    [elasticsearch]
    enabled = true                   # output records to Elasticsearch
    host = "http://localhost:9200"   # Elasticsearch URL
    index = "gracc-test"             # index
    
    [kafka]
    enabled = true                   # output records to Kafka
    brokers = ["localhost:9092"]     # list of brokers
    topic = "gracc-osg"              # topic
    
    [AMQP]
    enabled = true                   # output records to RabbitMQ
    host = "localhost"
    port = "5672"
    user = "guest"
    password = "guest"
    exchange = ""
    routingKey = ""
    format = "xml"     

## Usage

    gracc -c <config file> -l <log file>

Where `<log file>` can be "stdout", "stderr", or a file name.

See `sample/gracc.service` for a sample systemd unit configuration. Copy the file (with 
appropriate changes) to `/usr/lib/systemd/system/` then use standard systemd commands to
control the process.

* Start: `systemctl start gracc.service`
* Stop:  `systemctl stop gracc.service`
* Restart:  `systemctl restart gracc.service`
* Refresh log file:  `systemctl kill --signal=SIGUSR1 gracc.service`
* Toggle debug logging:  `systemctl kill --signal=SIGUSR2 gracc.service`

See `sample/gracc.logrotate` for a sample logrotate configuration. Copy the file (with
appropriate changes) to `/etc/logrotate.d/gracc`.

## Versions

### v0.01.07

Created internal queues for outputting records.

### v0.01.05

Add basic collector statistics served at /stats.

### v0.01.04

Add option to write records to RabbitMQ.

### v0.01.03

Add option to write records to file.

### v0.01.02

Add option to send records to Kafka topic.

### v0.01.01

Initial dev version, send records to Elasticsearch.

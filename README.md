[![Build Status](https://travis-ci.org/opensciencegrid/gracc-collector.svg?branch=master)](https://travis-ci.org/opensciencegrid/gracc-collector)

# Overview

The gracc-collector is a "Gratia-Compatible Collector" that acts as a 
transitional interface between the obsolete [Gratia](https://sourceforge.net/projects/gratia/)
accounting collector and probes and the new GRÅCC accounting system.

It listens for bundles of records (as would be sent via replication from a 
Gratia collector or from a Gratia probe) on HTTP, processes the bundle into 
individual usage records, and sends those to RabbitMQ or another 
AMQP 0.9.1 broker.

# Configuration

The config file is [TOML](https://github.com/toml-lang/toml) format.
Config options can also be specified by environment variables, shown
below in parentheses. Environment variables override file settings.

    address = "localhost" # address to listen on (GRACC_ADDRESS)
    port = "8888"         # port to listen on (GRACC_PORT)
    timeout = "60s"       # HTTP connection timeout (GRACC_TIMEOUT)
    loglevel = "debug"    # log level [debug|info|warn|error|fatal|panic] (GRACC_LOGLEVEL)
    
    [AMQP]
    host = "localhost"    # AMQP broker (GRACC_AMQP_HOST)
    port = "5672"         # (GRACC_AMQP_PORT)
    vhost = ""            # (GRACC_AMQP_VHOST)
    exchange = ""         # (GRACC_AMQP_EXCHANGE)
    routingKey = ""       # (GRACC_AMQP_ROUTINGKEY)
    durable = true        # keep exchange between server restarts (GRACC_AMQP_DURABLE)
    autoDelete = true     # delete exchange when there are no remaining bindings (GRACC_AMQP_AUTODELETE)
    user = "guest"        # (GRACC_AMQP_USER)
    password = "guest"    # (GRACC_AMQP_PASSWORD)
    format = "raw"        # format to send record in [raw|xml|json] (GRACC_AMQP_FORMAT)
    retry = "10s"         # AMQP connection retry interval (GRACC_AMQP_RETRY)

# Usage

    gracc-collector -c <config file> -l <log file>

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

# Release Notes

### v0.4.0

* Allow config options to be set by environment variable. 
  Precedence: env var > config file > default.
* Change timeout and AMQP.retry config options to duration strings.
  A duration string is a possibly signed sequence of decimal numbers, each with
  optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m". 
  Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". 

### v0.03.01

Fully flatten JobUsageRecord into simple key:value hash map.

### v0.03.00

Revised JSON output structure based on OGF UsageRecord and switch default output to JSON.

[Mapping documentation](https://opensciencegrid.github.io/gracc/dev-docs/raw-records/)

### v0.02.00

Significant rewrite to simplify while maintaining robust error handling.

* Removed unneeded file, elasticsearch, and kafka outputs.
* One AMQP channel is opened for each bundle received; the HTTP connection
  from the probe or collector is not closed until all records are confirmed
  by the AMQP broker. If an error occurs or a record is returned then a 500
  error code is returned.

### v0.01.10

* Handle AMQP returned records
* Improved parsing of record bundles

### v0.01.09

* Better handling for AMQP communication errors

### v0.01.08

* Support multiple AMQP workers
* Timeout if output filter is blocking too long, respond to request with error code so bundle gets re-sent later
* Improved log messages and responses

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

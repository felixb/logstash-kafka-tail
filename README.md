logstash-kafka-tail
===================

logstash-kafka-tail is a tail/grep like tool for tracking logs on command line when using logstash with kafka.
It consumes a kafka topic and parses the logstash/json formated log messages to print them on stdandard out.

Installation
------------

Install the binary with `go get` or download prebuid binarys from the release section.

    go get github.com/felixb/logstash-kafka-tail

Usage
-----

logstash-kafka-tail has the following command line options:

  -F, --filter=[]                                                       filter messages, specify like 'type:chaos-monkey'
                                                                          all filters must match when applied multiple times
  -f, --format=%{@timestamp} %{type} %{HOSTNAME} %{level} %{message}    format output in grok syntax
                                                                          also read from env 'KAFKA_LOGGING_FORMAT'
  -h, --hosts=[]                                                        kafka hosts
                                                                          also read from env 'KAFKA_LOGGING_HOSTS'
  -o, --offset=-1                                                       offset to start reading, -1 => newest, -2 => oldest
  -t, --topic=logstash                                                  kafka topic
                                                                          also read from env 'KAFKA_LOGGING_TOPIC'
  --version=false                                                       prints the version

It's possible to set `-hosts`, `-topic` and `-format` as evironment variable like this:

    export KAFKA_LOGGING_HOSTS="kafka-01.example.com:9092,kafka-02.example.com:9092"
    export KAFKA_LOGGING_TOPIC="customloggingtopic"
    export KAFKA_LOGGING_FORMAT="%{@timestamp} >> %{type} %{host} %{log_level} ### %{custom_field} %{message}"

The `-filter` option shows only those messages, which match every single key/value pair specified in the option.

Building
--------

Build logstash-kafka-tail by running

    make get build

You can run `go get . && go build` though.

Testing
-------

Run `make test` or `go test` to run the tests.

Contributing
------------

Please fork the project and send a pull request.

License
-------

This program is licensed under the MIT license. See LICENSE for details.

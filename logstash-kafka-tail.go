package main

import (
	"fmt"
	"os"
	"strings"

	sarama "github.com/Shopify/sarama"
	flag "github.com/docker/docker/pkg/mflag"
)

type args []string

const (
	version       = "0.2.0"
	defaultHosts  = "localhost:9092"
	usageHosts    = "kafka hosts\nalso read from env 'KAFKA_LOGGING_HOSTS'"
	defaultTopic  = "logstash"
	usageTopic    = "kafka topic\nalso read from env 'KAFKA_LOGGING_TOPIC'"
	defaultOffset = sarama.OffsetNewest
	usageOffset   = "offset to start reading, -1 => newest, -2 => oldest"
	defaultFormat = "%{@timestamp} %{type} %{HOSTNAME,hostname} %{level,loglevel,log_level,severity} %{message}"
	usageFormat   = "format output in grok syntax\nalso read from env 'KAFKA_LOGGING_FORMAT'"
	usageFilter   = "filter messages, specify like 'type:chaos-monkey'\nall filters must match when applied multiple times"
	usageVersion  = "prints the version"
)

func (a *args) String() string {
	return fmt.Sprint(*a)
}

func (a *args) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// parse args and start consumer
func main() {
	defHosts := os.Getenv("KAFKA_LOGGING_HOSTS")
	if defHosts == "" {
		defHosts = defaultHosts
	}

	defTopic := os.Getenv("KAFKA_LOGGING_TOPIC")
	if defTopic == "" {
		defTopic = defaultTopic
	}

	defFormat := os.Getenv("KAFKA_LOGGING_FORMAT")
	if defFormat == "" {
		defFormat = defaultFormat
	}

	var hosts args
	var topic string
	var offset int64
	var filterList args
	var showVersion bool
	var formatString string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Var(&hosts, []string{"h", "-hosts"}, usageHosts)
	flag.StringVar(&topic, []string{"t", "-topic"}, defTopic, usageTopic)
	flag.Int64Var(&offset, []string{"o", "-offset"}, defaultOffset, usageOffset)
	flag.StringVar(&formatString, []string{"f", "-format"}, defFormat, usageFormat)
	flag.Var(&filterList, []string{"F", "-filter"}, usageFilter)
	flag.BoolVar(&showVersion, []string{"-version"}, false, usageVersion)
	flag.Parse()

	if showVersion {
		fmt.Printf("logstash-kafka-tail v%s\n", version)
		os.Exit(0)
	}

	if len(hosts) == 0 {
		hosts = strings.Split(defHosts, ",")
	}

	if topic == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	formatter := NewFormatter(formatString)

	var printer Printer
	if len(filterList) > 0 {
		filters := map[string]string{}
		for _, fltr := range filterList {
			parts := strings.SplitN(fltr, ":", 2)
			if len(parts) == 2 {
				filters[parts[0]] = parts[1]
			}
		}
		filter := NewFilter(filters, &formatter)
		printer = &filter
	} else {
		printer = &formatter
	}

	async := NewAsyncPrinter(printer)
	async.Start()

	consumer := NewConsumer(hosts, topic, offset, &async)
	consumer.Start()
	async.Wait()
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"

	sarama "github.com/Shopify/sarama"
	flag "github.com/docker/docker/pkg/mflag"
)

type args []string

const (
	version       = "0.1.0"
	defaultHosts  = "localhost:9092"
	usageHosts    = "kafka hosts\nalso read from env 'KAFKA_LOGGING_HOSTS'"
	defaultTopic  = "logstash"
	usageTopic    = "kafka topic\nalso read from env 'KAFKA_LOGGING_TOPIC'"
	defaultOffset = sarama.OffsetNewest
	usageOffset   = "offset to start reading, -1 => newest, -2 => oldest"
	defaultFormat = "%{@timestamp} %{type} %{HOSTNAME} %{level} %{message}"
	usageFormat   = "format output in grok syntax\nalso read from env 'KAFKA_LOGGING_FORMAT'"
	usageFilter   = "filter messages, specify like 'type:chaos-monkey'\nall filters must match when applied multiple times"
	usageVersion  = "prints the version"
)

var (
	hosts     args
	topic     string
	offset    int64
	formatter Formatter
	filters   = map[string]string{}
)

// unmarshal the message
func unmarshal(msg *sarama.ConsumerMessage) (*Message, error) {
	var m *Message
	err := json.Unmarshal(msg.Value, m)
	return m, err
}

// unmarshals, filters and prints a formated message
func handleMessage(msg *sarama.ConsumerMessage, ch chan *Message) {
	m, err := unmarshal(msg)
	if err != nil {
		log.Printf("error (%s) parsing message: %s", err, msg.Value)
	} else {
		ch <- m
	}
}

// consumes a single partition
func consumePartition(wg *sync.WaitGroup, master sarama.Consumer, partition int32, ch chan *Message) {
	defer wg.Done()
	log.Printf("Starting consumer for partition %d", partition)

	consumer, err := master.ConsumePartition(topic, partition, offset)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	var msgCount int64

	for {
		select {
		case err := <-consumer.Errors():
			log.Println(err)
		case msg := <-consumer.Messages():
			handleMessage(msg, ch)
			msgCount++
		case <-signals:
			log.Printf("Stopping consumer for partition %d, processed %d messages", partition, msgCount)
			return
		}
	}
}

// reads and filters Messages from chan and prints them to stdout
func printMessages(wg *sync.WaitGroup, ch chan *Message) {
	defer wg.Done()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		select {
		case m := <-ch:
			if m.Filter() {
				formatter.Print(m)
			}
		case <-signals:
			return
		}
	}
}

// connects to one of a list of brokers
func connectToBroker(config *sarama.Config) (*sarama.Broker, error) {
	var err error
	for _, host := range hosts {
		broker := sarama.NewBroker(host)
		err = broker.Open(config)
		if err != nil {
			log.Printf("error connecting to broker: %s %v", host, err)
		} else {
			return broker, nil
		}
	}
	return nil, err
}

// connects to the broker and fetches current brokers' address and partition ids
func fetchMetadata(config *sarama.Config) ([]string, []int32, error) {
	broker, err := connectToBroker(config)
	if err != nil {
		return nil, nil, err
	}
	request := sarama.MetadataRequest{Topics: []string{topic}}
	response, err := broker.GetMetadata(&request)
	if err != nil {
		_ = broker.Close()
		return nil, nil, err
	}

	if len(response.Brokers) == 0 {
		return nil, nil, errors.New(fmt.Sprintf("Unable to find any broker for topic: %s", topic))
	}
	if len(response.Topics) != 1 {
		return nil, nil, errors.New(fmt.Sprintf("Invalid number of topics: %d", len(response.Topics)))
	}

	var brokers []string
	for _, broker := range response.Brokers {
		// log.Printf("broker: %q", broker.Addr())
		brokers = append(brokers, broker.Addr())
	}

	var partitions []int32
	for _, partition := range response.Topics[0].Partitions {
		// log.Printf("partition: %v, leader: %v", partition.ID, partition.Leader)
		partitions = append(partitions, partition.ID)
	}

	return brokers, partitions, broker.Close()
}

// consume messages from kafka and print them to stdout
func consume() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.ClientID = "logstash-kafka-tail"

	brokers, partitions, err := fetchMetadata(config)
	if err != nil {
		log.Fatalf("error fetching metadata from broker: %v", err)
	}

	master, err := sarama.NewConsumer([]string(brokers), config)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := master.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	var wg sync.WaitGroup
	ch := make(chan *Message, 10)

	// spawn the printer
	wg.Add(1)
	go printMessages(&wg, ch)

	// spawn the consumers
	for _, partition := range partitions {
		wg.Add(1)
		go consumePartition(&wg, master, partition, ch)
	}

	// wait for everyone to finish
	wg.Wait()
}

func (a *args) String() string {
	return fmt.Sprint(*a)
}

func (a *args) Set(value string) error {
	for _, arg := range strings.Split(value, ",") {
		*a = append(*a, arg)
	}
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

	formatter = NewFormatter(formatString)

	for _, fltr := range filterList {
		parts := strings.SplitN(fltr, ":", 2)
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}

	consume()
}

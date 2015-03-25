package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	sarama "github.com/Shopify/sarama"
)

type Consumer struct {
	hosts   []string
	topic   string
	offset  int64
	printer Printer
	wg      sync.WaitGroup
}

// create a new consumer
func NewConsumer(h args, t string, o int64, p Printer) Consumer {
	return Consumer{h, t, o, p, sync.WaitGroup{}}
}

// start the consumer and listen to topic
func (c *Consumer) Start() {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.ClientID = "logstash-kafka-tail"

	brokers, partitions, err := c.fetchMetadata(config)
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

	// spawn the consumers
	for _, partition := range partitions {
		c.wg.Add(1)
		go c.consumePartition(master, partition)
	}

	// wait for everyone to finish
	c.wg.Wait()
}

// unmarshal the message
func (c *Consumer) unmarshal(msg *sarama.ConsumerMessage) (*Message, error) {
	var m Message
	err := json.Unmarshal(msg.Value, &m)
	return &m, err
}

// unmarshals, filters and prints a formated message
func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	m, err := c.unmarshal(msg)
	if err != nil {
		log.Printf("error (%s) parsing message: %s", err, msg.Value)
	} else {
		c.printer.Print(m)
	}
}

// consumes a single partition
func (c *Consumer) consumePartition(master sarama.Consumer, partition int32) {
	defer c.wg.Done()
	log.Printf("Starting consumer for partition %d", partition)

	consumer, err := master.ConsumePartition(c.topic, partition, c.offset)
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
			c.handleMessage(msg)
			msgCount++
		case <-signals:
			log.Printf("Stopping consumer for partition %d, processed %d messages", partition, msgCount)
			return
		}
	}
}

// connects to one of a list of brokers
func (c *Consumer) connectToBroker(config *sarama.Config) (*sarama.Broker, error) {
	var err error
	for _, host := range c.hosts {
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
func (c *Consumer) fetchMetadata(config *sarama.Config) ([]string, []int32, error) {
	broker, err := c.connectToBroker(config)
	if err != nil {
		return nil, nil, err
	}
	request := sarama.MetadataRequest{Topics: []string{c.topic}}
	response, err := broker.GetMetadata(&request)
	if err != nil {
		_ = broker.Close()
		return nil, nil, err
	}

	if len(response.Brokers) == 0 {
		return nil, nil, errors.New(fmt.Sprintf("Unable to find any broker for topic: %s", c.topic))
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

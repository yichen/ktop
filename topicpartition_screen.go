package ktop

import (
	"fmt"
	"log"
	"sort"

	"github.com/Shopify/sarama"
	"github.com/nsf/termbox-go"
)

// TopicPartitionInfo provides detail informations for a topic
type TopicPartitionInfo struct {
	Earlest int64
	Latest  int64
}

type PartitionMetadata []*sarama.PartitionMetadata

func (pm PartitionMetadata) Len() int {
	return len(pm)
}

func (pm PartitionMetadata) Swap(i, j int) {
	pm[i], pm[j] = pm[j], pm[i]
}

func (pm PartitionMetadata) Less(i, j int) bool {
	return pm[i].ID < pm[j].ID
}

type TopicPartitionScreen struct {
	topic      string
	client     sarama.Client
	cluster    *Cluster
	broker     string
	brokers    []*sarama.Broker
	topics     []*sarama.TopicMetadata
	partitions PartitionMetadata
}

func NewTopicPartitionScreen(cluster *Cluster, client sarama.Client, topic string, broker string) *TopicPartitionScreen {
	return &TopicPartitionScreen{
		client:  client,
		topic:   topic,
		broker:  broker,
		cluster: cluster,
	}
}

func (s *TopicPartitionScreen) WillShow(screen Screen) {
	// get TopicPartition metadata
	broker := sarama.NewBroker(s.broker)

	config := sarama.NewConfig()

	err := broker.Open(config)
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	// first, obtain meta data about the topic
	var topics [1]string
	topics[0] = s.topic

	topicMetaReq := &sarama.MetadataRequest{
		Topics: topics[:],
	}

	metadata, err := broker.GetMetadata(topicMetaReq)
	if err != nil {
		log.Println(err)
		// panic(err)
	}

	s.brokers = metadata.Brokers
	s.topics = metadata.Topics
	s.partitions = s.topics[0].Partitions
	sort.Sort(s.partitions)
}

func (s *TopicPartitionScreen) Refresh(screen Screen) {

	// if topic metadata does not exist, do nothing
	if len(s.topics) == 0 {
		log.Println("ERROR, topic metadata shouldn't be empty")
		return
	}

	topicMetadata := s.topics[0]
	partitionMetadata := topicMetadata.Partitions

	header := fmt.Sprintf("%4s%10s%20s%20s", "ID", "Leader", "Replicas", "ISR")
	screen.Print(header, 0, 0, coldef, coldef)

	for r, p := range partitionMetadata {
		replicas := ""
		for _, rep := range p.Replicas {
			replicas += fmt.Sprintf("%v ", rep)
		}

		isrs := ""
		for _, isr := range p.Isr {
			isrs += fmt.Sprintf("%v ", isr)
		}

		text := fmt.Sprintf("%4v%10v%20s%20s", p.ID, p.Leader, replicas, isrs)
		screen.Print(text, 0, r+1, coldef, coldef)
	}
}

func (ts *TopicPartitionScreen) OnKeyInput(screen Screen, keyEvent termbox.Event) {

	switch keyEvent.Type {
	case termbox.EventKey:
		switch keyEvent.Key {
		case termbox.KeyEnter:
		case termbox.KeyArrowDown:
		case termbox.KeyArrowUp:
		case termbox.KeyArrowLeft, termbox.KeyCtrlQ:
			// go up
			screen.Pop()

		case termbox.KeyCtrlF, termbox.KeyPgdn:
		case termbox.KeyCtrlB, termbox.KeyPgup:
		case termbox.KeyBackspace, termbox.KeyBackspace2:

		default:
		}
	case termbox.EventError:
		panic(keyEvent.Err)
	}
}

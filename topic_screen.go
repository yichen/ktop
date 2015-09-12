package ktop

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/nsf/termbox-go"
	"github.com/yichen/suggest"
)

var LogFile *os.File

func init() {
	LogFile, _ := os.OpenFile("/tmp/ktop.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.SetOutput(LogFile)
}

const coldef = termbox.ColorDefault

var quit = false
var w, h int

type TopicList []string

func (tl TopicList) Len() int {
	return len(tl)
}

func (tl TopicList) Swap(i, j int) {
	tl[i], tl[j] = tl[j], tl[i]
}

func (tl TopicList) Less(i, j int) bool {
	return strings.ToLower(tl[i]) < strings.ToLower(tl[j])
}

type TopicInfo struct {
	Name          string
	NumPartitions int
	Partitions    []int32
}

type TopicScreen struct {

	// position in the FilteredTopic list, that is currently the begining of the screen
	// This is for marking the navigation. PageUp and PageDown will increase or decrease
	// this value by the size of the page. When Position changes, the Cursor follows it.
	Position int

	// the position in the FilteredTopic list, that is currently pointed by the cursor
	// ArrowUp and ArrowDown can change the Cursor if it is still visible in the current page
	Cursor int

	// all known topics
	Topics TopicList

	// filtered topics. If Query is not empty, we will filter the
	// topic list and only show a subset that matches the filter
	FilteredTopics TopicList

	// map to hold the topic information
	TopicInfos map[string]TopicInfo

	// query string
	Query string
	// filtered

	typeahead *suggest.Suggest
	client    sarama.Client
	cluster   *Cluster

	broker string
}

func NewTopicScreen(cluster *Cluster, client sarama.Client, broker string) *TopicScreen {
	return &TopicScreen{
		client:     client,
		TopicInfos: make(map[string]TopicInfo),
		typeahead:  suggest.NewSuggest(),
		broker:     broker,
		cluster:    cluster,
	}
}

func (s *TopicScreen) refreshTopicIndex() {
	for _, t := range s.Topics {
		if !s.typeahead.ContainsDocument(t) {
			s.typeahead.AddSymbol(t)
		}
	}
}

func (s *TopicScreen) filter() {
	if len(s.Query) == 0 {
		s.FilteredTopics = s.Topics
	} else {
		s.FilteredTopics = s.typeahead.SearchAll(s.Query)
	}

	log.Println("[filter] query:" + s.Query + ", query len: " + strconv.Itoa(len(s.Query)) + ", number of filtered topics: " + strconv.Itoa(len(s.FilteredTopics)))

	s.refreshTopicInformations()
}

func (s *TopicScreen) refreshTopicInformations() {
	for _, topic := range s.FilteredTopics {
		partitions, err := s.client.Partitions(topic)
		if err != nil {
			panic(err)
		}

		if info, ok := s.TopicInfos[topic]; !ok {
			info = TopicInfo{
				Name:          topic,
				NumPartitions: len(partitions),
				Partitions:    partitions,
			}
			s.TopicInfos[topic] = info
			continue
		}

		// or the value already exists
		info := s.TopicInfos[topic]
		info.NumPartitions = len(partitions)
		info.Partitions = partitions
		s.TopicInfos[topic] = info
	}
}

func (s *TopicScreen) WillShow(screen Screen) {
	s.Topics, _ = s.client.Topics()
	s.refreshTopicIndex()
	s.filter()
}

func (s *TopicScreen) Refresh(screen Screen) {
	log.Println("TopicScreen.Refresh")

	termbox.Clear(coldef, coldef)
	w, h = termbox.Size()

	s.drawHeader(screen)

	s.drawContent(screen, w, h)

	termbox.HideCursor()
	termbox.Flush()
}

func (s *TopicScreen) drawHeader(screen Screen) {
	w, _ = termbox.Size()

	summary := "Number of Topics: " + strconv.Itoa(len(s.Topics))

	screen.Print(summary, 0, 0, coldef, coldef)
	screen.Print(s.Query, 0, 1, termbox.ColorBlue, coldef)

	widthForTopic := strconv.Itoa(w - 20)
	titles := fmt.Sprintf("     %-"+widthForTopic+"s %11s", "TOPIC", "PARTITIONS")

	screen.Print(titles, 0, 2, coldef, coldef)
}

func (s *TopicScreen) drawContent(screen Screen, w int, h int) {
	sort.Sort(s.FilteredTopics)

	// the position of the last item show in the content area
	lastPos := s.Position + h - 3
	if lastPos > len(s.FilteredTopics) {
		lastPos = len(s.FilteredTopics)
	}

	for i := s.Position; i < lastPos; i++ {
		topic := s.FilteredTopics[i]
		parts := ""
		if info, ok := s.TopicInfos[topic]; ok {
			parts = strconv.Itoa(info.NumPartitions)
		}

		w := strconv.Itoa(w - 25)
		line := fmt.Sprintf("%-"+w+"s %16s", topic, parts)

		screen.Print(line, 5, i-s.Position+3, coldef, coldef)
	}

	// draw Cursor
	cursor := " -> "
	screen.Print(cursor, 0, s.Cursor-s.Position+3, coldef, coldef)

}

func (ts *TopicScreen) OnKeyInput(screen Screen, keyEvent termbox.Event) {

	// get the screen height
	_, h = termbox.Size()

	switch keyEvent.Type {
	case termbox.EventKey:
		switch keyEvent.Key {
		case termbox.KeyEnter, termbox.KeyArrowRight:
			// navigate to TopicPartition screen
			topic := ts.FilteredTopics[ts.Cursor]
			log.Println("Select topic at cursor: " + strconv.Itoa(ts.Cursor) + ", name: " + topic)
			topicPartitionScreen := NewTopicPartitionScreen(ts.cluster, ts.client, topic, ts.broker)
			screen.Push(topicPartitionScreen)

		case termbox.KeyArrowDown:
			// cursor cannot pass the last item
			if ts.Cursor == len(ts.FilteredTopics)-1 {
				return
			}
			// if cursor is still in the same screen, only move cursor
			// otherwise move the page down to half page
			ts.Cursor++
			if ts.Cursor > ts.Position+h-5 {
				ts.Position += h / 2
			}
			ts.Refresh(screen)

		case termbox.KeyArrowUp:
			if ts.Cursor == 0 {
				return
			}

			// if cursur moved up across position, page up
			ts.Cursor--
			if ts.Cursor < ts.Position {
				ts.Position -= h / 2
			}
			ts.Refresh(screen)

		case termbox.KeyCtrlF, termbox.KeyPgdn:
			log.Println("position before Pgdn:" + strconv.Itoa(ts.Position))
			pg := h - 3

			// if there is only one screen, do nothing
			if len(ts.Topics) <= pg {
				return
			}

			ts.Position += pg
			ts.Cursor += pg
			if ts.Position >= len(ts.FilteredTopics) {
				ts.Position = len(ts.FilteredTopics) - (pg / 2)
			}
			log.Println("position after Pgdn:" + strconv.Itoa(ts.Position))
			ts.Refresh(screen)

		case termbox.KeyCtrlB, termbox.KeyPgup:
			pg := h - 3

			// if there is only one screen, do nothing
			if len(ts.FilteredTopics) <= pg {
				return
			}

			ts.Position -= pg
			ts.Cursor -= pg

			if ts.Position < 0 {
				ts.Position = 0
			}
			ts.Refresh(screen)

		case termbox.KeyCtrlQ:
			screen.ExitChan <- true

		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(ts.Query) > 0 {
				ts.Query = ts.Query[0 : len(ts.Query)-1]
			}
			ts.filter()
			ts.Refresh(screen)

		default:
			// only take leters
			if keyEvent.Ch >= 'a' && keyEvent.Ch <= 'z' ||
				keyEvent.Ch >= 'A' && keyEvent.Ch <= 'Z' ||
				keyEvent.Ch >= '0' && keyEvent.Ch <= '9' {
				ts.Query += string(keyEvent.Ch)
				ts.filter()
				ts.Refresh(screen)

			}
		}
	case termbox.EventError:
		panic(keyEvent.Err)
	}
}

func Start(zkstr string) {

	kafkaCluster, err := NewCluster(zkstr)
	if err != nil {
		panic(err)
	}

	seedBroker := kafkaCluster.SeedBroker()

	log.Println("Seedbroker: " + seedBroker)

	// initialize logic
	client, err := sarama.NewClient([]string{seedBroker}, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer client.Close()

	content := NewTopicScreen(kafkaCluster, client, seedBroker)

	topicScreen := NewScreen(content)

	log.Println("showing the topic screen now")
	topicScreen.Show()
	topicScreen.WaitForExit()
}

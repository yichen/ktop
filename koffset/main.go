package main

import (
	"fmt"
	"os"

	"github.com/Shopify/sarama"
)

func main() {
	var seedBrokerStr string
	args := os.Args
	if len(args) == 1 {
		seedBrokerStr = "eat1-app1252.corp.linkedin.com:10251"
	} else {
		seedBrokerStr = args[1]
	}

	seedBroker := sarama.NewBroker(seedBrokerStr)
	config := sarama.NewConfig()

	err := seedBroker.Open(config)
	if err != nil {
		panic(err)
	}
	defer seedBroker.Close()

	offsetRequest := &sarama.OffsetRequest{}

	for i := 0; i < 8; i++ {
		offsetRequest.AddBlock("TestDBForKET", int32(i), sarama.LatestOffsets, 1)
	}

	offsetResponse, err := seedBroker.GetAvailableOffsets("ktop", offsetRequest)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 8; i++ {
		offsetResponseBlock := offsetResponse.GetBlock("TestDBForKET", int32(i))
		for _, offset := range offsetResponseBlock.Offsets {
			fmt.Println(offset)
		}
		fmt.Println("--")
	}

}

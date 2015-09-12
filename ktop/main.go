package main

import (
	"os"

	"bitbucket.org/yichen/ktop"
)

func main() {
	// zk: eat1-app397.stg.linkedin.com:12913/kafka-cluster
	var zkstr string
	args := os.Args
	if len(args) == 1 {
		// fmt.Println("Wrong argument. A seed broker URL is required.")
		//  "eat1-app1252.corp.linkedin.com:10251"
		// zkstr = "eat1-app397.stg.linkedin.com:12913/kafka-cluster"
		zkstr = "zk-ei1-kafka.stg.linkedin.com:12913/kafka-espresso-testing"
	} else {
		zkstr = args[1]
	}

	ktop.Start(zkstr)
}

package ktop

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type zkBrokerNode struct {
	Endpoints []string `json:"endpoints"`
	Host      string   `json:"host"`
	JmxPort   int      `json:"jmx_port"`
	Port      int      `json:"port"`
	Timestamp string   `json:"timestamp"`
	Version   int      `json:"version"`
}

type Cluster struct {
	Name       string
	zkconn     *zk.Conn
	brokers    map[string]zkBrokerNode
	keyBuilder KeyBuilder
}

func NewCluster(zkstr string) (*Cluster, error) {
	urlParts := strings.Split(zkstr, "/")
	if len(urlParts) != 2 {
		return nil, errors.New("Wrong Zookeeper URL")
	}

	conn, _, err := zk.Connect([]string{urlParts[0]}, time.Second*30)
	if err != nil {
		return nil, err
	}

	c := &Cluster{
		zkconn:     conn,
		Name:       urlParts[1],
		keyBuilder: KeyBuilder{urlParts[1]},
		brokers:    make(map[string]zkBrokerNode),
	}

	c.getBrokers()

	return c, nil
}

func (c *Cluster) getBrokers() {

	log.Println("finding all broker ids under zookeeper path: " + c.keyBuilder.brokers())

	brokerIDs, _, err := c.zkconn.Children(c.keyBuilder.brokers())
	if err != nil {

		panic(err)
	}

	for _, ID := range brokerIDs {
		// get the broker znode
		zn, _, err := c.zkconn.Get(c.keyBuilder.broker(ID))
		if err != nil {
			fmt.Println("### " + c.keyBuilder.broker(ID))
			panic(err)
		}

		log.Println("broker znode: " + string(zn))

		bn := zkBrokerNode{}
		json.Unmarshal([]byte(zn), &bn)
		c.brokers[ID] = bn
		log.Println("Broker: " + ID + ", Host: " + bn.Host + ", Port: " + strconv.Itoa(bn.Port))
	}
}

func (c *Cluster) Broker(ID string) string {
	return c.brokers[ID].Host + ":" + strconv.Itoa(c.brokers[ID].Port)
}

func (c *Cluster) SeedBroker() string {
	for _, bn := range c.brokers {
		return bn.Host + ":" + strconv.Itoa(bn.Port)
	}

	return ""
}

func (c *Cluster) Topics() []string {
	return []string{}
}

func (c *Cluster) Consumers() []string {
	return []string{}
}

func (c *Cluster) Close() {
	if c.zkconn != nil {
		c.zkconn.Close()
	}
}

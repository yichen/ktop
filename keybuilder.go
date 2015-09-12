package ktop

import "fmt"

type KeyBuilder struct {
	ClusterID string
}

func (k *KeyBuilder) cluster() string {
	return fmt.Sprintf("/%s", k.ClusterID)
}

func (k *KeyBuilder) brokers() string {
	if k.ClusterID == "" {
		return "/brokers/ids"
	}

	return fmt.Sprintf("/%s/brokers/ids", k.ClusterID)
}

func (k *KeyBuilder) broker(id string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/brokers/ids/%s", id)
	}
	return fmt.Sprintf("/%s/brokers/ids/%s", k.ClusterID, id)
}

func (k *KeyBuilder) topics() string {
	if k.ClusterID == "" {
		return "/brokers/topics"
	}
	return fmt.Sprintf("/%s/brokers/topics", k.ClusterID)
}

func (k *KeyBuilder) topic(name string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/brokers/topics/%s", name)
	}
	return fmt.Sprintf("/%s/brokers/topics/%s", k.ClusterID, name)
}

func (k *KeyBuilder) partitions(topic string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/brokers/topics/%s/partitions", topic)
	}
	return fmt.Sprintf("/%s/brokers/topics/%s/partitions", k.ClusterID, topic)
}

func (k *KeyBuilder) partition(topic string, partitionID string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/brokers/topics/%s/partitions/%s", topic, partitionID)
	}
	return fmt.Sprintf("/%s/brokers/topics/%s/partitions/%s", k.ClusterID, topic, partitionID)
}

func (k *KeyBuilder) partitionState(topic string, partitionID string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("brokers/topics/%s/partitions/%s/state", topic, partitionID)
	}
	return fmt.Sprintf("/%s/brokers/topics/%s/partitions/%s/state", k.ClusterID, topic, partitionID)
}

func (k *KeyBuilder) consumers() string {
	if k.ClusterID == "" {
		return "/consumers"
	}
	return fmt.Sprintf("/%s/consumers", k.ClusterID)
}

func (k *KeyBuilder) consumer(name string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/consumers/%s", name)
	}
	return fmt.Sprintf("/%s/consumers/%s", k.ClusterID, name)
}

func (k *KeyBuilder) consumerOffsets(consumer string, topic string) string {
	if k.ClusterID == "" {
		return fmt.Sprintf("/consumers/%s/offsets/%s", consumer, topic)
	}
	return fmt.Sprintf("/%s/consumers/%s/offsets/%s", k.ClusterID, consumer, topic)
}

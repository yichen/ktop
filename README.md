Topic Browser for Kafka
======

Ktop is a topic browser for Kafka. It makes it easier to list all topics in a kafka cluster, quickly find topics using typeahead, and inspect topic metadata. 

# Install

```shell
git clone https://github.com/yichen/ktop.git
cd ktop/ktop
go install
```

# Usages


```shell
ktop {zookeeperserver:port}/{kafkacluster}
```

This will start a console app listing all topics. Start typing to take advantage of typeahead filtering.

To exit the problem, use Ctrl-Q

To page down, use page-down key, or Ctrl-F
To page up, use page-up key, or Ctrl-B

Use the arrow key to nevigate to specific topic, and enter key to inspect the topic.



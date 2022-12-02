package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
	"logTransfer/es"
)

type KafkaClient struct {
	client sarama.Consumer
	addrs  []string
	topic  string
}

var (
	kafkaClient *KafkaClient
)

// Init 初始化client
func Init(addrs []string, topic string) (err error) {
	consumer, err := sarama.NewConsumer(addrs, nil)
	if err != nil {
		fmt.Printf("fail to start consumer, err:%v\n", err)
		return
	}
	kafkaClient = &KafkaClient{
		client: consumer,
		addrs:  addrs,
		topic:  topic,
	}
	return
}

// Run 将Kafka数据发往ES
func Run() {
	partitionList, err := kafkaClient.client.Partitions(kafkaClient.topic) // 根据 topic 取到所有的分区
	if err != nil {
		fmt.Printf("fail to get list of partition, err: %v\n", err)
		return
	}
	fmt.Println("分区: ", partitionList)
	// 1. 遍历所有的分区
	for partition := range partitionList {
		// 2. 针对每个分区创建一个对应的分区消费者
		var pc sarama.PartitionConsumer
		pc, err = kafkaClient.client.ConsumePartition(kafkaClient.topic, int32(partition), sarama.OffsetNewest)
		if err != nil {
			fmt.Printf("fail to start consumer for partition %d, err:%v\n", partition, err)
			return
		}
		defer pc.AsyncClose()

		// 3. 异步从每个分区消费信息
		go func(sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				fmt.Printf("Partition:%d Offset:%d Key:%s Value:%s\n", msg.Partition, msg.Offset, msg.Key, msg.Value)
				log_data := es.LogData{
					Topic: kafkaClient.topic,
					Data:  string(msg.Value),
				}
				es.SendToChan(log_data)
			}
		}(pc)
	}

	defer kafkaClient.client.Close()
	select {}
}

package kafka

import (
	"fmt"
	"github.com/Shopify/sarama"
)

var (
	client      sarama.SyncProducer // 声明一个全局的连接kafka的生产者client
	logDataChan chan *LogData
)

type LogData struct {
	topic string
	data  string
}

// init初始化client
func Init(addrs []string, chanMaxSize int) (err error) {
	// 1. 设置配置信息
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 2. 连接kafka
	client, err = sarama.NewSyncProducer([]string{"127.0.0.1:9092"}, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return
	}

	// 初始化 logDataChan
	logDataChan = make(chan *LogData, chanMaxSize)
	// 开启后台的 goroutine, 从通道中取数据发往 kafka
	go SendToKafka()
	return
}

// SendToChan 把日志数据发送到一个内部的 channel 中
func SendToChan(topic, data string) {
	msg := &LogData{
		topic: topic,
		data:  data,
	}
	logDataChan <- msg
}

func SendToKafka() {
	for {
		select {
		case log_data := <-logDataChan:
			// 1. 构造一个消息
			msg := &sarama.ProducerMessage{}
			msg.Topic = log_data.topic
			msg.Value = sarama.StringEncoder(log_data.data)

			// 2. 发送消息
			pid, offset, err := client.SendMessage(msg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
				return
			}
			fmt.Printf("pid:%v, offset:%v\n", pid, offset)
		}
	}
}

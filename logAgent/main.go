package main

import (
	"gopkg.in/ini.v1"

	"logAgent/conf"
	"logAgent/etcd"
	"logAgent/kafka"
	"logAgent/taillog"
	"logAgent/tools"

	"fmt"
	"strings"
	"sync"
	"time"
)

var config = new(conf.Config)

func main() {
	// 0. 加载配置文件
	err := ini.MapTo(config, "./conf/config.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		return
	}

	// 1. 初始化 kafka 连接
	err = kafka.Init(strings.Split(config.Kafka.Address, ";"), config.Kafka.ChanMaxSize)
	if err != nil {
		fmt.Printf("init kafka failed, err: %v\n", err)
		return
	}
	fmt.Println("Init kafka success.")

	// 2. 初始化etcd
	err = etcd.Init(config.Etcd.Address, time.Duration(config.Etcd.Timeout)*time.Second)
	if err != nil {
		fmt.Printf("init etcd failed, err: %v\n", err)
		return
	}
	fmt.Println("Init etcd success.")

	// 2.1 获取对外 IP 地址
	ip, err := tools.GetOurboundIP()
	if err != nil {
		panic(err)
	}
	etcdConfKey := fmt.Sprintf(config.Etcd.Key, ip)
	etcdConfValue := `[{"path":"d:/tmp/nginx.log","topic":"web_log"},{"path":"d:/tmp/redis.log","topic":"redis_log"},{"path":"d:/tmp/mysql.log","topic":"mysql_log"}]`
	err = etcd.Put(etcdConfKey, etcdConfValue)
	if err != nil {
		fmt.Printf("put to etcd failed, err:%v\n", err)
		return
	}

	// 2.2 从etcd中获取日志收集项的配置信息
	logEntryConf, err := etcd.GetConf(etcdConfKey)
	if err != nil {
		fmt.Printf("etcd.GetConf failed, err:%v\n", err)
		return
	}
	fmt.Printf("get conf from etcd success, %v\n", logEntryConf)
	// 2.3 派一个哨兵
	for index, value := range logEntryConf {
		fmt.Printf("index:%v value:%v\n", index, value)
	}

	// 3. 收集日志发往kafka
	taillog.Init(logEntryConf)

	var wg sync.WaitGroup
	wg.Add(1)
	go etcd.WatchConf(etcdConfKey, taillog.NewConfChan()) // 哨兵发现最新的配置信息会通知上面的通道
	wg.Wait()
}

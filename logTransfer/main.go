package main

import (
	"gopkg.in/ini.v1"
	"logTransfer/conf"
	"logTransfer/es"
	"logTransfer/kafka"

	"fmt"
	"strings"
)

func main() {
	// 0. 加载配置文件
	var cfg conf.LogTansfer
	err := ini.MapTo(&cfg, "./conf/config.ini")
	if err != nil {
		fmt.Println("init config, err:%v\n", err)
		return
	}
	fmt.Printf("cfg:%v\n", cfg)

	// 1. 初始化ES
	err = es.Init(cfg.ES.Address, cfg.ES.ChanMaxSize, cfg.ES.Workers)
	if err != nil {
		fmt.Printf("init ES client failed,err:%v\n", err)
		return
	}
	fmt.Println("init ES client success.")

	// 2. 初始化kafka
	err = kafka.Init(strings.Split(cfg.Kafka.Address, ";"), cfg.Kafka.Topic)
	if err != nil {
		fmt.Printf("init kafka consumer failed, err:%v\n", err)
		return
	}
	fmt.Println("init kafka success.")

	// 3. 从kafka取日志数据并放入channel中
	kafka.Run()
}

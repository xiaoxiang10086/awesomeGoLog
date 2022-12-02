package conf

type LogTansfer struct {
	Kafka Kafka `ini:"kafka"`
	ES    ES    `ini:"es"`
}

type Kafka struct {
	Address string `ini:"address"`
	Topic   string `ini:"topic"`
}

type ES struct {
	Address     string `ini:"address"`
	ChanMaxSize int    `ini:"chan_max_size"`
	Workers     int    `ini:"workers"`
}

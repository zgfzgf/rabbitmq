package mqengine

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type GbeConfig struct {
	DataSource DataSourceConfig `json:"dataSource"`
	Redis      RedisConfig      `json:"redis"`
	RabbitMq   RabbitMqConfig   `json:"rabbitMq"`
	ChanNum    ChanNumConfing   `json:"chanNum"`
	Log        LogConfig        `json:"log"`
}

type DataSourceConfig struct {
	DriverName        string `json:"driverName"`
	Addr              string `json:"addr"`
	Database          string `json:"database"`
	User              string `json:"user"`
	Password          string `json:"password"`
	EnableAutoMigrate bool   `json:"enableAutoMigrate"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
}

type RabbitMqConfig struct {
	Addr       string       `json:"addr"`
	User       string       `json:"user"`
	Password   string       `json:"password"`
	ReadQueue  QueueConfing `json:"readQueue"`
	StoreQueue QueueConfing `json:"storeQueue"`
	InfoQueue  QueueConfing `json:"infoQueue"`
}

type QueueConfing struct {
	QueueSuffix    string `json:"queueSuffix"`
	RouteKeySuffix string `json:"routeKeySuffix"`
	ExchangeSuffix string `json:"exchangeSuffix"`
}

type ChanNumConfing struct {
	Reader   int `json:"reader"`
	Store    int `json:"store"`
	Ack      int `json:"ack"`
	Info     int `json:"info"`
	LogStore int `json:"logStore"`
	LogInfo  int `json:"logInfo"`
}

type LogConfig struct {
	Level      int    `json:"level"`
	Filename   string `json:"filename"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

var config GbeConfig
var configOnce sync.Once

func GetConfig(filename string) *GbeConfig {
	configOnce.Do(func() {
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(bytes, &config)
		if err != nil {
			panic(err)
		}
	})
	return &config
}

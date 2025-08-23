package config

import (
	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"os"
	"payment-service/common/util"
)

var Config AppConfig

type AppConfig struct {
	Port                  int             `json:"port"`
	AppName               string          `json:"appName"`
	AppEnv                string          `json:"appEnv"`
	SignatureKey          string          `json:"signatureKey"`
	Database              Database        `json:"database"`
	RateLimiterMaxRequest float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond int             `json:"rateLimiterTimeSecond"`
	InternalService       InternalService `json:"internalService"`
	Kafka                 Kafka           `json:"kafka"`
	Midtrans              Midtrans        `json:"midtrans"`
}

type Database struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	MaxOpenConnection     int    `json:"maxOpenConnection"`
	MaxLifetimeConnection int    `json:"maxLifetimeConnection"`
	MaxIdleConnection     int    `json:"maxIdleConnection"`
	MaxIdleTime           int    `json:"maxIdleTime"`
}

type InternalService struct {
	User User `json:"user"`
}

type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Kafka struct {
	Brokers     []string `json:"brokers"`
	TimeOutInMS int      `json:"timeOutInMS"`
	MaxRetry    int      `json:"maxRetry"`
	Topic       string   `json:"topic"`
}

type Midtrans struct {
	ServerKey    string `json:"serverKey"`
	ClientKey    string `json:"clientKey"`
	IsProduction bool   `json:"isProduction"`
}

func Init() {
	err := util.BindFromJSON(&Config, "config.json", ".")
	if err != nil {
		logrus.Infof("failed to bind config: %v", err)
		err = util.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_KEY"))
		if err != nil {
			panic(err)
		}
	}
}

package config

import (
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/spf13/viper"
)

type server struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type upload struct {
	Key string `mapstructure:"key"`
}

type postgres struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type config struct {
	Server   server   `mapstructure:"server"`
	Upload   upload   `mapstructure:"upload"`
	Postgres postgres `mapstructure:"postgres"`
}

var AppConfig config

func Init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		hlog.Fatal(err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		hlog.Fatal(err)
	}
}

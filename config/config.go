package config

import (
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

type AppConfig struct {
	ServerConf  *ServerConfig
	WxBot4gConf *WxBot4gConfig
}

type ServerConfig struct {
	Mode       string `json:"mode"`
	Port       int    `json:"port"`
	AppKey     string `json:"appKey"`
	RetryTimes int    `json:"retryTimes"`
}

type WxBot4gConfig struct {
	WxQrDir      string `json:"wxqrDir"`
	HeartbeatURL string `json:"heartbeatURL"`
	ImageDir     string `json:"imageDir"`
}

var (
	Config AppConfig
)

func init() {
	if err := initConfig(); err != nil {
		panic(err)
	}
	initLog()
	watchConfig()

	Config = AppConfig{
		ServerConf: &ServerConfig{
			Mode:       viper.GetString("runmode"),
			Port:       viper.GetInt("addr"),
			AppKey:     viper.GetString("appKey"),
			RetryTimes: viper.GetInt("retryTimes"),
		},
		WxBot4gConf: &WxBot4gConfig{
			WxQrDir:      viper.GetString("wxbot4g.wxqrDir"),
			HeartbeatURL: viper.GetString("wxbot4g.heartbeatURL"),
			ImageDir:     viper.GetString("wxbot4g.imageDir"),
		},
	}
}

func initConfig() error {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("wxBot4g")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.Info("Config file changed: %s", e.Name)
	})
}

func initLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetReportCaller(true)
}

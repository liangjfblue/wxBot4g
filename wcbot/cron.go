package wcbot

//import "C"
import (
	"fmt"
	"wxBot4g/config"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/robfig/cron"
)

func InitHeartbeatCron() {
	c := cron.New()
	err := c.AddFunc("@every 180s", heartbeat)
	if err != nil {
		logrus.Error(err)
		return
	}

	c.Start()
	return
}

func heartbeat() {
	retryTimes := 0
	logrus.Debug(time.Now())
RETRY:
	urlStr := fmt.Sprintf("http://127.0.0.1:%d/v1/health/heartbeat?word=keepalive&appKey=%s",
		config.Config.ServerConf.Port, config.Config.ServerConf.AppKey)

	if _, err := http.Get(urlStr); err != nil {
		logrus.Error("wechat bot is die, now retry to send keepalive")
		if config.Config.ServerConf.RetryTimes > 0 && retryTimes < config.Config.ServerConf.RetryTimes {
			retryTimes++
			time.Sleep(time.Second)
			goto RETRY
		} else {
			logrus.Error("wechat bot is die, over send keepalive")
			//TODO 警报通知管理官，机器人挂了。钉钉/微信/企业微信/邮件
		}
	}
}

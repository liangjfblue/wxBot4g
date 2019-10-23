package wcbot

import (
	"errors"
	"wxBot4g/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Text(c *gin.Context) {
	path := c.Request.URL.Path
	logrus.Debug(path)
	if c.Request.URL.Path == "/favicon.ico" {
		return
	}

	//判断appkey
	appKey := c.Query("appKey")
	if appKey != config.Config.ServerConf.AppKey {
		c.Status(http.StatusBadRequest)
		_, _ = c.Writer.Write(nil)
	}

	//消息处理
	if err := handleMsg(c); err != nil {
		c.Status(http.StatusBadRequest)
		_, _ = c.Writer.Write(nil)
	}

	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(nil)
}

func handleMsg(c *gin.Context) error {
	to := c.Query("to")
	word := c.Query("word")

	if to == "" && word == "" {
		logrus.Error("param error")
		return errors.New("param error")
	}

	if to == "" {
		if ok := WechatBot.SendMsgByUid(word, "filehelper"); !ok {
			logrus.Error("send msg error")
			return errors.New("send msg error")
		}
	} else {
		if ok := WechatBot.SendMsg(to, word, false); !ok {
			logrus.Error("send msg error")
			return errors.New("send msg error")
		}
	}
	return nil
}

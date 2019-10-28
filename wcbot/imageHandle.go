package wcbot

import (
	"net/http"
	"os"
	"path"
	"wxBot4g/config"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

/*
  to: 目的好友/群
  image: 通过上传接口
*/
func ImageHandle(c *gin.Context) {
	//消息处理
	if err := handleImageMsg(c); err != nil {
		c.Status(http.StatusBadRequest)
		_, _ = c.Writer.Write(nil)
	}

	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(nil)
}

func handleImageMsg(c *gin.Context) error {
	to := c.Query("to")

	file, err := c.FormFile("file")
	if err != nil {
		logrus.Error(err)
		return err
	}

	filename := path.Base(file.Filename)
	filename = config.Config.WxBot4gConf.ImageDir + filename
	err = c.SaveUploadedFile(file, filename)
	if err != nil {
		logrus.Error(err)
		return err
	}

	if _, err := os.Stat(filename); err == nil {
		if to == "" {
			if err := WechatBot.SendMediaMsgByUid(filename, "filehelper"); err != nil {
				logrus.Error(err)
				return err
			}
		} else {
			if err := WechatBot.SendMedia(filename, to); err != nil {
				logrus.Error(err)
				return err
			}
		}
		if err = os.Remove(filename); err != nil {
			logrus.Error(err)
			return err
		}
		return nil
	} else {
		logrus.Error(err)
		return err
	}
}

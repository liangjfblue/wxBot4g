package wcbot

import (
	"net/http"
	"wxBot4g/config"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logrus.Debug(c.Request.URL)
		if c.Request.URL.Path == "/favicon.ico" {
			c.Abort()
			return
		}

		appKey := c.Query("appKey")
		if appKey != config.Config.ServerConf.AppKey {
			c.Status(http.StatusBadRequest)
			_, _ = c.Writer.Write(nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

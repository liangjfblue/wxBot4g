package main

import (
	"wxBot4g/models"
	"wxBot4g/pkg/define"
	"wxBot4g/wcbot"

	"github.com/sirupsen/logrus"
)

func HandleMsg(msg models.RealRecvMsg) {
	logrus.Info(
		"消息类型:", define.MsgIdString(msg.MsgTypeId), " ",
		"数据类型:", define.MsgTypeIdString(msg.Content.Type), " ",
		"发送人:", msg.SendMsgUSer.Name, " ",
		"内容:", msg.Content.Data)
}

func main() {
	bot := wcbot.New(HandleMsg)
	bot.Debug = true
	bot.Run()
}

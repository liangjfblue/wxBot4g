package main

import (
	"wxBot4g/models"
	"wxBot4g/pkg/define"
	"wxBot4g/wcbot"

	"github.com/sirupsen/logrus"
)

var (
	Bot *wcbot.WcBot
)

type WeChatBot struct {
}

func (w *WeChatBot)HandleMessage(msg models.RealRecvMsg) {
	//过滤不支持消息99
	if msg.MsgType == 99 || msg.MsgTypeId == 99 {
		return
	}

	//获取unknown的username
	contentUser := msg.Content.User.Name
	if msg.Content.User.Name == "unknown" {
		contentUser = Bot.GetGroupUserName(msg.Content.User.Uid)
	}

	logrus.Debug(
		"消息类型:", define.MsgIdString(msg.MsgTypeId), " ",
		"数据类型:", define.MsgTypeIdString(msg.Content.Type), " ",
		"发送者:", msg.FromUserName, " ",
		"发送人:", msg.SendMsgUSer.Name, " ",
		"发送内容人:", contentUser, " ",
		"内容:", msg.Content.Data)
}

func main() {
	Bot = wcbot.New()
	Bot.Debug = true
	//Bot.QrCodeInTerminal() //默认在 wxqr 目录生成二维码，调用此函数，在终端打印二维码

	Bot.AddHandler(&WeChatBot{})

	Bot.Run()
}

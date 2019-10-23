wxBot4g 是基于go的微信机器人

本项目长期维护更新，欢迎star、fork、 pull requests、 issue

## 来源
[wxBot](https://github.com/liuwons/wxBot)是一个非常优秀的开源微信个人号接口，使用Python语言开发。[wxBot4g](https://github.com/liuwons/wxBot)是[wxBot](https://github.com/liuwons/wxBot)基于go的版本。

## 项目用途
- 把个人微信号扩展为神奇的"机器人"，自动回复？哄女朋友？just小case。
- 集成在你自己的项目中，为应用提供微信机器人的能力。
- 记录自己客服群的情况，不错过每个好评，更不错过每个差评。
- 控制智能设备，真正实现“物联网”。

## 特点
- 1、简单易用。实现消息回调函数，实例化wcbot，wcbot.run()即可。
- 2、稳定不断线。底层实现了“心跳”维持微信会话机制，对用户透明。
- 3、提供restful api。支持发送消息到指定群和指定好友。
- 4、依赖包少而简洁。

## 技术
- gin（http框架）
- cron（定时任务）
- etree（解析xml）
- viper（配置文件读取）
- logrus（日志框架）
- go-qrcode（登陆二维码生成）

## 目前支持的消息类型
### 好友消息
- [x] 文本
- [x] 图片
- [x] 地理位置
- [x] 个人名片
- [x] 语音
- [x] 小视频
- [ ] 动画

### 群消息
- [x] 文本
- [x] 图片
- [x] 地理位置
- [x] 个人名片
- [x] 语音
- [ ] 动画

### TODO功能
- [ ] 提供restful api，发送消息到指定好友/群
- [ ] 文件/图片上传阿里云oss
- [ ] 监听指定群报警
- [ ] 聊天记录中文分析，情感分析

## 使用例子
24行代码就实现微信机器人的监听消息功能

    package main

    import (
        "wxBot4g/models"
        "wxBot4g/pkg/define"
        "wxBot4g/wcbot"
    
        "github.com/sirupsen/logrus"
    )
    
    func HandleMsg(msg models.RealRecvMsg) {
        logrus.Debug("MsgType: ", msg.MsgType, " ", " MsgTypeId: ", msg.MsgTypeId)
        logrus.Info(
            "消息类型:", define.MsgIdString(msg.MsgType), " ",
            "数据类型:", define.MsgTypeIdString(msg.MsgTypeId), " ",
            "发送人:", msg.SendMsgUSer.Name, " ",
            "内容:", msg.Content.Data)
    }
    
    func main() {
        bot := wcbot.New(HandleMsg)
        bot.Debug = true
        bot.Run()
    }


## 消息类型和数据类型

### MsgType（消息类型）

数据类型编号|数据类型|说明
--|--|--|
0|Init|初始化消息，内部数据
1|Self|自己发送的消息
2|FileHelper|文件消息
3|Group|群消息
4|Contact|联系人消息
5|Public|公众号消息
6|Special|特殊账号消息
51|获取wxid|获取wxid消息
99|Unknown|未知账号消息


### MsgTypeId（数据类型）

数据类型编号|数据类型|说明
--|--|--|
0|Text|文本消息的具体内容
1|Location|地理位置
3|Image|图片数据的url，HTTP POST请求此url可以得到jpg文件格式的数据
4|Voice|语音数据的url，HTTP POST请求此url可以得到mp3文件格式的数据
5|Recommend|包含 nickname (昵称)， alias (别名)，province (省份)，city (城市)， gender (性别)字段
6|Animation|动画url, HTTP POST请求此url可以得到gif文件格式的数据
7|Share|字典，包含 type (类型)，title (标题)，desc (描述)，url (链接)，from (源网站)字段
8|Video|视频，未支持
9|VideoCall|视频电话，未支持
10|Redraw|撤回消息
11|Empty|内容，未支持
99|Unknown|未支持


## 参考
- [挖掘微信Web版通信的全过程](http://www.tanhao.me/talk/1466.html/)
- [Python网页微信API](https://github.com/liuwons/wxBot)
- [微信个人号机器人](https://github.com/newflydd/itchat4go)

wxBot4g 是基于go的微信机器人

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
- [x] 提供restful api，发送消息到指定好友/群
- [ ] 文件/图片上传阿里云oss
- [ ] 监听指定群报警
- [ ] 聊天记录中文分析，情感分析

## 使用例子

24行代码就实现微信机器人的监听消息功能

```
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
```

## 消息类型和数据类型

### MsgType（消息类型）

| 数据类型编号 | 数据类型   | 说明                 |
| ------------ | ---------- | -------------------- |
| 0            | Init       | 初始化消息，内部数据 |
| 1            | Self       | 自己发送的消息       |
| 2            | FileHelper | 文件消息             |
| 3            | Group      | 群消息               |
| 4            | Contact    | 联系人消息           |
| 5            | Public     | 公众号消息           |
| 6            | Special    | 特殊账号消息         |
| 51           | 获取wxid   | 获取wxid消息         |
| 99           | Unknown    | 未知账号消息         |

### MsgTypeId（数据类型）

| 数据类型编号 | 数据类型  | 说明                                                         |
| ------------ | --------- | ------------------------------------------------------------ |
| 0            | Text      | 文本消息的具体内容                                           |
| 1            | Location  | 地理位置                                                     |
| 3            | Image     | 图片数据的url，HTTP POST请求此url可以得到jpg文件格式的数据   |
| 4            | Voice     | 语音数据的url，HTTP POST请求此url可以得到mp3文件格式的数据   |
| 5            | Recommend | 包含 nickname (昵称)， alias (别名)，province (省份)，city (城市)， gender (性别)字段 |
| 6            | Animation | 动画url, HTTP POST请求此url可以得到gif文件格式的数据         |
| 7            | Share     | 字典，包含 type (类型)，title (标题)，desc (描述)，url (链接)，from (源网站)字段 |
| 8            | Video     | 视频，未支持                                                 |
| 9            | VideoCall | 视频电话，未支持                                             |
| 10           | Redraw    | 撤回消息                                                     |
| 11           | Empty     | 内容，未支持                                                 |
| 99           | Unknown   | 未支持                                                       |

## 功能api

### 发送文本消息(好友/群)

```
http://127.0.0.1:7788/v1/msg/text?to=测试群&word=你好, 测试一下&appKey=khr1244o1oh
```

### 发送图片消息(好友/群)

请参考`wxBot4g/wcbot/imageHandle_test.go`

v1.1

- 增加终端二维码扫码登录
- 增加api，发送文本、图片消息到指定群
- 增加单元测试
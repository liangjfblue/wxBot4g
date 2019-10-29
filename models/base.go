package models

import "fmt"

// 同步数据和keys
type SyncKeysJsonData struct {
	Count    int       `json:"Count"`
	SyncKeys []SyncKey `json:"List"`
}

type SyncKey struct {
	Key int64 `json:"Key"`
	Val int64 `json:"Val"`
}

func (sks SyncKeysJsonData) ToString() string {
	resultStr := ""

	for i := 0; i < sks.Count; i++ {
		resultStr = resultStr + fmt.Sprintf("%d_%d|", sks.SyncKeys[i].Key, sks.SyncKeys[i].Val)
	}

	return resultStr[:len(resultStr)-1]
}

// RecvMsgs 微信消息对象列表
type RecvMsgs struct {
	MsgCount        int              `json:"AddMsgCount"`
	MsgList         []RecvMsg        `json:"AddMsgList"`
	SyncKeys        SyncKeysJsonData `json:"SyncKey"`
	ModContactCount int              `json:"ModContactCount"`
	ModContactList  []interface{}    `json:"ModContactList"`
}

// RecvMsg 微信消息对象
type RecvMsg struct {
	MsgId                string      `json:"MsgId"`
	FromUserName         string      `json:"FromUserName"`
	ToUserName           string      `json:"ToUserName"`
	MsgType              int         `json:"MsgType"`
	Content              string      `json:"Content"`
	CreateTime           int64       `json:"CreateTime"`
	RecommendInfo        interface{} `json:"RecommendInfo"`
	FileName             string      `json:"FileName"`
	AppMsgType           int         `json:"AppMsgType"`
	StatusNotifyCode     int         `json:"StatusNotifyCode"`
	StatusNotifyUserName string      `json:"StatusNotifyUserName"`
	Url                  string      `json:"Url"`
}

// RecvMsg 微信消息对象
type RealRecvMsg struct {
	MsgTypeId    int     `json:"MsgTypeId"`
	MsgId        string  `json:"MsgId"`
	FromUserName string  `json:"FromUserName"`
	ToUserName   string  `json:"ToUserName"`
	MsgType      int     `json:"MsgType"` //消息类型
	Content      Content `json:"Content"`
	CreateTime   int64   `json:"CreateTime"`
	SendMsgUSer  MsgUser `json:"SendMsgUSer"`
}

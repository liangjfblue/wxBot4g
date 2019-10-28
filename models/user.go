package models

/**
User结构
{
    "Uin":0,
    "UserName":"@d662309d027cb60f57991e409c4dd59a02cc718df7ab45fe9448c94e90f0581e",
    "NickName":"小二",
    "HeadImgUrl":"/cgi-bin/mmwebwx-bin/webwxgeticon?seq=693951199u0026username=@d662309d026skey=@crypt_7914bf7f_7772e6fd5ceb2c04c8",
    "ContactFlag":42343,
    "MemberCount":0,
    "MemberList":[

    ],
    "RemarkName":"行行行423",
    "HideInputBarFlag":0,
    "Sex":2,
    "Signature":"有亲友可待",
    "VerifyFlag":0,
    "OwnerUin":0,
    "PYInitial":"YE",
    "PYQuanPin":"darwr",
    "RemarkPYInitial":"QXMY",
    "RemarkPYQuanPin":"4654465",
    "StarFriend":0,
    "AppAccountFlag":0,
    "Statues":0,
    "AttrStatus":242173,
    "Province":"广东",
    "City":"广州",
    "Alias":"",
    "SnsFlag":49,
    "UniFriend":0,
    "DisplayName":"",
    "ChatRoomId":0,
    "KeyWord":"",
    "EncryChatRoomId":"",
    "IsOwner":0
}
*/
// ContactList 微信获取所有联系人列表
type ContactList struct {
	Seq         int    `json:"Seq"`
	MemberCount int    `json:"MemberCount"`
	MemberList  []User `json:"MemberList"`
}

// GroupList 微信获取所有群列表
type GroupList struct {
	MemberCount int    `json:"MemberCount"`
	MemberList  []User `json:"ContactList"`
}

// User 微信通用User结构
type User struct {
	Uin             int64  `json:"Uin"`
	UserName        string `json:"UserName"`
	NickName        string `json:"NickName"`
	DisplayName     string `json:"DisplayName"`
	RemarkName      string `json:"RemarkName"`
	Sex             int8   `json:"Sex"`
	Province        string `json:"Province"`
	City            string `json:"City"`
	VerifyFlag      int    `json:"VerifyFlag"`
	Signature       string `json:"Signature"`       //个性签名
	EncryChatRoomId string `json:"EncryChatRoomId"` //群id
}

//accountInfo通信录
type AccountInfo struct {
	Type  string      `json:"type"`
	User  User        `json:"user"`
	Group interface{} `json:"group"`
}

// MsgUser 消息User结构
type MsgUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ContentUser 消息User结构
type ContentUser struct {
	Uid  string `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
}

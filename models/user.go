package models

// User 微信通用User结构
type User struct {
	Uin        int64  `json:"Uin"`
	UserName   string `json:"UserName"`
	NickName   string `json:"NickName"`
	RemarkName string `json:"RemarkName"`
	Sex        int8   `json:"Sex"`
	Province   string `json:"Province"`
	City       string `json:"City"`
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

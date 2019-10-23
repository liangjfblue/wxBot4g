package define

var (
	MsgIdList     = make(map[int]string) //消息类型
	MsgTypeIdList = make(map[int]string) //消息数据类型
)

func init() {
	MsgIdList = map[int]string{
		0:  "init data",  //初始化消息，内部数据
		1:  "Self",       //自己发送的消息
		2:  "FileHelper", //文件消息
		3:  "Group",      //群消息
		4:  "Contact",    //联系人消息
		5:  "Public",     //公众号消息
		6:  "Special",    //特殊账号消息
		51: "pull wxid",  //获取wxid消息
		99: "Unknown",    //	未知账号消息}
	}
	MsgTypeIdList = map[int]string{
		0:  "Text",
		1:  "Location",
		3:  "Image",
		4:  "Voice",
		5:  "Recommend",
		6:  "Animation",
		7:  "Share",
		8:  "Video",
		9:  "VideoCall",
		10: "Redraw",
		11: "Empty",
		99: "Unknown",
	}
}

func MsgIdString(msgId int) string {
	if typeName, ok := MsgIdList[msgId]; ok {
		return typeName
	} else {
		return "Unknown"
	}
}

func MsgTypeIdString(msgTypeId int) string {
	if typeName, ok := MsgTypeIdList[msgTypeId]; ok {
		return typeName
	} else {
		return "Unknown"
	}
}

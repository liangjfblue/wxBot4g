package wcbot

import (
	"fmt"
	"regexp"
	"strings"
	"wxBot4g/models"

	"github.com/sirupsen/logrus"
)

func (wc *WcBot) isContact(uid string) bool {
	for _, contact := range wc.contactList {
		if uid == contact.UserName {
			return true
		}
	}
	return false
}

func (wc *WcBot) isPublic(uid string) bool {
	for _, contact := range wc.publicList {
		if uid == contact.UserName {
			return true
		}
	}
	return false
}

func (wc *WcBot) isSpecial(uid string) bool {
	for _, contact := range wc.specialList {
		if uid == contact.UserName {
			return true
		}
	}
	return false
}

func (wc *WcBot) getContactInfo(uid string) models.AccountInfo {
	if _, ok := wc.accountInfo["normal_member"][uid]; ok {
		return wc.accountInfo["normal_member"][uid]
	}
	return models.AccountInfo{}
}

func (wc *WcBot) getGroupMemberInfo(uid string) models.AccountInfo {
	if _, ok := wc.accountInfo["group_member"][uid]; ok {
		return wc.accountInfo["group_member"][uid]
	}
	return models.AccountInfo{}
}

func (wc *WcBot) getContactName(uid string) *models.GroupMember {
	info := wc.getContactInfo(uid)

	//info = info["info"].(map[string]interface{})

	var groupMember models.GroupMember
	if info.User.RemarkName != "" {
		groupMember.RemarkName = info.User.RemarkName
	}

	if info.User.DisplayName != "" {
		groupMember.DisplayName = info.User.DisplayName
	}

	if info.User.NickName != "" {
		groupMember.Nickname = info.User.NickName
	}
	return &groupMember
}

func (wc *WcBot) getContactPreferName(name *models.GroupMember) string {
	if name == nil {
		return ""
	}
	if name.RemarkName != "" {
		return name.RemarkName
	}

	if name.DisplayName != "" {
		return name.DisplayName
	}

	if name.Nickname != "" {
		return name.Nickname
	}
	return ""
}

func (wc *WcBot) getGroupMemberPreferName(name *models.GroupMember) string {
	if name == nil {
		return ""
	}
	if name.RemarkName != "" {
		return name.RemarkName
	}

	if name.DisplayName != "" {
		return name.DisplayName
	}

	if name.Nickname != "" {
		return name.Nickname
	}
	return ""
}

/**
getGroupMemberName 获取群聊中指定成员的名称信息

param gid: 群id
param uid: 群聊成员id
return: 名称信息，类似 {"display_name": "test_user", "nickname": "test", "remark_name": "for_test" }
*/
func (wc *WcBot) getGroupMemberName(gid, uid string) *models.GroupMember {
	groups, ok := wc.groupMembers[gid]
	if !ok {
		return nil
	}
	for _, group := range groups {
		if group.UserName == uid {
			groupMember := new(models.GroupMember)
			if group.RemarkName != "" {
				groupMember.RemarkName = group.RemarkName
			}

			if group.DisplayName != "" {
				groupMember.DisplayName = group.DisplayName
			}

			if group.NickName != "" {
				groupMember.Nickname = group.NickName
			}

			return groupMember
		}
	}
	return nil
}

func (wc *WcBot) GetGroupUserName(uId string) string {
	if uId == "" {
		return "unknown"
	}
	if err := wc.batchGetGroupMembers(); err != nil {
		logrus.Error(err)
		return "unknown"
	}
	for _, groupUsers := range wc.groupMembers {
		for _, user := range groupUsers {
			if uId == user.UserName {
				if uId == user.DisplayName {
					return user.DisplayName
				} else if uId == user.NickName {
					return user.NickName
				} else {
					return "unknown"
				}
			}
		}
	}
	return "unknown"
}

func (wc *WcBot) searchContent(key, content, fmat string) string {
	return "unknown"
}

func (wc *WcBot) procAtInfo(msg string) (string, string, []models.Detail) {
	if msg == "" {
		return "", "", nil
	}

	segs := strings.Split(msg, `\u2005`)
	strMsgAll := ""
	strMsg := ""
	infos := make([]models.Detail, 0)
	if len(segs) > 1 {
		for i := 0; i < len(segs)-1; i++ {
			segs[i] += `\u2005`
			reg := regexp.MustCompile(`@.*\u2005`)
			pmm := reg.FindAllStringSubmatch(segs[i], -1)
			if pmm[0] != nil {
				pm := ""
				for key, value := range pmm[0] {
					if key >= 2 {
						pm = pm + value
					}
				}
				name := pm
				str := strings.Replace(segs[i], pm, "", -1)
				strMsgAll = strMsgAll + str + "@" + name + " "
				strMsg += str
				if str != "" {
					infos = append(infos, models.Detail{Type: "str", Value: str})
				}
				infos = append(infos, models.Detail{Type: "at", Value: str})
			} else {
				infos = append(infos, models.Detail{Type: "str", Value: segs[i]})
				strMsgAll += segs[i]
				strMsg += segs[i]
			}
		}
		strMsgAll = strMsgAll + segs[len(segs)-1]
		strMsg += segs[len(segs)-1]
		infos = append(infos, models.Detail{Type: "str", Value: segs[(len(segs) - 1)]})
	} else {
		infos = append(infos, models.Detail{Type: "str", Value: segs[(len(segs) - 1)]})
		strMsgAll = msg
		strMsg = msg
	}
	return strings.ReplaceAll(strMsgAll, "\u2005", ""), strings.ReplaceAll(strMsg, "\u2005", ""), infos
}
func (wc *WcBot) getMsgImgUrl(msgid string) string {
	return wc.baseUri + fmt.Sprintf(`/webwxgetmsgimg?MsgID=%s&skey=%s`, msgid, wc.sKey)
}

func (wc *WcBot) getVoiceUrl(msgid string) string {
	return wc.baseUri + fmt.Sprintf(`/webwxgetvoice?msgid=%s&skey=%s`, msgid, wc.sKey)

}

func (wc *WcBot) getVideoUrl(msgid string) string {
	return wc.baseUri + fmt.Sprintf(`/webwxgetvideo?msgid=%s&skey=%s`, msgid, wc.sKey)
}

func (wc *WcBot) sendMsgImgAliyun(msgid, uin string) string {
	return ""
}

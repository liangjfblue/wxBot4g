package wcbot

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (wc *WcBot) SendMsgByUid(word, dst string) bool {
	urlStr := wc.baseUri + fmt.Sprintf("/webwxsendmsg?pass_ticket=%s", wc.passTicket)

	msgId := strconv.Itoa(int(time.Now().UnixNano()/int64(time.Millisecond))) +
		strings.Replace(fmt.Sprintf("%f", rand.Float64()), ".", "", -1)

	body := struct {
		BaseRequest interface{} `json:"BaseRequest"`
		Msg         interface{} `json:"Msg"`
	}{
		BaseRequest: wc.baseRequest,
		Msg: map[string]interface{}{
			"Type":         1,
			"Content":      word,
			"FromUserName": wc.myAccount["UserName"],
			"ToUserName":   dst,
			"LocalID":      msgId,
			"ClientMsgId":  msgId,
		},
	}

	data, err := wc.httpClient.Post(urlStr, body)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	mData := struct {
		Ret int `json:"Ret"`
	}{}
	if err := json.Unmarshal(data, &mData); err != nil {
		logrus.Error(err.Error())
		return false
	}

	ret := mData.Ret == 0

	return ret
}

func (wc *WcBot) SendMsg(name, word string, isFile bool) bool {
	uId := wc.GetUserId(name)
	if uId != "" {
		if isFile {
			f, err := os.Open(word)
			if err != nil {
				logrus.Error(err)
				return false
			}
			defer f.Close()
			rd := bufio.NewReader(f)
			result := true
			for {
				line, err := rd.ReadString('\n')
				if err != nil || io.EOF == err {
					break
				}
				logrus.Debug("-> " + name + ": " + line)
				result = wc.SendMsgByUid(line, uId)
				if !result {
					logrus.Error("send msg by uid error")
				}

				time.Sleep(time.Second)
			}
			return result
		} else {
			if wc.SendMsgByUid(word, uId) {
				return true
			} else {
				return false
			}
		}
	} else {
		logrus.Error("user is not exist")
		return false
	}
}

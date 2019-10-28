package wcbot

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/h2non/filetype.v1/types"

	"github.com/sirupsen/logrus"

	"gopkg.in/h2non/filetype.v1"
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

func (wc *WcBot) SendMediaMsgByUid(filepath, to string) error {
	info, err := os.Stat(filepath)
	if err != nil {
		logrus.Error(err)
		return err
	}

	buf, err := ioutil.ReadFile(filepath)
	if err != nil {
		logrus.Error(err)
		return err
	}

	kind, err := filetype.Get(buf)
	if err != nil {
		logrus.Error(err)
		return err
	}

	media, err := wc.UploadMedia(buf, kind, info, to)
	if err != nil {
		logrus.Error(err)
		return err
	}

	var msg = make(map[string]interface{})
	msg["FromUserName"] = wc.myAccount["UserName"].(string)
	msg["ToUserName"] = to
	msg["LocalID"] = fmt.Sprintf("%d", time.Now().Unix())
	msg["ClientMsgId"] = msg["LocalID"]

	if filetype.IsImage(buf) {
		if strings.HasSuffix(kind.MIME.Value, `gif`) {
			msg["Type"] = 47
			msg["MediaId"] = media
			msg["EmojiFlag"] = 2
			//SendMsgEmoticon(msg)
		} else {
			msg["Type"] = 3
			msg["MediaId"] = media
			if err := wc.sendMsgImage(msg); err != nil {
				return err
			}
		}
	} else {
		info, _ := os.Stat(filepath)
		if filetype.IsVideo(buf) {
			msg["Type"] = 43
			msg["MediaId"] = media
			//SendMsgVideo(msg)
		} else {
			msg["Type"] = 6
			msg[`Content`] = fmt.Sprintf(`<appmsg appid='wxeb7ec651dd0aefa9' sdkver=''><title>%s</title><des></des><action></action><type>6</type><content></content><url></url><lowurl></lowurl><appattach><totallen>10</totallen><attachid>%s</attachid><fileext>%s</fileext></appattach><extinfo></extinfo></appmsg>`, info.Name(), media, kind.Extension)
			//SendMsgFile(msg)
		}
	}

	return err
}

func (wc *WcBot) SendMedia(imagePath, toName string) error {
	to := wc.GetUserId(toName)
	if to != "" {
		if err := wc.SendMediaMsgByUid(imagePath, to); err == nil {
			return nil
		} else {
			return err
		}
	} else {
		logrus.Error("user is not exist")
		return errors.New("user is not exist")
	}
}

func (wc *WcBot) UploadMedia(buf []byte, kind types.Type, info os.FileInfo, to string) (string, error) {
	head := buf[:261]

	var mediaType string
	if filetype.IsImage(head) {
		mediaType = `pic`
	} else if filetype.IsVideo(head) {
		mediaType = `video`
	} else {
		mediaType = `doc`
	}

	fields := map[string]string{
		`id`:                `WU_FILE_` + fmt.Sprintf("%d", wc.fileIndex),
		`name`:              info.Name(),
		`type`:              kind.MIME.Value,
		`lastModifiedDate`:  info.ModTime().UTC().String(),
		`size`:              fmt.Sprintf("%d", info.Size()),
		`mediatype`:         mediaType,
		`pass_ticket`:       wc.passTicket,
		`webwx_data_ticket`: wc.httpClient.GetCookieByName("webwx_data_ticket").Value,
	}
	md5Ctx := md5.New()
	md5Ctx.Write(buf)
	cipherStr := md5Ctx.Sum(nil)
	media, err := json.Marshal(&map[string]interface{}{
		`BaseRequest`:   wc.baseRequest,
		`ClientMediaId`: fmt.Sprintf("%d", time.Now().Unix()),
		`TotalLen`:      fmt.Sprintf("%d", info.Size()),
		`StartPos`:      0,
		`DataLen`:       fmt.Sprintf("%d", info.Size()),
		`MediaType`:     4,
		`UploadType`:    2,
		`ToUserName`:    to,
		`FromUserName`:  wc.myAccount["UserName"].(string),
		`FileMd5`:       hex.EncodeToString(cipherStr),
	})

	if err != nil {
		logrus.Error(err)
		return "", err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile(`filename`, info.Name())
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	_, err = fw.Write(buf)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	for k, v := range fields {
		err = writer.WriteField(k, v)
	}
	err = writer.WriteField(`uploadmediarequest`, string(media))

	if err != nil {
		logrus.Error(err)
		return "", err
	}

	writer.Close()
	data, _ := ioutil.ReadAll(body)

	header := make(map[string]string)
	header["Content-Type"] = writer.FormDataContentType()
	wc.httpClient.SetHeader(header)

	strUrl := `https://file.wx.qq.com/cgi-bin/mmwebwx-bin/webwxuploadmedia?f=json`
	resp, err := wc.httpClient.PostMedia(strUrl, data)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	wc.httpClient.DelHeader(header)

	wc.fileIndex++

	mData := make(map[string]interface{})
	err = json.Unmarshal(resp, &mData)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	return mData["MediaId"].(string), nil
}

func (wc *WcBot) sendMsgImage(con map[string]interface{}) error {
	strUrl := fmt.Sprintf("https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxsendmsgimg?fun=async&f=json&pass_ticket=%s",
		wc.passTicket)
	jCon, err := json.Marshal(con)
	if err != nil {
		logrus.Error(err)
		return err
	}
	body := fmt.Sprintf(`{"BaseRequest":{"Uin":%s,"Sid":"%s","Skey":"%s","DeviceID":"%s"},"Msg":%s,"Scene":0}`,
		wc.uin, wc.sid, wc.sKey, wc.deviceId, jCon)

	header := make(map[string]string)
	header["Content-Type"] = `application/json;charset=UTF-8`
	wc.httpClient.SetHeader(header)

	_, err = wc.httpClient.PostString(strUrl, body)
	if err != nil {
		logrus.Error(err)
		return err
	}
	wc.httpClient.DelHeader(header)
	return nil
}

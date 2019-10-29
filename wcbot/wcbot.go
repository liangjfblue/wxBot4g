package wcbot

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"wxBot4g/config"
	"wxBot4g/models"
	"wxBot4g/pkg/httpClient"
	"wxBot4g/pkg/utils"

	"github.com/mdp/qrterminal"

	"github.com/gin-gonic/gin"

	_ "wxBot4g/config"

	"github.com/beevik/etree"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	qrcodetl "github.com/tuotoo/qrcode"
)

type HandleMsgAll func(models.RealRecvMsg)

type WcBot struct {
	Debug          bool
	QrCodeTerminal bool
	uuid           string
	baseUri        string
	baseHost       string
	redirectUri    string
	uin            string
	sid            string
	sKey           string
	passTicket     string
	deviceId       string
	Cookies        []*http.Cookie

	baseRequest         map[string]interface{}
	syncKeyStr          string
	syncKey             interface{}
	syncHost            string
	status              string
	batchCount          int      //一次拉取50个联系人的信息
	fullUserNameList    []string //直接获取不到通讯录时，获取的username列表
	wxIdList            []string //获取到的wxid的列表
	cursor              int      //拉取联系人信息的游标
	isBigContact        bool     //通讯录人数过多，无法直接获取
	tempPwd             string
	httpClient          *httpClient.Client
	conf                map[string]interface{}
	myAccount           map[string]interface{}
	chatSet             string                                   //当前登录用户
	memberList          []models.User                            //所有相关账号: 联系人, 公众号, 群组, 特殊账号
	groupMembers        map[string][]models.User                 //所有群组的成员, {'group_id1': [member1, member2, ...], ...}
	accountInfo         map[string]map[string]models.AccountInfo //所有账户, {'group_member':{'id':{'type':'group_member', 'info':{}}, ...}, 'normal_member':{'id':{}, ...}}
	contactList         []models.User                            // 联系人列表
	publicList          []models.User                            // 公众账号列表
	groupList           []models.User                            // 群聊列表
	specialList         []models.User                            // 特殊账号列表
	encryChatRoomIdList map[string]string                        // 存储群聊的EncryChatRoomId，获取群内成员头像时需要用到
	groupIdName         map[string]interface{}
	fileIndex           int
	send2oss            bool
	ossUrl              string
	handleMsgAll        HandleMsgAll
}

var (
	UNKONWN   = "unkonwn"
	SUCCESS   = "200"
	SCANED    = "201"
	TIMEOUT   = "408"
	ERRSYSTEM = "500"
)

var (
	WechatBot *WcBot
)

func New(handleMsgAll HandleMsgAll) *WcBot {
	wcBot := new(WcBot)
	wcBot.Debug = true
	wcBot.QrCodeTerminal = false
	wcBot.uuid = ""
	wcBot.baseUri = ""
	wcBot.baseHost = ""
	wcBot.redirectUri = ""
	wcBot.uin = ""
	wcBot.sid = ""
	wcBot.sKey = ""
	wcBot.passTicket = ""
	wcBot.deviceId = ""
	wcBot.Cookies = make([]*http.Cookie, 0)

	wcBot.baseRequest = make(map[string]interface{})
	wcBot.syncKeyStr = ""
	wcBot.syncHost = ""
	wcBot.status = "wait4login"
	wcBot.batchCount = 50
	wcBot.fullUserNameList = make([]string, 0)
	wcBot.wxIdList = make([]string, 0)
	wcBot.cursor = 0
	wcBot.isBigContact = false
	wcBot.tempPwd = config.Config.WxBot4gConf.WxQrDir
	wcBot.httpClient = httpClient.New(map[string]string{"User-Agent": "Mozilla/5.0 (X11; Linux i686; U;) Gecko/20070322 Kazehakase/0.4.5"})
	wcBot.conf = make(map[string]interface{})

	wcBot.chatSet = ""
	wcBot.myAccount = make(map[string]interface{})
	wcBot.memberList = make([]models.User, 0)
	wcBot.groupMembers = make(map[string][]models.User)

	wcBot.accountInfo = make(map[string]map[string]models.AccountInfo)
	wcBot.accountInfo["normal_member"] = make(map[string]models.AccountInfo)
	wcBot.accountInfo["group_member"] = make(map[string]models.AccountInfo)

	wcBot.contactList = make([]models.User, 0)
	wcBot.publicList = make([]models.User, 0)
	wcBot.groupList = make([]models.User, 0)
	wcBot.specialList = make([]models.User, 0)
	wcBot.encryChatRoomIdList = make(map[string]string)
	wcBot.groupIdName = make(map[string]interface{})
	wcBot.fileIndex = 0
	wcBot.send2oss = false
	wcBot.ossUrl = ""
	wcBot.handleMsgAll = handleMsgAll
	WechatBot = wcBot

	if _, err := os.Stat(wcBot.tempPwd); err != nil {
		if !os.IsExist(err) {
			if err := os.Mkdir(wcBot.tempPwd, os.ModePerm); err != nil {
				logrus.Error(err)
				return nil
			}
		}
	}

	return wcBot
}

func (wc *WcBot) QrCodeInTerminal() {
	wc.QrCodeTerminal = true
}

func initHttpServer() {
	g := gin.New()
	gin.SetMode(config.Config.ServerConf.Mode)

	g.Use(gin.Recovery())
	g.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route")
	})

	g.GET(config.Config.WxBot4gConf.HeartbeatURL, TextHandle)
	v1 := g.Group("/v1/msg")
	{
		v1.GET("/text", TextHandle)
		v1.POST("/image", ImageHandle)
	}

	go InitHeartbeatCron()

	logrus.Error(http.ListenAndServe(":"+strconv.Itoa(config.Config.ServerConf.Port), g).Error())
}

func (wc *WcBot) Run() {
	if err := wc.getUuid(); err != nil {
		logrus.Error(err.Error())
		return
	}

	if err := wc.genQrCode(path.Join(wc.tempPwd, "wxqr.png")); err != nil {
		logrus.Error(err.Error())
		return
	}

	if code := wc.wait4login(); code != SUCCESS {
		logrus.Error("web wechat login failed, failed code=", code)
		wc.status = "loginout"
		return
	}

	if ok := wc.login(); ok {
		logrus.Info("succeed: web wechat login")
	} else {
		logrus.Error("failed: web wechat login")
		wc.status = "loginout"
		return
	}

	if ok := wc.init(); ok {
		logrus.Info("succeed: web wechat init")
	} else {
		logrus.Info("failed: web wechat init")
	}

	if ok := wc.statusNotify(); ok {
		logrus.Info("succeed: web wechat status notify")
	} else {
		logrus.Info("failed: web wechat status notify")
	}

	if ok := wc.GetContact(false, ""); ok == "unknown" {
		logrus.Info(fmt.Sprintf("Get %d contacts", len(wc.contactList)))
		logrus.Info("succeed: start to process messages")
	}

	//监听 api 服务
	go initHttpServer()

	wc.procMsgLoop()

	wc.status = "loginout"
}

func (wc *WcBot) getUuid() error {
	urlStr := "https://login.weixin.qq.com/jslogin?"
	params := url.Values{
		"appid": []string{"wx782c26e4c19acffb"},
		"fun":   []string{"new"},
		"lang":  []string{"zh_CN"},
		"_":     []string{strconv.Itoa(int(time.Now().Unix())*1000 + rand.Intn(1000))},
	}
	data, err := wc.httpClient.Get(urlStr, params)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	regx := `window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+?)"`
	pm := utils.RegexpMatchStr(regx, string(data))
	if pm != nil && pm[0] != nil && len(pm[0]) >= 3 {
		code := pm[0][1]
		wc.uuid = pm[0][2]
		if code == SUCCESS {
			return nil
		} else {
			return errors.New(fmt.Sprintf("error code is : %s", code))
		}
	}
	return errors.New("regexp code uuid error")
}

func (wc *WcBot) genQrCode(filePath string) error {
	//wc.show_image(filePath)
	if wc.QrCodeTerminal {
		urlStr := "https://login.weixin.qq.com/qrcode/" + wc.uuid
		data, err := wc.httpClient.Get(urlStr, nil)
		if err != nil {
			logrus.Error(err)
			return err
		}
		M, err := qrcodetl.Decode(bytes.NewReader(data))
		if err != nil {
			logrus.Error(err)
			return err
		}
		qrterminal.GenerateHalfBlock(M.Content, qrterminal.M, os.Stdout)
	} else {
		urlStr := "https://login.weixin.qq.com/l/" + wc.uuid
		if err := qrcode.WriteFile(urlStr, qrcode.High, 256, filePath); err != nil {
			logrus.Error(err)
			return err
		}
	}

	logrus.Info("please use WeChat to scan the QR code")
	return nil
}

func (wc *WcBot) wait4login() string {
	/**
	http comet:
	tip=1, 等待用户扫描二维码,
		   201: scaned
		   408: timeout
	tip=0, 等待用户确认登录,
		   200: confirmed
	*/
	var (
		tip          = 1
		tryLaterSecs = 1
		maxRetryTime = 10
		code         = UNKONWN
		loginUrl     = "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login?tip=%d&uuid=%s&_=%s"
	)
	for retryTime := maxRetryTime; retryTime > 0; retryTime-- {
		urlStr := fmt.Sprintf(loginUrl, tip, wc.uuid, strconv.Itoa(int(time.Now().Unix())))

		code, data, err := wc.doRequest(urlStr)

		if err != nil {
			logrus.Error(err.Error())
			return ERRSYSTEM
		}

		switch code {
		case SCANED:
			logrus.Info("please confirm to login")
			tip = 0
		case TIMEOUT:
			logrus.Warnf(" WeChat login timeout. retry in %d secs later", tryLaterSecs)
			tip = 1
			retryTime--
			time.Sleep(time.Duration(tryLaterSecs))
		case SUCCESS:
			regx := `window.redirect_uri="(\S+?)";`
			param := utils.RegexpMatchStr(regx, string(data))
			if len(param) < 1 || len(param[0]) < 2 {
				err = errors.New("param less 1 param or param[0] less 2")
				return ERRSYSTEM
			}
			wc.redirectUri = param[0][1] + `&fun=new&version=v2`
			wc.baseUri = wc.redirectUri[:strings.LastIndex(wc.redirectUri, "/")]
			tempHost := wc.baseUri[8:]
			wc.baseHost = tempHost[:strings.Index(tempHost, "/")]
			return code
		default:
			logrus.Warnf("WeChat login exception return_code=%s. retry in %d secs later", code, tryLaterSecs)
			tip = 1
			retryTime--
			time.Sleep(time.Duration(tryLaterSecs))
		}
	}
	return code
}

func (wc *WcBot) login() bool {
	if len(wc.redirectUri) < 4 {
		logrus.Error("Login failed due to network problem, please try again")
		return false
	}

	data, err := wc.httpClient.Get(wc.redirectUri, nil)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(string(data)); err != nil {
		panic(err)
	}

	root := doc.SelectElement("error")
	if root == nil {
		return false
	}

	wc.sKey = root.SelectElement("skey").Text()
	wc.sid = root.SelectElement("wxsid").Text()
	wc.uin = root.SelectElement("wxuin").Text()
	wc.passTicket = root.SelectElement("pass_ticket").Text()

	if wc.sKey == "" || wc.sid == "" || wc.uin == "" || wc.passTicket == "" {
		return false
	}

	wc.baseRequest["Uin"] = wc.uin
	wc.baseRequest["Sid"] = wc.sid
	wc.baseRequest["Skey"] = wc.sKey
	wc.baseRequest["DeviceID"] = wc.deviceId

	wc.Cookies = wc.httpClient.GetCookie()

	return true

}

func (wc *WcBot) init() bool {
	var (
		wxMsgs = models.RecvMsgs{}
	)

	urlStr := wc.baseUri + fmt.Sprintf("/webwxinit?r=%d&lang=en_US&pass_ticket=%s", int(time.Now().Unix()), wc.passTicket)

	body := struct {
		BaseRequest interface{} `json:"BaseRequest"`
	}{
		BaseRequest: wc.baseRequest,
	}

	data, err := wc.httpClient.Post(urlStr, body)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	mData := make(map[string]interface{})
	if err := json.Unmarshal(data, &mData); err != nil {
		logrus.Error(err.Error())
		return false
	}

	err = json.Unmarshal(data, &wxMsgs)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	for _, item := range mData["ContactList"].([]interface{}) {
		if mItem, ok := item.(map[string]interface{}); ok {
			if mItem["UserName"].(string)[0:2] == "@@" {
				wc.groupIdName[mItem["UserName"].(string)] = mItem["NickName"].(string)
			}
		}
	}

	wc.syncKey = wxMsgs.SyncKeys
	wc.syncKeyStr = wxMsgs.SyncKeys.ToString()

	wc.myAccount = mData["User"].(map[string]interface{})
	wc.chatSet = mData["ChatSet"].(string)

	mmData := struct {
		Ret int `json:"Ret"`
	}{}
	if err := json.Unmarshal(data, &mmData); err != nil {
		logrus.Error(err.Error())
		return false
	}

	ret := mmData.Ret == 0
	return ret
}

func (wc *WcBot) statusNotify() bool {
	urlStr := wc.baseUri + fmt.Sprintf("/webwxstatusnotify?lang=zh_CN&pass_ticket=%s", wc.passTicket)

	wc.baseRequest["Uin"], _ = strconv.Atoi(wc.baseRequest["Uin"].(string))

	body := struct {
		BaseRequest  interface{} `json:"BaseRequest"`
		Code         int         `json:"Code"`
		FromUserName string      `json:"FromUserName"`
		ToUserName   string      `json:"ToUserName"`
		ClientMsgId  int         `json:"ClientMsgId"`
	}{
		BaseRequest:  wc.baseRequest,
		Code:         3,
		FromUserName: wc.myAccount["UserName"].(string),
		ToUserName:   wc.myAccount["UserName"].(string),
		ClientMsgId:  int(time.Now().Unix()),
	}

	data, err := wc.httpClient.Post(urlStr, body)
	if err != nil {
		logrus.Error(err.Error())
		return false
	}

	mData := make(map[string]interface{})
	if err := json.Unmarshal(data, &mData); err != nil {
		logrus.Error(err.Error())
		return false
	}

	mmData := struct {
		Ret int `json:"Ret"`
	}{}
	if err := json.Unmarshal(data, &mmData); err != nil {
		logrus.Error(err.Error())
		return false
	}

	ret := mmData.Ret == 0
	return ret
}

func (wc *WcBot) GetContact(isUnknow bool, uId string) string {
	contactMap := make(map[string]models.User, 0)
	urlStr := wc.baseUri + fmt.Sprintf("/webwxgetcontact?lang=zh_CN&seq=%s&pass_ticket=%s&skey=%s&r=%s",
		"0", wc.passTicket, wc.sKey, strconv.Itoa(int(time.Now().Unix())))

	//如果通讯录联系人过多，这里会直接获取失败
	data, err := wc.httpClient.Post(urlStr, nil)
	if err != nil {
		logrus.Error(err.Error())
		return ""
	}

	var contactList models.ContactList
	err = json.Unmarshal(data, &contactList)
	if err != nil {
		logrus.Error(err)
		return ""
	}

	for i := 0; i < contactList.MemberCount; i++ {
		contactMap[contactList.MemberList[i].UserName] = contactList.MemberList[i]
	}

	for contactList.Seq != 0 {
		logrus.Info(fmt.Sprintf("Geting contacts. Get %d contacts for now", contactList.MemberCount))

		urlStr := wc.baseUri + fmt.Sprintf("/webwxgetcontact?seq=%s&pass_ticket=%s&skey=%s&r=%d",
			strconv.Itoa(contactList.Seq), wc.passTicket, wc.sKey, int(time.Now().Unix()))
		data, err := wc.httpClient.Post(urlStr, nil)
		if err != nil {
			logrus.Error(err.Error())
			return ""
		}

		var contactList models.ContactList
		err = json.Unmarshal(data, &contactList)
		if err != nil {
			logrus.Error(err)
			return ""
		}

		for i := 0; i < contactList.MemberCount; i++ {
			contactMap[contactList.MemberList[i].UserName] = contactList.MemberList[i]
		}
	}

	wc.memberList = append(wc.memberList, contactList.MemberList...)

	specialUsers := map[string]bool{
		"newsapp": true, "fmessage": true, "filehelper": true, "weibo": true, "qqmail": true,
		"qmessage": true, "qqsync": true, "floatbottle": true,
		"lbsapp": true, "medianote": true, "qqfriend": true, "readerapp": true,
		"blogapp": true, "facebookapp:true": true, "masssendapp": true, "meishiapp": true,
		"feedsapp": true, "voip:true": true, "blogappweixin": true, "weixin": true, "brandsessionholder": true,
		"weixinreminder": true, "officialaccounts": true, "wxid_novlwrv3lqwv11": true,
		"gh_22b87fa7cb3c": true, "wxitil": true, "userexperience_alarm": true, "notification_messages": true,
	}

	if len(wc.memberList) <= 0 {
		return ""
	}

	for _, user := range wc.memberList {
		if user.VerifyFlag&8 != 0 {
			// 公众号
			wc.publicList = append(wc.publicList, user)
			wc.accountInfo["normal_member"][user.UserName] = models.AccountInfo{Type: "public", User: user}
		} else if _, ok := specialUsers[user.UserName]; ok {
			// 特殊账户
			wc.accountInfo["normal_member"][user.UserName] = models.AccountInfo{Type: "special", User: user}
		} else if strings.Contains(user.UserName, "@@") {
			// 群聊
			wc.groupList = append(wc.groupList, user)
			wc.accountInfo["normal_member"][user.UserName] = models.AccountInfo{Type: "group", User: user}
		} else if user.UserName == wc.myAccount["UserName"].(string) {
			// 自己
			wc.accountInfo["normal_member"][user.UserName] = models.AccountInfo{Type: "self", User: user}
		} else {
			wc.contactList = append(wc.contactList, user)
			wc.accountInfo["normal_member"][user.UserName] = models.AccountInfo{Type: "contact", User: user}
		}
	}

	if err := wc.batchGetGroupMembers(); err != nil {
		logrus.Error(err)
		return ""
	}

	if wc.Debug {
		if err = utils.WriteFile(wc.tempPwd+"groupList.json", wc.groupList); err != nil {
			logrus.Error(err)
			return ""
		}

		if err = utils.WriteFile(wc.tempPwd+"accountInfo.json", wc.accountInfo); err != nil {
			logrus.Error(err)
			return ""
		}
	}

	for _, groups := range wc.groupMembers {
		for _, group := range groups {
			if _, ok := wc.accountInfo["normal_member"][group.UserName]; !ok {
				wc.accountInfo["group_member"][group.UserName] = models.AccountInfo{Type: "contact", User: group, Group: group}

				//暂时不在此获取昵称，请调用GetGroupUserName
				//if isUnknow && uId != "" {
				//	if uId == group.UserName {
				//		return group.UserName
				//	} else if uId == group.DisplayName {
				//		return group.DisplayName
				//	} else if uId == group.NickName {
				//		return group.NickName
				//	} else {
				//		return "unknown"
				//	}
				//}
			}
		}
	}

	return "unknown"
}

func (wc *WcBot) procMsgLoop() {
	wc.testSyncCheck()
	wc.status = "loginsuccess" //WxbotManage使用
	for {
		retCode, selector, err := wc.syncCheck()
		logrus.Debug(retCode, " ", selector)
		if err != nil {
			logrus.Error(err)
		}
		switch retCode {
		case "1100":
			//从微信客户端上登出
		case "1101":
			//从其它设备上登了网页微信
		case "0":
			//msg="微信正常"
			switch selector {
			case "2":
				//有新消息
				if r, err := wc.sync(); err == nil {
					wc.handleMsg(r)
				} else {
					logrus.Error(err)
				}
			case "3":
				//未知
				if r, err := wc.sync(); err == nil {
					wc.handleMsg(r)
				}
			case "4":
				//通讯录更新
				if r, err := wc.sync(); err == nil {
					wc.handleMsg(r)
				}
			case "6":
				//可能是红包
				if r, err := wc.sync(); err == nil {
					wc.handleMsg(r)
				}
			case "7":
				//在手机上操作了微信
				if r, err := wc.sync(); err == nil {
					wc.handleMsg(r)
				}
			case "0":
				//无事件
			}
		default:
			logrus.Errorf("sync_check, retcode:%s selector:%s", retCode, selector)
		}
		wc.Schedule()

		time.Sleep(time.Second)
	}
}

func (wc *WcBot) Schedule() {
	/**
		做任务型事情的函数，如果需要，可以在子类中覆盖此函数
	 	此函数在处理消息的间隙被调用，请不要长时间阻塞此函数
	*/
}

func (wc *WcBot) doRequest(url string) (code string, data []byte, err error) {
	data, err = wc.httpClient.Get(url, nil)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	regx := `window.code=(\d+);`
	codes := utils.RegexpMatchStr(regx, string(data))
	if len(codes) < 1 || len(codes[0]) < 2 {
		err = errors.New("codes less 1 param or codes[0] less 2")
		return
	}
	code = codes[0][1]
	return
}

/**
{
    "BaseResponse":{
        "Ret":0,
        "ErrMsg":""
    },
    "Count":10,
    "ContactList":[
        {
            "Uin":0,
            "UserName":"@@40bccd2526c469d875a325076c1afefc35b1f0a677aa6f266a019ff8d4cd1aae",
            "NickName":"吃货群",
            "HeadImgUrl":"/cgi-bin/mmwebwx-bin/webwxgetheadimg?seq=657825175&username=@@40bccd2526c469d875a325076c1afefc35b1f0a677aa6f266a019ff8d4cd1aae&skey=",
            "ContactFlag":3,
            "MemberCount":6,
            "MemberList":[
                {
                    "Uin":0,
                    "UserName":"@2c301cc8ad2d753b22cac512b13de1be",
                    "NickName":"阿花 <span class="emoji emoji1f33b"></span>",
                    "AttrStatus":33784319,
                    "PYInitial":"",
                    "PYQuanPin":"",
                    "RemarkPYInitial":"",
                    "RemarkPYQuanPin":"",
                    "MemberStatus":0,
                    "DisplayName":"",
                    "KeyWord":"blu"
                },
                {
                    "Uin":0,
                    "UserName":"@a42ee05b2f48f05ad8e5caff36c72972",
                    "NickName":"子杰",
                    "AttrStatus":242279,
                    "PYInitial":"",
                    "PYQuanPin":"",
                    "RemarkPYInitial":"",
                    "RemarkPYQuanPin":"",
                    "MemberStatus":0,
                    "DisplayName":"",
                    "KeyWord":"jzz"
                },
			//...
		}
		//...
	]
}
*/
func (wc *WcBot) batchGetGroupMembers() error {
	urlStr := wc.baseUri + fmt.Sprintf("/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s",
		strconv.Itoa(int(time.Now().Unix())), wc.passTicket)

	body := struct {
		BaseRequest interface{}   `json:"BaseRequest"`
		Count       interface{}   `json:"Count"`
		List        []interface{} `json:"List"`
	}{
		BaseRequest: wc.baseRequest,
		Count:       len(wc.groupList),
	}

	for _, group := range wc.groupList {
		body.List = append(body.List, struct {
			UserName        string `json:"UserName"`
			EncryChatRoomId string `json:"EncryChatRoomId"`
		}{
			group.UserName,
			"",
		})
	}

	data, err := wc.httpClient.Post(urlStr, body)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	var groupList models.GroupList
	err = json.Unmarshal(data, &groupList)
	if err != nil {
		logrus.Error(err)
		return err
	}

	groupMembers := make(map[string][]models.User)
	encryChatRoomId := make(map[string]string)

	if wc.Debug {
		if err = utils.WriteFile(wc.tempPwd+"batchGetGroupMembers.json", data); err != nil {
			logrus.Error(err)
			return err
		}
	}

	for _, group := range groupList.ContactList {
		gid := group.UserName
		for _, member := range group.MemberList {
			groupMembers[gid] = append(groupMembers[gid], member)
			encryChatRoomId[gid] = member.EncryChatRoomId
		}
	}

	wc.groupMembers = groupMembers
	wc.encryChatRoomIdList = encryChatRoomId

	if wc.Debug {
		if err = utils.WriteFile(wc.tempPwd+"groupMembers.json", wc.groupMembers); err != nil {
			logrus.Error(err)
			return err
		}
	}

	return nil
}

func (wc *WcBot) testSyncCheck() bool {
	//host1 := []string{"webpush.", "webpush2."}
	host1 := []string{"webpush."}
	host2 := []string{"wx.qq.com", wc.baseHost}

	for _, h1 := range host1 {
		for _, h2 := range host2 {
			wc.syncHost = h1 + h2
			retCode, _, err := wc.syncCheck()
			if err != nil {
				retCode = "-1"
			}
			if retCode == "0" {
				return true
			}
		}
	}
	return false
}

func (wc *WcBot) syncCheck() (string, string, error) {
	tt := time.Now().UnixNano() / 1000000
	params := url.Values{
		"r":        []string{strconv.Itoa(int(tt))},
		"sid":      []string{wc.sid},
		"uin":      []string{wc.uin},
		"skey":     []string{wc.sKey},
		"deviceid": []string{wc.deviceId},
		"synckey":  []string{wc.syncKeyStr},
		"_":        []string{strconv.Itoa(int(tt))},
	}

	urlStr := "https://" + wc.syncHost + "/cgi-bin/mmwebwx-bin/synccheck?"

	wc.httpClient.SetCookie(wc.Cookies)

	data, err := wc.httpClient.Get(urlStr, params)
	if err != nil {
		logrus.Error(err.Error())
		return "-1", "-1", err
	}

	regx := `window.synccheck=\{retcode:"(\d+)",selector:"(\d+)"\}`
	pm := utils.RegexpMatchStr(regx, string(data))
	if pm != nil && pm[0] != nil && len(pm[0]) >= 3 {
		retCode := pm[0][1]
		selector := pm[0][2]

		return retCode, selector, nil
	}
	return "-1", "-1", errors.New("regexp error")
}

func (wc *WcBot) sync() (models.RecvMsgs, error) {
	var (
		wxMsges = models.RecvMsgs{}
	)
	urlStr := wc.baseUri + fmt.Sprintf("/webwxsync?sid=%s&skey=%s&lang=en_US&pass_ticket=%s",
		wc.sid, wc.sKey, wc.passTicket)

	body := struct {
		BaseRequest interface{} `json:"BaseRequest"`
		SyncKey     interface{} `json:"SyncKey"`
		RR          int         `json:"rr"`
	}{
		BaseRequest: wc.baseRequest,
		SyncKey:     wc.syncKey,
		RR:          int(time.Now().UnixNano()),
	}

	wc.httpClient.SetCookie(wc.Cookies)

	data, err := wc.httpClient.Post(urlStr, body)
	if err != nil {
		logrus.Error(err.Error())
		return wxMsges, err
	}

	err = json.Unmarshal(data, &wxMsges)
	if err != nil {
		return wxMsges, err
	}

	wc.syncKey = wxMsges.SyncKeys
	wc.syncKeyStr = wxMsges.SyncKeys.ToString()

	if wxMsges.ModContactCount == 1 && len(wxMsges.ModContactList) > 0 {
		groupName := wxMsges.ModContactList[0].(map[string]interface{})["NickName"].(string)
		if groupName != "" {
			groupId := wxMsges.ModContactList[0].(map[string]interface{})["UserName"].(string)
			wc.groupIdName[groupId] = groupName
		}
	}

	return wxMsges, nil
}

func (wc *WcBot) GetUserId(name string) string {
	if name == "" {
		return ""
	}

	for _, contact := range wc.contactList {
		if contact.RemarkName != "" && name == contact.RemarkName {
			return contact.UserName
		}

		if contact.DisplayName != "" && name == contact.DisplayName {
			return contact.UserName
		}

		if contact.NickName != "" && name == contact.NickName {
			return contact.UserName
		}
	}

	for _, group := range wc.groupList {
		if group.RemarkName != "" && name == group.RemarkName {
			return group.UserName
		}

		if group.DisplayName != "" && name == group.DisplayName {
			return group.UserName
		}

		if group.NickName != "" && name == group.NickName {
			return group.UserName
		}
	}

	for gid, gName := range wc.groupIdName {
		if gName == name {
			return gid
		}
	}

	return ""
}

/**
content_type_id:
	0 -> Text
	1 -> Location
	3 -> Image
	4 -> Voice
	5 -> Recommend
	6 -> Animation
	7 -> Share
	8 -> Video
	9 -> VideoCall
	10 -> Redraw
	11 -> Empty
	99 -> Unknown
msg_type_id: 消息类型id
msg: 消息结构体
return: 解析的消息
*/
func (wc *WcBot) extractMsgContent(msgTypeId int, msg models.RecvMsg) models.Content {
	mType := msg.MsgType
	content := html.UnescapeString(msg.Content)
	msgId := msg.MsgId

	var msgContent models.Content
	if msgTypeId == 0 {
		msgContent.Type = 11
		msgContent.Data = ""
		return msgContent
	} else if msgTypeId == 2 {
		//File Helper
		msgContent.Type = 0
		msgContent.Data = strings.Replace(content, `<br/>`, "\n", -1)
		return msgContent
	} else if msgTypeId == 3 {
		//群聊
		sp := strings.Index(content, `<br/>`)
		uId := content[:sp]
		content = content[sp:]
		content = strings.Replace(content, `<br/>`, "", -1)
		uId = uId[:(len(uId) - 1)]
		name := wc.getContactPreferName(wc.getContactName(uId))
		if name == "" {
			name = wc.getGroupMemberPreferName(wc.getGroupMemberName(msg.FromUserName, uId))
		}
		if name == "" {
			name = "unknown"
		}
		msgContent.User = models.ContentUser{Uid: uId, Name: name}
	} else {
		// Self, Contact, Special, Public, Unknown
		//pass
	}

	msgPrefix := ""
	if msgContent.User.Name != "" {
		msgPrefix = msgContent.User.Name
	}

	if mType == 1 {
		if strings.Contains(content, `http: //weixin.qq.com/cgi-bin/redirectforward?args=`) {
			data, err := wc.httpClient.Get(content, nil)
			if err != nil {
				logrus.Error(err)
			}
			pos := wc.searchContent("title", string(data), "xml")
			msgContent.Type = 1
			msgContent.Data = pos
			msgContent.Detail = models.Detail{Type: "str", Value: string(data)}
		} else {
			msgContent.Type = 0
			if msgTypeId == 3 || (msgTypeId == 1 && msg.ToUserName[:2] == "@@") {
				msgContent.Data, msgContent.Desc, msgContent.Other = wc.procAtInfo(content)
			} else {
				msgContent.Data = content
			}
		}
	} else if mType == 3 {
		//发送图片
		msgContent.Type = 3
		msgContent.Data = wc.getMsgImgUrl(msgId)
		data, err := wc.httpClient.Get(msgContent.Data, nil)
		if err != nil {
			logrus.Error(err)
		}

		maxEnLen := hex.EncodedLen(len(data)) // 最大编码长度
		dst1 := make([]byte, maxEnLen)
		hex.Encode(dst1, data)

		msgContent.Img = make([]byte, 0)
		msgContent.Img = append(msgContent.Img, dst1...)

		//TODO 发送照片到阿里云oss
		if wc.send2oss {
			wc.ossUrl = wc.sendMsgImgAliyun(msgId, wc.uin)
		}
	} else if mType == 34 {
		//发送语音
		msgContent.Type = 4
		msgContent.Data = wc.getVoiceUrl(msgId)

		data, err := wc.httpClient.Get(msgContent.Data, nil)
		if err != nil {
			logrus.Error(err)
		}

		maxEnLen := hex.EncodedLen(len(data)) // 最大编码长度
		dst1 := make([]byte, maxEnLen)
		hex.Encode(dst1, data)

		msgContent.Img = make([]byte, 0)
		msgContent.Voice = append(msgContent.Img, dst1...)
	} else if mType == 37 {
		// TODO 添加好友
		msgContent.Type = 37
		msgContent.Other = msg.RecommendInfo
	} else if mType == 42 {
		msgContent.Type = 5
		info := msg.RecommendInfo

		allSex := map[int]interface{}{0: "unknown", 1: "male", 2: "female"}

		msgContent.Other = map[string]interface{}{
			"nickname": info.(map[string]interface{})["NickName"],
			"alias":    info.(map[string]interface{})["Alias"],
			"province": info.(map[string]interface{})["Province"],
			"city":     info.(map[string]interface{})["City"],
			"gender":   allSex[info.(map[string]interface{})["Sex"].(int)]}
	} else if mType == 47 {
		msgContent.Type = 6
		msgContent.Data = wc.searchContent("cdnurl", content, "attr")
		if wc.Debug {
			logrus.Infof("%s[Animation] %s", msgPrefix, msgContent.Data)
		}
	} else if mType == 49 {
		var appMsgType string
		msgContent.Type = 7
		if msg.AppMsgType == 3 {
			appMsgType = "music"
		} else if msg.AppMsgType == 5 {
			appMsgType = "link"
		} else if msg.AppMsgType == 7 {
			appMsgType = "weibo"
		} else {
			appMsgType = "unknown"
		}
		msgContent.Other = map[string]interface{}{
			"type":    appMsgType,
			"title":   msg.FileName,
			"desc":    wc.searchContent("des", content, "xml"),
			"url":     msg.Url,
			"from":    wc.searchContent("appname", content, "xml"),
			"content": msg.Content, //有的公众号会发一次性3 4条链接一个大图,如果只url那只能获取第一条,content里面有所有的链接
		}
	} else if mType == 62 {
		msgContent.Type = 8
		msgContent.Data = content
		if wc.Debug {
			logrus.Infof("%s[Video] Please check on mobiles", msgPrefix)
		}
	} else if mType == 53 {
		msgContent.Type = 9
		msgContent.Data = content
		if wc.Debug {
			logrus.Infof("%s[Video Call]", msgPrefix)
		}
	} else if mType == 10002 {
		msgContent.Type = 10
		msgContent.Data = content
		if wc.Debug {
			logrus.Infof("%s[Redraw]", msgPrefix)
		}
	} else if mType == 10000 {
		msgContent.Type = 12
		msgContent.Data = msg.Content
		if wc.Debug {
			logrus.Info("[Unknown]")
		}
	} else if mType == 43 {
		msgContent.Type = 13
		msgContent.Data = wc.getVideoUrl(msgId)
		if wc.Debug {
			logrus.Infof("%s[video] %s", msgPrefix, msgContent.Data)
		}
	} else {
		msgContent.Type = 99
		msgContent.Data = content
		if wc.Debug {
			logrus.Warnf("[Unknown] msg content type:%d", 99)
		}
	}

	return msgContent
}

/**
处理原始微信消息的内部函数
	msg_type_id:
		0 -> Init			//初始化消息，内部数据
		1 -> Self			//自己发送的消息
		2 -> FileHelper		//文件消息
		3 -> Group			//群消息
		4 -> Contact		//联系人消息
		5 -> Public			//公众号消息
		6 -> Special		//特殊账号消息
		51 -> 获取wxid		//获取wxid消息
		99 -> Unknown		//	未知账号消息
*/
func (wc *WcBot) handleMsg(data models.RecvMsgs) {
	//wc.handleMsgAll(data)
	for _, msg := range data.MsgList {
		msgUser := models.MsgUser{
			ID:   msg.FromUserName,
			Name: UNKONWN,
		}

		msgTypeId := 0

		if msg.MsgType == 51 && msg.StatusNotifyCode == 4 {
			//系统消息
			msgTypeId = 0
			msgUser.Name = "system"
			//获取所有联系人的username 和 wxid，但是会收到3次这个消息，只取第一次
			if wc.isBigContact && len(wc.fullUserNameList) == 0 {
				wc.fullUserNameList = strings.Split(msg.StatusNotifyUserName, ",")
				//wc.wxid_list = re.search(r"username&gt;(.*?)&lt;/username", msg.Content).group(1).split(",")
			}
		} else if msg.MsgType == 37 {
			//好友消息
			msgTypeId = 37
			msgUser.Name = wc.getContactPreferName(wc.getContactName(msgUser.ID))
		} else if msg.FromUserName == msg.ToUserName {
			//发给自己
		} else if msg.ToUserName == "filehelper" {
			//文件助手
			msgTypeId = 2
			msgUser.Name = "file_helper"
		} else if msg.FromUserName[:2] == "@@" {
			//群消息
			msgTypeId = 3
			msgUser.Name = wc.getContactPreferName(wc.getContactName(msgUser.ID))
		} else if wc.isContact(msg.FromUserName) {
			//Contact
			msgTypeId = 4
			msgUser.Name = wc.getContactPreferName(wc.getContactName(msgUser.ID))
		} else if wc.isPublic(msg.FromUserName) {
			//Public
			msgTypeId = 5
			msgUser.Name = wc.getContactPreferName(wc.getContactName(msgUser.ID))
		} else if wc.isSpecial(msg.FromUserName) {
			//Special
			msgTypeId = 6
			msgUser.Name = wc.getContactPreferName(wc.getContactName(msgUser.ID))
		} else {
			msgTypeId = 99
			msgUser.Name = UNKONWN
		}

		content := wc.extractMsgContent(msgTypeId, msg)
		realMsg := models.RealRecvMsg{
			MsgTypeId:    msgTypeId,
			MsgId:        msg.MsgId,
			FromUserName: msg.FromUserName,
			ToUserName:   msg.ToUserName,
			MsgType:      msg.MsgType,
			Content:      content,
			CreateTime:   msg.CreateTime,
			SendMsgUSer:  msgUser,
		}
		go wc.handleMsgAll(realMsg)
	}
}

package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"configcenter/src/common/blog"
)

type AlarmOptions struct {
	AlarmID            string    `json:"alarmID"`
	ConvergenceSeconds *uint16   `json:"convergenceSeconds"`
	AlarmKind          AlarmType `json:"alarmKind"`
	Receivers          string    `-`
	AlarmMsg           string    `json:"alarmMsg"`
	VoiceReadMsg       string    `json:"voiceReadMsg"`
}

func NewAlarmSet(cfg Config) (*AlarmerSet, error) {
	alarmSet := &AlarmerSet{
		Filter:       &AlarmFilter{AlarmIdSet: make(map[string]AlarmCfg)},
		DefaultRecvs: cfg.Receivers,
		Address:      cfg.Address,
		AppCode:      cfg.AppCode,
		AppSecret:    cfg.AppSecret,
		Operator:     cfg.Operator,
		BKTicket:     cfg.BKTicket,
		RtxAlarm:     newRtx,
		WeiXinAlarm:  newWebchat,
		VoiceAlarm:   newVoice,
	}
	alarmSet.Sync()
	return alarmSet, nil
}

type AlarmCfg struct {
	StartTime      time.Time
	TimeOutSeconds uint16
}

type AlarmFilter struct {
	locker     sync.RWMutex
	AlarmIdSet map[string]AlarmCfg
}

func (f *AlarmFilter) CanSend(alarmID string, convergenceSeconds *uint16) bool {
	if len(alarmID) == 0 {
		return true
	}
	f.locker.RLock()
	defer f.locker.RUnlock()
	conf, exist := f.AlarmIdSet[alarmID]
	if !exist || time.Since(conf.StartTime) > time.Duration(conf.TimeOutSeconds)*time.Second {
		var cfg AlarmCfg
		cfg.StartTime = time.Now()
		cfg.TimeOutSeconds = *convergenceSeconds
		f.AlarmIdSet[alarmID] = cfg
		return true
	}
	return false
}

type AlarmerSet struct {
	Filter       *AlarmFilter
	DefaultRecvs string
	Address      string
	AppCode      string
	AppSecret    string
	Operator     string
	BKTicket     string
	Biz          string
	RtxAlarm     func() *AlarmMessage
	WeiXinAlarm  func() *AlarmMessage
	VoiceAlarm   func() *AlarmMessage
}

func (a *AlarmerSet) Sync() {
	duration := time.Duration(500 * time.Millisecond)
	go func() {
		a.Filter.locker.Lock()
		for id, cfg := range a.Filter.AlarmIdSet {
			if time.Since(cfg.StartTime) > time.Duration(cfg.TimeOutSeconds)*time.Second {
				delete(a.Filter.AlarmIdSet, id)
			}
		}
		a.Filter.locker.Unlock()
		time.Sleep(duration)
	}()
}

func (a *AlarmerSet) SendAlarm(op *AlarmOptions) error {
	// do filter operation first.
	if false == op.AlarmKind.IsValid() {
		// if not set, use alarm kind.
		op.AlarmKind = INFO_ALARM
	}

	if true == op.AlarmKind.IsVoice() && len(op.VoiceReadMsg) == 0 {
		return errors.New("send voice alarm, but voice message is empty.")
	}

	// set default convergence second.
	if nil == op.ConvergenceSeconds {
		sec := uint16(60)
		op.ConvergenceSeconds = &sec
	}

	if false == a.Filter.CanSend(op.AlarmID, op.ConvergenceSeconds) {
		blog.Infof("received same alarm id[%s] request, drop it...", op.AlarmID)
		return nil
	}

	if err := a.DoSend(op); nil != err {
		return err
	}

	return nil
}
func (a *AlarmerSet) DoSend(op *AlarmOptions) error {
	var errs []error
	var recvs string
	recvs = a.DefaultRecvs
	if len(op.Receivers) > 0 {
		recvs = op.Receivers
	}

	if op.AlarmKind.IsRtx() {
		if err := a.RtxAlarm().SetResourcePath(a.Address + "component/compapi/tof/send_rtx/").
			SetAppCode(a.AppCode).SetAppSecret(a.AppSecret).SetOperator(a.Operator).SetMessage(op.AlarmMsg).
			SetReceivers(recvs).Do(); nil != err {
			errs = append(errs, err)
		}
	}

	if op.AlarmKind.IsWeiXin() {
		if err := a.WeiXinAlarm().SetResourcePath(a.Address + "component/compapi/tof/send_wx/").
			SetAppCode(a.AppCode).SetAppSecret(a.AppSecret).SetOperator(a.Operator).SetMessage(op.AlarmMsg).
			SetReceivers(recvs).Do(); nil != err {
			errs = append(errs, err)
		}
	}

	if op.AlarmKind.IsVoice() {
		if err := a.VoiceAlarm().SetResourcePath(a.Address + "component/compapi/uwork/noc_notice/").
			SetAppCode(a.AppCode).SetAppSecret(a.AppSecret).SetOperator(a.Operator).SetBKTicket(a.BKTicket).
			SetNoticeInfo(op.AlarmMsg).SetUserList(recvs).SetReadMsg(op.VoiceReadMsg).Do(); nil != err {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}

type AlarmMessage struct {
	Metadata
	AppCode             string            `json:"app_code"`
	AppSecret           string            `json:"app_secret"`
	BKTicket            string            `json:"bk_ticket,omitempty"`
	Operator            string            `json:"operator"`
	Sender              string            `json:"sender,omitempty"`
	Receiver            string            `json:"receiver"`
	Title               string            `json:"title,omitempty"`
	Message             string            `json:"message,omitempty"`
	Priority            string            `json:"priority,omitempty"`
	IsIgnoreFormerStaff bool              `json:"is_ignore_former_staff,omitempty"`
	AutoReadMessage     string            `json:"auto_read_message,omitempty"`
	KeyOptions          map[string]string `json:"key_options,omitempty"`
	HeadDescribe        string            `json:"head_desc,omitempty"`
	DataList            []VoiceDataInfo   `json:"busi_data_list,omitempty"`
	UserList            []UserInfo        `json:"user_list_information,omitempty"`
	NoticeInfo          string            `json:"notice_information,omitempty"`
}

type VoiceDataInfo struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type UserInfo struct {
	UserName string `json:"username"`
	Phone    string `json:"mobile_phone"`
}

func (a *AlarmMessage) SetReceivers(receivers string) *AlarmMessage {
	a.Receiver = receivers
	return a
}

func (a *AlarmMessage) SetMessage(msg string) *AlarmMessage {
	a.Message = msg
	return a
}

func (a *AlarmMessage) SetResourcePath(resourcePath string) *AlarmMessage {
	a.Metadata.ResourcePath = resourcePath
	return a
}

func (a *AlarmMessage) SetAppCode(appCode string) *AlarmMessage {
	a.AppCode = appCode
	return a
}

func (a *AlarmMessage) SetAppSecret(appSecret string) *AlarmMessage {
	a.AppSecret = appSecret
	return a
}

func (a *AlarmMessage) SetOperator(operator string) *AlarmMessage {
	a.Operator = operator
	return a
}

func (a *AlarmMessage) SetBKTicket(bkTicket string) *AlarmMessage {
	a.BKTicket = bkTicket
	return a
}

func (a *AlarmMessage) SetUserList(users string) *AlarmMessage {
	a.UserList = make([]UserInfo, 0)
	for _, r := range strings.Split(users, ",") {
		a.UserList = append(a.UserList, UserInfo{UserName: r})
	}
	return a
}

func (a *AlarmMessage) SetNoticeInfo(msg string) *AlarmMessage {
	a.NoticeInfo = msg
	return a
}

func (a *AlarmMessage) SetReadMsg(msg string) *AlarmMessage {
	a.AutoReadMessage = msg
	return a
}

func (a AlarmMessage) Do() error {
	js, err := json.Marshal(a)
	if nil != err {
		return fmt.Errorf("marshal %s alarm failed. error: %v", a.Type, err)
	}

	if err := a.DoPost(js); nil != err {
		return fmt.Errorf("send %s alarm failed. error: %v", a.Type, err)
	}
	return nil
}

func newRtx() *AlarmMessage {
	return &AlarmMessage{
		Metadata: Metadata{
			Type: "rtx",
		},
		Sender:              "cmdb-alarm",
		Title:               "CMDB 组件上下线告警",
		Message:             "",
		Priority:            "",
		IsIgnoreFormerStaff: true,
	}
}

func newWebchat() *AlarmMessage {
	return &AlarmMessage{
		Metadata: Metadata{
			Type: "weixin",
		},
		Message:             "",
		Priority:            "",
		IsIgnoreFormerStaff: true,
	}
}

func newVoice() *AlarmMessage {
	return &AlarmMessage{
		Metadata: Metadata{
			Type: "voice",
		},
		HeadDescribe: "CMDB组件上下线告警",
		KeyOptions:   map[string]string{},
		DataList:     []VoiceDataInfo{},
	}
}

type AlarmResponse struct {
	Result  bool        `json:"result"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Metadata struct {
	ResourcePath string `json:"-"`
	Type         string `json:"-"`
}

func (m *Metadata) DoPost(data []byte) error {
	req, err := http.NewRequest("POST", m.ResourcePath, strings.NewReader(string(data)))
	if nil != err {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if nil != err {
		return fmt.Errorf("http post failed, type: %s . err: %v", m.Type, err)
	}
	result, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if nil != err {
		return fmt.Errorf("read request body failed. type: %s,  err: %v", m.Type, err)
	}
	alarmResp := new(AlarmResponse)
	if err := json.Unmarshal(result, alarmResp); nil != err {
		return fmt.Errorf("unamrshal body failed. type: %s, err: %v, body: %s", m.Type, err, string(result))
	}
	if alarmResp.Result == true {
		blog.Infof("send %s alarm message success. reply message: %#v\n", m.Type, *alarmResp)
		return nil
	} else {
		return fmt.Errorf("send %s alarm message failed. error code: %s, error message %s.", m.Type, alarmResp.Code, alarmResp.Message)
	}
}

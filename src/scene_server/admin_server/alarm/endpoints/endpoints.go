package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"configcenter/src/apimachinery/discovery"
	"configcenter/src/common/blog"
	"configcenter/src/common/types"
	alarmtype "configcenter/src/scene_server/admin_server/alarm/types"
	"configcenter/src/scene_server/admin_server/alarm/zk"
)

type AlarmType string

const (
	AddedAlarmType         AlarmType = "组件上线"
	LostDetectedAlarmType  AlarmType = "组件下线"
	LostConfirmedAlarmType AlarmType = "组件下线超过60秒"

	GrayTimeInSecond time.Duration = 5
)

type EndpointsAlarm struct {
	zkAddrs          []string
	startTime        time.Time
	sendTo           *alarmtype.AlarmerSet
	endpointTemplate *template.Template
	users            string
	zkWatcher        *zk.ZkWatcher
	discovery.ServiceManageInterface
}

func NewEndpointsAlarm(cfg alarmtype.Config, sendTo *alarmtype.AlarmerSet, serviceManage discovery.ServiceManageInterface, addr string) (*EndpointsAlarm, error) {
	t, err := template.New("endpoints").Parse(endpointEventTemplate)
	if nil != err {
		return nil, fmt.Errorf("new endpoint event template failed. err: %v", err)
	}
	endpointsAlarm := &EndpointsAlarm{
		startTime:              time.Now(),
		sendTo:                 sendTo,
		endpointTemplate:       t,
		users:                  cfg.Receivers,
		ServiceManageInterface: serviceManage,
	}
	addrs := []string{addr}
	watcher, err := zk.NewZkWatcher(addrs, types.CC_SERV_BASEPATH, endpointsAlarm, )
	if nil != err {
		return nil, err
	}
	endpointsAlarm.zkWatcher = watcher
	return endpointsAlarm, nil
}

func (r *EndpointsAlarm) Run(<-chan struct{}) error {
	if err := r.zkWatcher.Run(); nil != err {
		return err
	}
	return nil
}

func (r *EndpointsAlarm) SendAlarm(branch string, child string, alarmType AlarmType, data string) {
	if false == r.ServiceManageInterface.IsMaster() {
		blog.Warnf("received endpoints alarm, but drop it because of running in slave mode.")
		return
	}

	var srvInfo types.ServerInfo
	if err := json.Unmarshal([]byte(data), &srvInfo); nil != err {
		blog.Errorf("unmarshal leaf %s/%s failed. err: %v", branch, child, err)
		return
	}
	ver := srvInfo.Version[:strings.Index(srvInfo.Version,"Tag")]
	endpointAlarm := EndpointEvent{
		Type:       string(alarmType),
		Component:  branch[strings.LastIndex(branch, "/")+1:],
		Instance:   srvInfo.Instance(),
		Version:    ver,
		ReportTime: time.Now().Format(time.RFC3339Nano),
		Biz:        r.sendTo.Biz,
	}

	if time.Since(r.startTime) < time.Duration(GrayTimeInSecond*time.Second) {
		blog.Warnf("received alarm but now is in gray time, skip! path: %s", branch+"/"+child)
		return
	}

	w := bytes.Buffer{}
	if err := r.endpointTemplate.Execute(&w, endpointAlarm); nil != err {
		blog.Errorf("format endpoint alarm failed. alarm: %#v, err: %v", endpointAlarm, err)
		return
	}

	users := r.users

	options := &alarmtype.AlarmOptions{
		AlarmID: srvInfo.Instance() + string(alarmType),
		Receivers: users,
		AlarmMsg:  w.String(),
	}

	switch alarmType {
	case AddedAlarmType:
		options.AlarmKind = alarmtype.INFO_ALARM
	case LostDetectedAlarmType:
		options.AlarmKind = alarmtype.INFO_ALARM
	case LostConfirmedAlarmType:
		readVoice := fmt.Sprintf("配置平台组件下线告警, 组件：%s, IP地址: %s, 端口: %d", branch[strings.LastIndex(branch, "/")+1:], srvInfo.IP, srvInfo.Port)
		options.AlarmKind = alarmtype.VOICE_ALARM
		options.VoiceReadMsg = readVoice
	default:
		options.AlarmKind = alarmtype.INFO_ALARM
	}

	err := r.sendTo.SendAlarm(options)
	if nil != err {
		blog.Errorf("send endpoints alarm failed. err :%v; content: %s", err, w.String())
		return
	}
	blog.Info("send endpoints alarm success： %s.", w.String())
}

func (e *EndpointsAlarm) OnAddLeaf(branch, leaf, value string) {
	e.SendAlarm(branch, leaf, AddedAlarmType, value)
}

func (e *EndpointsAlarm) OnDeleteLeaf(branch, leaf, value string, errorAlarm bool) {
	if errorAlarm {
		e.SendAlarm(branch, leaf, LostConfirmedAlarmType, value)
	} else {
		e.SendAlarm(branch, leaf, LostDetectedAlarmType, value)
	}
}

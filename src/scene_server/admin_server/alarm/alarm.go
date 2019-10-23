package alarm

import (
	"configcenter/src/apimachinery/discovery"
	"configcenter/src/scene_server/admin_server/alarm/endpoints"
	"configcenter/src/scene_server/admin_server/alarm/types"
)

func NewAlarm(cfg types.Config, serviceManage discovery.ServiceManageInterface, addr string) (*Alarm, error) {
	alarmSet, err := types.NewAlarmSet(cfg)
	if nil != err {
		return nil, err
	}
	endpointAlarm, err := endpoints.NewEndpointsAlarm(cfg, alarmSet, serviceManage, addr)
	if nil != err {
		return nil, err
	}

	alert := &Alarm{
		Objects: map[string]AlarmObjectInterface{
			"endpoints": endpointAlarm,
		},
	}

	return alert, nil
}

type Alarm struct {
	Objects map[string]AlarmObjectInterface
}

type AlarmObjectInterface interface {
	Run(<-chan struct{}) error
}

func (a *Alarm) Run() error {
	neverStop := make(chan struct{})
	for _, obj := range a.Objects {
		if err := obj.Run(neverStop); nil != err {
			return err
		}
	}
	return nil
}

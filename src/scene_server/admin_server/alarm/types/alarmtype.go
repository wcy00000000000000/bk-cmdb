package types

type AlarmType int32

const (
	RTX_ALALRM    AlarmType = 1 << 2
	WEIXIN_ALALRM AlarmType = 1 << 3
	VOICE_ALARM   AlarmType = 1 << 5

	INFO_ALARM  = RTX_ALALRM | WEIXIN_ALALRM
)

func (a AlarmType) IsValid() bool {
	return int32(a) != 0
}

func (a AlarmType) IsRtx() bool {
	return (a & RTX_ALALRM) == RTX_ALALRM
}

func (a AlarmType) IsWeiXin() bool {
	return (a & WEIXIN_ALALRM) == WEIXIN_ALALRM
}

func (a AlarmType) IsVoice() bool {
	return (a & VOICE_ALARM) == VOICE_ALARM
}

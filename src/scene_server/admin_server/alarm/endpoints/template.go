package endpoints

type EndpointEvent struct {
	Type       string
	Component  string
	Instance   string
	Version    string
	ReportTime string
	Biz        string
}

const endpointEventTemplate string = `《配置平台》
通知类型: {{.Type}}
所属业务：{{.biz}}
告警对象: {{.Instance}}
告警内容: {{.Component}}组件上下线
告警时间: {{.ReportTime}}
`

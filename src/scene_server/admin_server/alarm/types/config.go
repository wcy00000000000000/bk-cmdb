package types

import (
	"errors"
	"strconv"
	"strings"
)

type Config struct {
	Enable    bool
	Receivers string
	Address   string
	AppCode   string
	AppSecret string
	Operator  string
	BKTicket  string
	Biz       string
}

func ParseConfigFromKV(prefix string, configmap map[string]string) (Config, error) {
	var err error
	var cfg Config
	enable, exist := configmap[prefix+".enable"]
	if !exist {
		return Config{}, nil
	}
	cfg.Enable, err = strconv.ParseBool(enable)
	if err != nil {
		return Config{}, errors.New(`invalid alarm "enable" value`)
	}
	if !cfg.Enable {
		return Config{}, nil
	}

	address, exist := configmap[prefix+".address"]
	if !exist {
		return cfg, errors.New(`missing "address" configuration for alarm`)
	}
	cfg.Address = address
	if !strings.HasSuffix(cfg.Address, "/") {
		cfg.Address = cfg.Address + "/"
	}

	cfg.AppSecret, exist = configmap[prefix+".appSecret"]
	if !exist {
		return cfg, errors.New(`missing "appSecret" configuration for alarm`)
	}
	if len(cfg.AppSecret) == 0 {
		return cfg, errors.New(`invalid "appSecret" configuration for alarm`)
	}

	cfg.AppCode, exist = configmap[prefix+".appCode"]
	if !exist {
		return cfg, errors.New(`missing "appCode" configuration for alarm`)
	}
	if len(cfg.AppCode) == 0 {
		return cfg, errors.New(`invalid "appCode" configuration for alarm`)
	}

	cfg.Receivers, exist = configmap[prefix+".receivers"]
	if !exist {
		return cfg, errors.New(`missing "receivers" configuration for alarm`)
	}
	if len(cfg.Receivers) == 0 {
		return cfg, errors.New(`invalid "receivers" configuration for alarm`)
	}

	cfg.Operator, exist = configmap[prefix+".operator"]
	if !exist {
		cfg.Operator = ""
	}

	cfg.BKTicket, exist = configmap[prefix+".bkTicket"]
	if !exist {
		cfg.BKTicket = ""
	}

	cfg.Biz, exist = configmap[prefix+".biz"]
	if !exist {
		cfg.Biz = ""
	}

	return cfg, nil
}

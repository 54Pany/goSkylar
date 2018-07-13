package lib

import (
	"github.com/go-ini/ini"
)

var Config = &ConfigUtil{}

type ConfigUtil struct {
	fileName string
}

func (this ConfigUtil) GetString(Section string, Key string) (string, error) {
	var (
		cfg, err = ini.InsensitiveLoad("conf/config.ini")
	)

	sec1, err := cfg.GetSection(Section)
	if err != nil {
		return "", err
	}
	keys, err := sec1.GetKey(Key)
	if err != nil {
		return "", err
	}
	return keys.String(), err
}

func NewConfigUtil(file string) (*ConfigUtil) {
	if file == "" {
		file = "config.ini"
	}
	return &ConfigUtil{
		fileName: file,
	}
}

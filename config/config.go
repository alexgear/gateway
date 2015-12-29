package config

import "github.com/BurntSushi/toml"

type config struct {
	CalendarId string
}

var err error

func New(configPath string) (config, error) {
	var conf config
	_, err = toml.DecodeFile(configPath, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

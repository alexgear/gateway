package config

import "github.com/BurntSushi/toml"

type config struct {
	CalendarId string
	ServerHost string
	ServerPort int
}

var err error
var cfg config

func Init(configPath string) error {
	_, err = toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return err
	}
	return nil
}

func GetConfig() *config {
	return &cfg
}

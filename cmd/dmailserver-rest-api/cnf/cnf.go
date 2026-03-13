package cnf

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

type ServerConfig struct {
	Port          int
	Host          string
	APIKey        string
	CommandPrefix string
	MAIL_HOST     string
}

type Config struct {
	ServerConfig ServerConfig
}

func GetConfig(configPath string) (Config, error) {
	c := Config{}
	// config.PconfigPath stringv: will parse env var in string value. eg: shell: ${SHELL}
	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)

	err := config.LoadFiles(configPath)
	if err != nil {
		return c, err
	}
	config.BindStruct("ServerConfig", &c.ServerConfig)
	return c, nil
}

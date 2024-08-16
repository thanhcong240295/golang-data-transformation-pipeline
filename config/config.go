package config

import (
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

type Config struct {
	Environment           string `mapstructure:"ENVIRONMENT"`
	AWS_REGION            string `mapstructure:"AWS_REGION"`
	AWS_SECRET_ACCESS_KEY string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	AWS_ACCESS_KEY_ID     string `mapstructure:"AWS_ACCESS_KEY_ID"`
	MYSQL_HOST            string `mapstructure:"MYSQL_HOST"`
	MYSQL_PORT            string `mapstructure:"MYSQL_PORT"`
	MYSQL_DATABASE        string `mapstructure:"MYSQL_DATABASE"`
	MYSQL_USERNAME        string `mapstructure:"MYSQL_USERNAME"`
	MYSQL_PASSWORD        string `mapstructure:"MYSQL_PASSWORD"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}

	err = viper.Unmarshal(&config)
	return
}

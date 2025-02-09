package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Proto     string
	Host      string
	Port      int
	ConnCount int
}

var cfg *RedisConfig

func initConfig() {

	configDir := getConfigDir()
	viper.SetConfigFile(fmt.Sprintf("%s/default.json", configDir))
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Using default config file: ", viper.ConfigFileUsed())
	} else {
		fmt.Println("Error parsing default config file: ", viper.ConfigFileUsed())
		panic(err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
}

func GetConfig() *RedisConfig {
	if cfg == nil {
		initConfig()
	}
	return cfg
}

func getConfigDir() string {
	devConfigPath := os.Getenv("GOPATH")
	relativePath := "mockredis/config"
	absolutePath := path.Join(devConfigPath, relativePath)
	return absolutePath
}

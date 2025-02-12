package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type RedisConfig struct {
	Protocol    string `mapstructure:"protocol"`
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Connections int    `mapstructure:"connections"`
}

func initConfig() *RedisConfig {

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error parsing default config file: ", viper.ConfigFileUsed())
		fmt.Println(err.Error())
		panic(err)
	}
	fmt.Println("default config file: ", viper.ConfigFileUsed())

	var cfg *RedisConfig

	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Println("unable to decode into struct: ", err.Error())
		panic(err)
	}

	return cfg
}

func GetConfig() *RedisConfig {
	return initConfig()
}

package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Service struct {
	Service string `mapstructure:"service"`
	URL     string `mapstructrue:url`
}

type Services struct {
	Services []Service `mapstructure:"services"`
}

func Load() Services {
	v := viper.New()
	v.SetConfigName("configs")
	v.SetConfigType("yml")
	v.AddConfigPath("./configmap/")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var c Services
	if err := v.Unmarshal(&c); err != nil {
		log.Fatalf("Error unmarshalung config %v", err)
	}
	return c
}

package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

var config struct {
	Addr          string
	Port          string
	MongoUri      string
	MongoUser     string
	MongoPass     string
	MongoDatabase string
}

func readConfig() {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/fakescan/")
	v.AddConfigPath(".")

	v.SetDefault("addr", "127.0.0.1")
	v.SetDefault("port", "9999")
	v.SetDefault("mongoUri", "mongo://localhost:27017")
	v.SetDefault("mongoDatabase", "fakeScan")

	if err := v.ReadInConfig(); err != nil {
		log.Panic(fmt.Sprintf("Read config file failed: %s \n", err))
	}

	if err := v.Unmarshal(&config); err != nil {
		log.Panic(fmt.Sprintf("Unmarshal config failed: %s \n", err))
	}

	log.Println("success load config")
}

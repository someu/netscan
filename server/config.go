package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

//
//type Config struct {
//	ListenPort int
//	MongoUrl   string
//	MongoUser  string
//	MongoPass  string
//}

func readConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/fakescan/")
	viper.AddConfigPath(".")

	viper.SetDefault("ListenPort", "9999")
	viper.SetDefault("MongoUrl", "mongo://localhost:27017")
	viper.SetDefault("MongoDatabase", "fakescan")

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		log.Panic(fmt.Errorf("Read config file failed: %s \n", err))
	}
	log.Println("success load config")
}

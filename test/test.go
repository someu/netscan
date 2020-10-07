package main

import (
	"fmt"
	"github.com/spf13/viper"
)

//定义config结构体
var C struct {
	Addr          string
	Port          string
	MongoUri      string
	MongoUser     string
	MongoPass     string
	MongoDatabase string
}

func main() {
	config := viper.New()
	config.AddConfigPath(".")
	config.SetConfigName("config")
	config.SetConfigType("json")
	if err := config.ReadInConfig(); err != nil {
		panic(err)
	}

	//直接反序列化为Struct
	//var configjson Config
	if err := config.Unmarshal(&C); err != nil {
		fmt.Println(err)
	}

	fmt.Println(C)
}

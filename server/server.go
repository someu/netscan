package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func init() {
	readConfig()
	connectMongo()
}

func main() {
	r := gin.Default()
	r.GET("/api/asset", getAssetList)
	r.POST("/api/scan", createScan)
	r.GET("/api/scan", getScanList)

	r.Run(fmt.Sprintf("%s:%s", config.Addr, config.Port))
}

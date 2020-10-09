package main

import (
	"context"
	"fakescan/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"strings"
	"time"
)

type CreateScanRequestData struct {
	Name string `form:"name"`
	IP   string `form:"ip"`
	Port string `form:"port"`
}

func getAssetList(c *gin.Context) {
	result := struct {
		List  []Asset
		Total int64
	}{}
	current, _ := strconv.ParseInt(c.DefaultQuery("current", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
	search := c.DefaultQuery("search", "")

	findOptions := options.Find().SetSort(bson.D{{"CreatedAt", -1}}).SetLimit(pageSize).SetSkip((current - 1) * pageSize)

	filter := bson.M{
		"Address":      bson.M{"$regex": search},
		"Fingers.Name": bson.M{"$regex": search},
	}

	var err error
	cursor, err := assetCollection.Find(context.TODO(), filter, findOptions)

	if err != nil {
		log.Printf("find scan fained %s", err)
		c.JSON(500, result)
		return
	}

	err = cursor.All(context.TODO(), &result.List)

	if err != nil {
		log.Printf("all scan fained %s", err)
		c.JSON(500, result)
		return
	}

	result.Total, err = assetCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Printf("get scans count %s", err)
		c.JSON(500, result)
		return
	}

	c.JSON(200, result)
}

func getScanList(c *gin.Context) {
	result := struct {
		List  []Scan
		Total int64
	}{}
	current, _ := strconv.ParseInt(c.DefaultQuery("current", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("pageSize", "10"), 10, 64)
	search := c.DefaultQuery("search", "")

	findOptions := options.Find().SetSort(bson.D{{"CreatedAt", -1}}).SetLimit(pageSize).SetSkip((current - 1) * pageSize)

	filter := bson.M{
		"Target": bson.M{"$regex": search},
	}

	cursor, err := scanCollection.Find(context.TODO(), filter, findOptions)

	if err != nil {
		log.Printf("find scan fained %s", err)
		c.JSON(500, result)
		return
	}

	err = cursor.All(context.TODO(), &result.List)

	if err != nil {
		log.Printf("all scan fained %s", err)
		c.JSON(500, result)
		return
	}

	result.Total, err = scanCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Printf("all scan fained %s", err)
		c.JSON(500, result)
		return
	}
	log.Println(util.Stringify(result))
	c.JSON(200, result)
}

func createScan(c *gin.Context) {
	var reqData CreateScanRequestData
	if err := c.ShouldBind(&reqData); err != nil {
		c.JSON(400, "请求错误"+err.Error())
		return
	}

	if len(reqData.Name) == 0 || len(reqData.IP) == 0 || len(reqData.Port) == 0 {
		c.JSON(500, "请求数据错误")
		return
	}

	scan := Scan{
		Name:     reqData.Name,
		Ip:       strings.Split(reqData.IP, "\n"),
		Port:     strings.Split(reqData.Port, ","),
		StartAt:  time.Now(),
		FinishAt: time.Time{},
	}
	s, err := scanCollection.InsertOne(context.TODO(), scan)

	if err != nil {
		c.JSON(500, "生成任务错误"+err.Error())
		return
	}

	scheduler.CreateScan(scan)

	c.JSON(200, s)
}

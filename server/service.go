package main

import (
	"context"
	"fakescan/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"strconv"
	"time"
)

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
	target := c.PostFormArray("target")

	if target == nil || len(target) == 0 {
		c.JSON(500, nil)
		return
	}

	scan := Scan{
		Target:  target,
		StartAt: time.Now(),
	}
	s, err := scanCollection.InsertOne(context.TODO(), scan)

	if err != nil {
		c.JSON(500, nil)
		return
	}

	c.JSON(200, s)
}

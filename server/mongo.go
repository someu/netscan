package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var mongoClient *mongo.Client
var mongoDB *mongo.Database
var assetCollection *mongo.Collection
var scanCollection *mongo.Collection

func connectMongo() {
	uri := viper.GetString("MongoUrl")
	database := viper.GetString("MongoDatabase")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(20))
	if err != nil {
		log.Panic(fmt.Sprintf("connect mongo error: %s", err))
	}

	if err = mongoClient.Ping(context.TODO(), nil); err != nil {
		log.Panic(fmt.Sprintf("connect mongo error: %s", err))
	}

	mongoDB = mongoClient.Database(database)
	assetCollection = mongoDB.Collection("asset")
	scanCollection = mongoDB.Collection("scan")
	log.Println(fmt.Sprintf("success connect to %s %s", uri, database))
}

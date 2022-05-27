package main

import (
	"Concord/Authentication"
	"Concord/CustomErrors"
	"Concord/Messaging"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

func main() {

	fmt.Printf("Program started at %s \n", time.Now().Format(time.RFC822))

	//Read required program args
	runTimeArgs := readRunTimeArgs()
	if !runTimeArgs.valid {
		fmt.Printf("Exiting, program args invalid\n")
		os.Exit(1)
	}

	//Connect to Mongo database
	MongoURI := "mongodb://" + runTimeArgs.dbUserMongo + ":" + runTimeArgs.dbPassMongo + "@" + runTimeArgs.dbHostMongo + ":" + runTimeArgs.dbPortMongo + "/" + runTimeArgs.dbNameMongo + "?authSource=admin"
	fmt.Printf(MongoURI + "\n")
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(MongoURI))
	if err != nil {
		CustomErrors.LogError(5002, "FATAL", true, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = mongoClient.Connect(ctx)
	if err != nil {
		CustomErrors.LogError(5001, "FATAL", true, err)
	}
	defer func(mongoClient *mongo.Client, ctx context.Context) {
		err := mongoClient.Disconnect(ctx)
		if err != nil {
			CustomErrors.LogError(5003, "FATAL", true, err)
		}
	}(mongoClient, ctx)

	//Attempt to ping database
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		CustomErrors.LogError(5004, "FATAL", true, err)
	}

	//Mongo database pointer
	dbClient := mongoClient.Database(runTimeArgs.dbNameMongo)

	//Connect to redis database
	redisGlobalClient := redis.NewClient(&redis.Options{
		Addr:     runTimeArgs.redisGlobalHostAddr + ":" + runTimeArgs.redisGlobalHostPort,
		Password: runTimeArgs.redisGlobalPassword,
		DB:       0, // use default DB
	})

	//Ping redis global database
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = redisGlobalClient.Ping(ctx).Result()
	if err != nil {
		CustomErrors.LogError(5021, "FATAL", true, err)
	}

	//Create certs for
	Authentication.CheckAndCreateKeys()

	//Start message hub
	messageHub := Messaging.NewHub()
	go messageHub.Run()

	//Start RPC server
	go Messaging.StartRPCServer(messageHub)

	//Start serving client api requests
	startRestAPI(dbClient, redisGlobalClient, messageHub)

}

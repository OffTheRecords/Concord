package main

import (
	"Concord/Authentication"
	"Concord/CustomErrors"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

func main() {

	fmt.Printf("Program started at %s \n", time.Now().Format(time.RFC822))

	//Read required program args
	runTimeArgs := readRunTimeArgs()
	if !runTimeArgs.valid {
		fmt.Printf("Exiting, program args invalid")
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

	//Create certs for
	Authentication.CheckAndCreateKeys()

	//Start serving client api requests
	startRestAPI(dbClient)

}

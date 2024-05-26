package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client *mongo.Client
	dbName = "Jobly"
	DB     *mongo.Database
)

func DbConnection() error {
	clientOptions := options.Client().ApplyURI("mongodb+srv://weldonkipchirchir23:xoskcLqTzTWCgtJI@cluster0.pnysint.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error connecting to the database")
		return err
	}

	//check connection
	err = Client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging the MongoDB")
	}

	log.Println("Successfully  connected to the database")

	DB = Client.Database(dbName)

	return err
}

func GetCLient() *mongo.Client {
	return Client
}

func GetCollection(collectionName string) *mongo.Collection {
	return Client.Database(dbName).Collection(collectionName)
}

func DbDisconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()
		err := Client.Disconnect(ctx)
		if err != nil {
			log.Fatal("Error disconnecting from the database")
		}
		log.Println("Successfully disconnected from the database")
	}
}

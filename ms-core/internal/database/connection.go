package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection names
const (
	GroupsCollection               = "groups"
	StudentsCollection             = "students"
	GroupsHistoryCollection        = "groups_historydata"
	DemographicTestsCollection     = "demographic_tests"
	DemographicResponsesCollection = "demographic_responses"
)

var DB *mongo.Database

func ConnectDB(mongoURI string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(200).
		SetMinPoolSize(20).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	DB = client.Database("ms_core")
	log.Println("Connected to MongoDB successfully")

	// Create Indexes in background
	go func() {
		idxCtx, cancelIdx := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelIdx()

		_, err := DB.Collection(GroupsCollection).Indexes().CreateOne(idxCtx, mongo.IndexModel{
			Keys: bson.D{{Key: "section_id", Value: 1}},
		})
		if err != nil {
			log.Printf("Warning: Failed to create index on %s: %v", GroupsCollection, err)
		}

		_, err = DB.Collection(GroupsHistoryCollection).Indexes().CreateOne(idxCtx, mongo.IndexModel{
			Keys: bson.D{{Key: "section_id", Value: 1}, {Key: "formed_at", Value: -1}},
		})
		if err != nil {
			log.Printf("Warning: Failed to create index on %s: %v", GroupsHistoryCollection, err)
		}
	}()
}

func DisconnectDB() {
	if DB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := DB.Client().Disconnect(ctx)
		if err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Disconnected from MongoDB")
		}
	}
}

func GetDatabase() *mongo.Database {
	return DB
}

func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}

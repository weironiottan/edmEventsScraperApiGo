package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addEdmEventsToLasVegasEdmEventsCollection() {
	// Set connection options
	clientOptions := options.Client().ApplyURI("mongodb://mongo:AK23at0qveZ27v92ylwX@containers-us-west-83.railway.app:7617")

	// Set authentication options
	credential := options.Credential{
		Username: "",
		Password: "",
	}

	clientOptions.SetAuth(credential)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Disconnect from MongoDB
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get a handle to the "test" database and "persons" collection
	collection := client.Database("edmEvents").Collection("lasVegasEdmEventsCollection")

	// Insert a document
	edmEvent := EdmEvent{
		ClubName: "dummy club name",
	}
	insertResult, err := collection.InsertOne(ctx, edmEvent)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted document ID:", insertResult.InsertedID)

	fmt.Println("Connected to MongoDB!")
}

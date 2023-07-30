package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

func (app *application) addEdmEventsToLasVegasEdmEventsCollection(w http.ResponseWriter, r *http.Request) {
	allEdmEvents := getEdmEventsFromAllLasVegas()
	app.deleteAllDocumentsInLasVegasEdmEventsCollection()
	app.insertEdmEventsIntoLasVegasEdmEventsCollection(allEdmEvents)

	//printOK := "Everything processed Successfully"
	err := app.writeJSON(w, http.StatusOK, "", nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (app *application) deleteAllDocumentsInLasVegasEdmEventsCollection() {
	// Set connection options
	clientOptions := options.Client().ApplyURI(app.dbConfig.mongoUrl)

	// Set authentication options
	credential := options.Credential{
		Username: app.dbConfig.mongoUser,
		Password: app.dbConfig.mongoPassword,
	}

	clientOptions.SetAuth(credential)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get a handle to the "test" database and "persons" collection
	collection := client.Database("edmEvents").Collection("lasVegasEdmEventsCollection")

	fmt.Println("Connected to MongoDB!")
	deleteResult, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted documents count:", deleteResult.DeletedCount)

	// Disconnect from MongoDB
	defer cancel()
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
}

func (app *application) insertEdmEventsIntoLasVegasEdmEventsCollection(allEdmEvents []EdmEvent) {
	// Set connection options
	clientOptions := options.Client().ApplyURI(app.dbConfig.mongoUrl)

	// Set authentication options
	credential := options.Credential{
		Username: app.dbConfig.mongoUser,
		Password: app.dbConfig.mongoPassword,
	}

	clientOptions.SetAuth(credential)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get a handle to the "test" database and "persons" collection
	collection := client.Database("edmEvents").Collection("lasVegasEdmEventsCollection")

	fmt.Println("Connected to MongoDB!")

	// Convert slice of persons to a slice of documents
	documents := make([]interface{}, len(allEdmEvents))
	for i, edmEvent := range allEdmEvents {
		documents[i] = edmEvent
	}

	// Insert multiple documents
	insertResult, err := collection.InsertMany(ctx, documents)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted document IDs:", insertResult.InsertedIDs)

	// Disconnect from MongoDB
	defer cancel()
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()
}

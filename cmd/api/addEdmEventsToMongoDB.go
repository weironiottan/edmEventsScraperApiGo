package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

func (app *application) addEdmEventsToLasVegasEdmEventsCollection(w http.ResponseWriter, r *http.Request) {
	edmEvents := getEdmEventsFromAllLasVegas()
	_, err := app.dbSnippets.DeleteMany(edmEvents)

	if err != nil {
		app.logger.Fatal("Error Deleting documents from collection: %v", err)
	}

	_, err = app.dbSnippets.InsertMany(edmEvents)

	if err != nil {
		app.logger.Fatal("Error Inserting documents from collection: %v", err)
	}

	//printOK := "Everything processed Successfully"
	err = app.writeJSON(w, http.StatusOK, "Successfully Scrapped Data and Updated Mongo DB", nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

type SnippetModelInterface interface {
	InsertMany(edmEvents []EdmEvent) (*mongo.InsertManyResult, error)
	DeleteMany(edmEvents []EdmEvent) (*mongo.DeleteResult, error)
}

// SnippetModel Define a SnippetModel type which wraps a MongoDB connection pool.
type SnippetModel struct {
	DB         *mongo.Client
	collection *mongo.Collection
}

func (m *SnippetModel) DeleteMany(edmEvents []EdmEvent) (*mongo.DeleteResult, error) {
	// Grab the Collection from MongoDB
	//collection := m.DB.Database("edmEvents").Collection("lasVegasEdmEventsCollection")

	// Delete all Documents in the Collection
	deleteResult, err := m.collection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Deleted documents count:", deleteResult.DeletedCount)
	return deleteResult, nil
}

func (m *SnippetModel) InsertMany(edmEvents []EdmEvent) (*mongo.InsertManyResult, error) {

	// Convert slice of edmEvents to a slice of documents
	documents := make([]interface{}, len(edmEvents))
	for i, edmEvent := range edmEvents {
		documents[i] = edmEvent
	}

	// InsertMany multiple documents
	insertResult, err := m.collection.InsertMany(context.TODO(), documents)
	if err != nil {
		return nil, err
	}

	return insertResult, nil
}

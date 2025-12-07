package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type SnippetModelInterface interface {
	InsertMany(edmEvents []EdmEvent) error
	DeleteMany(edmEvents []EdmEvent) error
}

// SnippetModel Define a SnippetModel type which wraps a Firestore client.
type SnippetModel struct {
	Client     *firestore.Client
	ctx        context.Context
	Collection string
}

func (m *SnippetModel) DeleteMany(edmEvents []EdmEvent) error {
	ctx := context.Background()
	fmt.Println("Started To Delete Documents in Collection: ", m.Collection)

	// Get a reference to the collection
	collRef := m.Client.Collection(m.Collection)

	// Get all documents in the collection
	iter := collRef.Documents(ctx)
	numDeleted := 0

	// Use a batched write for better performance
	batch := m.Client.BulkWriter(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate documents: %v", err)
		}

		// Add a delete operation to the batch
		batch.Delete(doc.Ref)
		numDeleted++
	}

	// Commit the batch
	batch.End()

	fmt.Printf("Deleted documents count: %d\n", numDeleted)
	return nil
}

func (m *SnippetModel) InsertMany(edmEvents []EdmEvent) error {
	ctx := context.Background()

	// Use a batched write for better performance
	batch := m.Client.BulkWriter(ctx)

	// Add each event to the batch
	for _, event := range edmEvents {
		docRef := m.Client.Collection(m.Collection).NewDoc()
		batch.Set(docRef, event)
	}

	// Commit the batch
	batch.End()

	fmt.Printf("Inserted %d documents\n", len(edmEvents))
	return nil
}

func (app *application) addEdmEventsToFirestore(ScrapingURLs ScrapingURLs) {
	edmEvents := getEdmEventsFromAllLasVegas(ScrapingURLs)

	err := app.dbSnippets.DeleteMany(edmEvents)
	if err != nil {
		app.logger.Fatalf("Error deleting documents from Firestore: %v", err)
	}

	err = app.dbSnippets.InsertMany(edmEvents)
	if err != nil {
		app.logger.Fatalf("Error inserting documents to Firestore: %v", err)
	}

	app.logger.Print("Successfully scraped data and updated Firestore")
}

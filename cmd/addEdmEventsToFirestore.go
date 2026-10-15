package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// Scrape metadata lives in this collection, one document per events collection, so each
// environment's events collection gets its own timestamp: meta/{COLLECTION_NAME}
const metaCollection = "meta"

type SnippetModelInterface interface {
	InsertMany(edmEvents []EdmEvent) error
	DeleteMany(edmEvents []EdmEvent) error
	SetLastUpdated(lastUpdated time.Time) error
}

// ScrapeMeta is the document the API reads to show when the shows were last updated
type ScrapeMeta struct {
	LastUpdated time.Time `firestore:"lastUpdated"`
}

// SnippetModel Define a SnippetModel type which wraps a Firestore client.
type SnippetModel struct {
	Client     *firestore.Client
	ctx        context.Context
	Collection string
}

// waitForJobs blocks until every BulkWriter job finishes and joins their errors
func waitForJobs(jobs []*firestore.BulkWriterJob) error {
	var errs []error
	for _, job := range jobs {
		if _, err := job.Results(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d of %d writes failed: %w", len(errs), len(jobs), errors.Join(errs...))
	}
	return nil
}

func (m *SnippetModel) DeleteMany(edmEvents []EdmEvent) error {
	ctx := context.Background()
	fmt.Println("Started To Delete Documents in Collection: ", m.Collection)

	// Get a reference to the collection
	collRef := m.Client.Collection(m.Collection)

	// Get all documents in the collection
	iter := collRef.Documents(ctx)
	defer iter.Stop()

	// Use a batched write for better performance
	batch := m.Client.BulkWriter(ctx)
	var jobs []*firestore.BulkWriterJob

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			batch.End()
			return fmt.Errorf("failed to iterate documents: %v", err)
		}

		// Add a delete operation to the batch
		job, err := batch.Delete(doc.Ref)
		if err != nil {
			batch.End()
			return fmt.Errorf("failed to queue delete: %v", err)
		}
		jobs = append(jobs, job)
	}

	// Commit the batch
	batch.End()
	if err := waitForJobs(jobs); err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}

	fmt.Printf("Deleted documents count: %d\n", len(jobs))
	return nil
}

func (m *SnippetModel) InsertMany(edmEvents []EdmEvent) error {
	ctx := context.Background()

	// Use a batched write for better performance
	batch := m.Client.BulkWriter(ctx)
	jobs := make([]*firestore.BulkWriterJob, 0, len(edmEvents))

	// Add each event to the batch
	for _, event := range edmEvents {
		docRef := m.Client.Collection(m.Collection).NewDoc()
		job, err := batch.Set(docRef, event)
		if err != nil {
			batch.End()
			return fmt.Errorf("failed to queue insert: %v", err)
		}
		jobs = append(jobs, job)
	}

	// Commit the batch
	batch.End()
	if err := waitForJobs(jobs); err != nil {
		return fmt.Errorf("failed to insert documents: %w", err)
	}

	fmt.Printf("Inserted %d documents\n", len(edmEvents))
	return nil
}

func (m *SnippetModel) SetLastUpdated(lastUpdated time.Time) error {
	doc := m.Client.Collection(metaCollection).Doc(m.Collection)
	if _, err := doc.Set(context.Background(), ScrapeMeta{LastUpdated: lastUpdated}); err != nil {
		return fmt.Errorf("failed to set last updated: %w", err)
	}
	return nil
}

// replaceEdmEvents swaps the stored events for a fresh scrape and records when it happened.
// The timestamp is only written once the events are safely stored, so a failed run never
// makes stale data look fresh.
func (app *application) replaceEdmEvents(edmEvents []EdmEvent, scrapedAt time.Time) error {
	if err := app.dbSnippets.DeleteMany(edmEvents); err != nil {
		return fmt.Errorf("deleting documents from Firestore: %w", err)
	}

	if err := app.dbSnippets.InsertMany(edmEvents); err != nil {
		return fmt.Errorf("inserting documents to Firestore: %w", err)
	}

	if err := app.dbSnippets.SetLastUpdated(scrapedAt); err != nil {
		return fmt.Errorf("recording scrape time in Firestore: %w", err)
	}

	return nil
}

func (app *application) addEdmEventsToFirestore(ScrapingURLs ScrapingURLs) {
	edmEvents := getEdmEventsFromAllLasVegas(ScrapingURLs)

	if err := app.replaceEdmEvents(edmEvents, time.Now().UTC()); err != nil {
		app.logger.Fatalf("Error updating Firestore: %v", err)
	}

	app.logger.Print("Successfully scraped data and updated Firestore")
}

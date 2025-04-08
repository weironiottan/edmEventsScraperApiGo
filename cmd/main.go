package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"os"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware.
type application struct {
	config     config
	logger     *log.Logger
	dbConfig   DBConfig
	dbSnippets SnippetModelInterface
}

type DBConfig struct {
	projectID  string
	databaseID string
	collection string
}

func main() {

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Debug Scraper
	//dmEvents := getEdmEventsFromAllLasVegas()
	//println(dmEvents)

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		log.Fatalf("Environment variable not set %v", projectID)
	}

	databaseID := os.Getenv("DATABASE_ID")
	if databaseID == "" {
		log.Fatalf("Environment variable not set %v", databaseID)
	}

	collection := os.Getenv("COLLECTION_NAME")
	if collection == "" {
		log.Fatalf("Environment variable not set %v", collection)
	}
	// Declare an instance of the config struct.
	var cfg config

	dbConfig := DBConfig{
		projectID:  projectID,
		databaseID: databaseID,
		collection: collection,
	}

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &application{
		config:   cfg,
		dbConfig: dbConfig,
		logger:   logger,
	}

	db, err := app.openDB()

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	app.dbSnippets = &SnippetModel{
		Client:     db,
		Collection: collection,
	}

	// This should bubble up and error in case there is a fatal error, such as a wrong scrape or something
	// We will have to differentiate between a bad scrape that can continue scrapping other events and still fail the job
	// And really bad ones where we stop the process
	app.addEdmEventsToFirestore()

}

func (app *application) openDB() (*firestore.Client, error) {

	ctx := context.Background()

	client, err := firestore.NewClientWithDatabase(ctx, app.dbConfig.projectID, app.dbConfig.databaseID)

	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to Firestore!")

	return client, nil
}

// Alternative initialization with credentials file
func (app *application) openDBDebuggingMode() (*firestore.Client, error) {
	// Create a Firestore client with credentials from GCP SA
	credentialsJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")
	if credentialsJSON == "" {
		log.Fatalf("Environment variable not set %v", credentialsJSON)
	}
	ctx := context.Background()
	option.WithCredentialsJSON([]byte(credentialsJSON))
	client, err := firestore.NewClientWithDatabase(ctx, app.dbConfig.projectID, app.dbConfig.databaseID, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, err
	}
	return client, nil
}

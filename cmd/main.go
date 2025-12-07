package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
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

type ScrapingURLs struct {
	TaoGroupHospitality string
	Liv                 string
	Wynn                string
	Zouk                string
}

func main() {

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Debug Scraper
	// dmEvents := getEdmEventsFromAllLasVegas()
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

	ScrapingURLs := ScrapingURLs{
		TaoGroupHospitality: "https://taogroup.com/wp-json/wp/v2/events?event_city%%5B%%5D=81&filter%%5Bmeta_compare%%5D=%%3E%%3D&filter%%5Bmeta_key%%5D=event_start_date&filter%%5Bmeta_value%%5D=1720422000000&filter%%5Border%%5D=asc&filter%%5Borderby%%5D=meta_value&",
		Liv:                 "https://www.livnightclub.com/wp-admin/admin-ajax.php?action=uvpx&uvaction=uwspx_loadevents&date=",
		Wynn:                "https://www.wynnsocial.com/events/",
		Zouk:                "https://zoukgrouplv.com/wp-admin/admin-ajax.php?action=uvwp_loadmoreevents&venuegroup=all&caldate=",
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
	app.addEdmEventsToFirestore(ScrapingURLs)

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

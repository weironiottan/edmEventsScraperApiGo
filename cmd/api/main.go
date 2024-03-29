package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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
	dbConfig   dbConfiguration
	dbSnippets SnippetModelInterface
}

type dbConfiguration struct {
	mongoUrl      string
	mongoHost     string
	mongoUser     string
	mongoPassword string
	mongoPort     string
}

func main() {
	// Declare an instance of the config struct.
	var cfg config
	var dbConfig dbConfiguration

	// Initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &application{
		config:   cfg,
		logger:   logger,
		dbConfig: dbConfig,
	}
	app.fetchEnvVariables()

	db, err := app.openDB()

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Disconnect(context.TODO())

	app.dbSnippets = &SnippetModel{
		DB:         db,
		collection: db.Database("edmEvents").Collection("lasVegasEdmEventsCollection"),
	}

	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.port, "port", app.config.port, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Declare a new servemux and add a /v1/healthcheck route which dispatches requests
	// to the healthcheckHandler method (which we will create in a moment).
	mux := http.NewServeMux()
	mux.HandleFunc("/home", app.home)
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("/v1/fetchWynnEdmEvents", app.fetchWynnEdmEvents)
	mux.HandleFunc("/v1/fetchTaoGroupHospitalityEdmEvents", app.fetchTaoGroupHospitalityEdmEvents)
	mux.HandleFunc("/v1/fetchZoukEdmEvents", app.fetchZoukEdmEvents)
	mux.HandleFunc("/v1/fetchLivEdmEvents", app.fetchLivEdmEvents)
	mux.HandleFunc("/", app.notFoundRoute)

	// Declare a HTTP server with some sensible timeout settings, which listens on the
	// port provided in the config struct and uses the servemux we created above as the
	// handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the HTTP server.
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	app.scheduledTaskToGrabEdmEventsEvery24hrs()
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func (app *application) fetchEnvVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	portStr := os.Getenv("PORT")

	port, err := strconv.Atoi(portStr)
	app.config.port = port
	if err != nil {
		log.Fatalf("Error converting port to integer: %v", err)
	}

	app.dbConfig = dbConfiguration{
		mongoUrl:      os.Getenv("MONGO_URL"),
		mongoHost:     os.Getenv("MONGOHOST"),
		mongoPort:     os.Getenv("MONGOPORT"),
		mongoUser:     os.Getenv("MONGOUSER"),
		mongoPassword: os.Getenv("MONGOPASSWORD"),
	}
}

func (app *application) openDB() (*mongo.Client, error) {

	// Set connection options
	clientOptions := options.Client().ApplyURI(app.dbConfig.mongoUrl)
	// Set authentication options
	credential := options.Credential{
		Username: app.dbConfig.mongoUser,
		Password: app.dbConfig.mongoPassword,
	}
	clientOptions.SetAuth(credential)

	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	client.Database("edmEvents").Collection("lasVegasEdmEventsCollection")

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

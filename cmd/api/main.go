package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const version = "1.0.0"

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
type config struct {
	port int
	env  string
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type application struct {
	config   config
	logger   *log.Logger
	dbConfig dbConfiguration
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

	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	app.fetchEnvVariables()
	flag.IntVar(&cfg.port, "port", app.config.port, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// Declare a new servemux and add a /v1/healthcheck route which dispatches requests
	// to the healthcheckHandler method (which we will create in a moment).
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("/v1/fetchWynnEdmEvents", app.fetchWynnEdmEvents)
	mux.HandleFunc("/v1/fetchHakkasanGroupEdmEvents", app.fetchHakkasanGroupEdmEvents)
	mux.HandleFunc("/v1/fetchZoukEdmEvents", app.fetchZoukEdmEvents)
	mux.HandleFunc("/v1/addEdmEventsToLasVegasEdmEventsCollection", app.addEdmEventsToLasVegasEdmEventsCollection)

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
	err := srv.ListenAndServe()
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

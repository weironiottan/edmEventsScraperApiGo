# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based web scraper that collects EDM (Electronic Dance Music) event data from various Las Vegas nightclub websites and stores them in Google Cloud Firestore. The application runs as a containerized job on Google Cloud Platform.

## Build and Run Commands

```bash
# Build the application
go build -o edmEventsScraperJob ./cmd

# Run the application locally
go run ./cmd

# Run tests
go test ./cmd

# Run tests with coverage
go test -cover ./cmd

# Build Docker image
docker build -t edm-events-scraper .

# Run in Docker
docker run --env-file .env edm-events-scraper
```

## Required Environment Variables

The application requires these environment variables to run:
- `GOOGLE_CLOUD_PROJECT` - GCP project ID
- `DATABASE_ID` - Firestore database ID
- `COLLECTION_NAME` - Firestore collection name for storing events
- `GOOGLE_APPLICATION_CREDENTIALS_JSON` - (optional) For local debugging with service account credentials

## Architecture

### Core Flow

1. **main.go** - Entry point that initializes Firestore connection and triggers the scraping process
2. **addEdmEventsToFirestore.go** - Orchestrates the full scrape-and-replace workflow:
   - Calls `getEdmEventsFromAllLasVegas()` to aggregate events from all venues
   - Deletes all existing documents in Firestore collection
   - Inserts newly scraped events in batch
3. **fetchEdmEventsHelper.go** - Aggregates events from all venue scrapers

### Venue Scrapers

Each venue has a dedicated scraper file that returns `[]EdmEvent`:

- **fetchWynnEdmEvents.go** - Scrapes Wynn venues using Colly HTML scraping
- **fetchZoukEdmEvents.go** - Scrapes Zouk using pagination API (lazy-loaded events)
- **fetchTaoGroupHospitalityEdmEvents.go** - Scrapes Tao Group venues via WordPress JSON API with pagination
- **fetchLivEdmEvents.go** - Scrapes LIV nightclub using pagination API

All scrapers:
- Generate unique IDs for each event using `getGUID()`
- Filter out past events using `isPastDate()`
- Filter out unwanted venues/events using `filterUnwantedEvents()`
- Normalize dates to RFC3339 format
- Return standardized `EdmEvent` structs

### Data Models

**EdmEvent** (types.go) - Core event structure:
```go
type EdmEvent struct {
    Id             string // UUID
    ClubName       string // Venue name
    ArtistName     string // Performer name
    EventDate      string // RFC3339 formatted date
    TicketUrl      string // Link to event/tickets
    ArtistImageUrl string // Artist photo URL
}
```

### Database Layer

**SnippetModel** (addEdmEventsToFirestore.go):
- `InsertMany()` - Batch writes events using Firestore BulkWriter
- `DeleteMany()` - Clears collection using batch delete operations
- Uses Firestore client with database-specific connection

### Utility Functions

- **dateHelpers.go** - Date parsing and validation (multiple format conversions, past date checking)
- **helpers.go** - JSON serialization, event filtering, URL parsing
- **guid.go** - UUID generation for event IDs

## Important Implementation Details

### Scraper Strategy Variations

Each venue requires a different scraping approach:
- **Wynn**: Direct HTML scraping with Colly
- **Zouk/LIV**: Pagination APIs that return HTML fragments
- **Tao Group**: WordPress REST API with structured JSON responses

### Date Handling

The codebase handles multiple date formats from different venues:
- `formatDateFrom_YYYYMMDD_toRFC3339()` - Wynn format
- `formatDateFrom_MM_DD_YYYY_toRFC3339()` - Tao Group format
- All dates standardized to RFC3339 before storage

### Filtering Logic

Two levels of filtering:
1. **Temporal**: `isPastDate()` removes events that already occurred
2. **Venue-specific**: `filterUnwantedEvents()` removes non-EDM events (restaurants, festivals, etc.)

### Database Strategy

The application uses a **full replace strategy**:
- Deletes entire collection on each run
- Re-inserts all current/future events
- Simple but ensures no stale data

## Deployment

- Uses Google Cloud Build (`cloudbuild.yaml`) to build and push container images
- Configured to deploy to GCP Container Registry: `gcr.io/$PROJECT_ID/edm-events-scraper:latest`
- Intended to run as a scheduled Cloud Run job or similar GCP compute service

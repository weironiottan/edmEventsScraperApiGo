# EDM Events Scraper API (Go)

A Go-based web scraper that collects Electronic Dance Music (EDM) event data from various Las Vegas nightclub websites and stores them in Google Cloud Firestore. The application runs as a containerized job on Google Cloud Platform.

## 🎯 Features

- **Multi-venue scraping**: Collects events from 4 major Las Vegas nightclub groups:
  - Wynn (HTML scraping with Colly)
  - Zouk (Pagination API)
  - Tao Group Hospitality (WordPress REST API)
  - LIV (Pagination API)
- **Smart filtering**: Automatically filters out past events and non-EDM venues (restaurants, etc.)
- **Date normalization**: Handles multiple date formats and standardizes to RFC3339
- **Batch operations**: Efficient Firestore storage with BulkWriter
- **Full test coverage**: 93.2% coverage on core scraping functions

## 📋 Prerequisites

- Go 1.23.0 or higher
- Docker (for containerized deployment)
- Google Cloud Platform account with Firestore enabled
- Pre-commit (optional, for development)

## 🚀 Quick Start

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd edmEventsScraperApiGo
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   export GOOGLE_CLOUD_PROJECT="your-project-id"
   export DATABASE_ID="your-database-id"
   export COLLECTION_NAME="edm-events"
   
   # For local debugging only
   export GOOGLE_APPLICATION_CREDENTIALS_JSON='{"type":"service_account",...}'
   ```

4. **Run the scraper**
   ```bash
   go run ./cmd
   ```

### Running Tests

```bash
# Run all tests
go test ./cmd

# Run tests with coverage
go test -cover ./cmd

# Run tests with verbose output
go test -v ./cmd

# Generate coverage report
go test -coverprofile=coverage.out ./cmd
go tool cover -html=coverage.out
```

### Build and Run with Docker

```bash
# Build the Docker image
docker build -t edm-events-scraper .

# Run the container
docker run --env-file .env edm-events-scraper
```

## 🏗️ Project Structure

```
.
├── cmd/                                           # Main application package
│   ├── main.go                                    # Entry point
│   ├── addEdmEventsToFirestore.go                 # Firestore operations
│   ├── fetchEdmEventsHelper.go                    # Aggregates all scrapers
│   ├── fetchWynnEdmEvents.go                      # Wynn scraper
│   ├── fetchZoukEdmEvents.go                      # Zouk scraper
│   ├── fetchTaoGroupHospitalityEdmEvents.go       # Tao Group scraper
│   ├── fetchLivEdmEvents.go                       # LIV scraper
│   ├── dateHelpers.go                             # Date parsing utilities
│   ├── helpers.go                                 # General utilities
│   ├── guid.go                                    # UUID generation
│   ├── types.go                                   # Data models
│   └── *_test.go                                  # Test files
├── Dockerfile                                      # Container configuration
├── cloudbuild.yaml                                 # GCP Cloud Build config
├── .pre-commit-config.yaml                         # Pre-commit hooks
├── go.mod                                          # Go module definition
├── CLAUDE.md                                       # Claude Code documentation
└── README.md                                       # This file
```

## 🔧 Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `GOOGLE_CLOUD_PROJECT` | GCP project ID | Yes |
| `DATABASE_ID` | Firestore database ID | Yes |
| `COLLECTION_NAME` | Firestore collection name | Yes |
| `GOOGLE_APPLICATION_CREDENTIALS_JSON` | Service account JSON (for local dev) | No |

### Scraper Configuration

Each venue scraper can be configured in `fetchEdmEventsHelper.go`. To add or remove venues, modify the `getEdmEventsFromAllLasVegas()` function.

Unwanted events are filtered in each scraper file. For example, in `fetchTaoGroupHospitalityEdmEvents.go`:

```go
edmEvents = filterUnwantedEvents(edmEvents, []string{
    "lavo italian restaurant las vegas", 
    "lavo italian restaurant",
})
```

## 🧪 Testing

This project uses table-driven tests for comprehensive coverage.

### Test Structure

- **Positive tests**: Validate successful scraping scenarios
- **Negative tests**: Test error handling and edge cases

### Running Specific Tests

```bash
# Run only Tao Group tests
go test -run TestScrapeTaoGroupHospitalityEdmEvents ./cmd

# Run only positive tests
go test -run TestScrapeTaoGroupHospitalityEdmEvents_Positive ./cmd

# Run only negative tests
go test -run TestScrapeTaoGroupHospitalityEdmEvents_Negative ./cmd
```

### Coverage Requirements

Files with test coverage should maintain **at least 90%** coverage. Current coverage:

- `scrapeTaoGroupHospitalityEdmEvents`: 93.2%
- `getTaoGroupHospitalityEdmEvents`: 87.5%
- `filterOutTimeFromDate`: 100%
- `filterOutLasVegasFromTitle`: 100%
- `formatPaginatedURL`: 100%

## 🔄 Development Workflow

### Pre-commit Hooks

This project uses pre-commit hooks to ensure code quality:

```bash
# Install pre-commit hooks
pre-commit install

# Run hooks manually
pre-commit run --all-files
```

Current hooks:
- `go-fmt`: Format Go code
- `go-mod-tidy`: Clean up dependencies
- `go-test-mod`: Run all tests

### Adding a New Venue Scraper

1. Create scraper function in `cmd/fetchNewVenueEdmEvents.go`
2. Implement: `scrapeNewVenueEdmEvents(url string) []EdmEvent`
3. Add tests in `cmd/fetchNewVenueEdmEvents_test.go`
4. Update `fetchEdmEventsHelper.go` to include the new scraper
5. Ensure 90%+ test coverage

## 📊 Data Model

### EdmEvent

```go
type EdmEvent struct {
    Id             string  // UUID
    ClubName       string  // Venue name (lowercase, no "- Las Vegas")
    ArtistName     string  // Performer name (lowercase)
    EventDate      string  // RFC3339 formatted date
    TicketUrl      string  // Link to event/tickets
    ArtistImageUrl string  // Artist photo URL
}
```

## 🚢 Deployment

### Google Cloud Platform

The application is designed to run as a Cloud Run job or similar GCP service:

1. **Build and push image**:
   ```bash
   gcloud builds submit --config cloudbuild.yaml
   ```

2. **Deploy to Cloud Run**:
   ```bash
   gcloud run jobs create edm-events-scraper \
     --image gcr.io/$PROJECT_ID/edm-events-scraper:latest \
     --set-env-vars GOOGLE_CLOUD_PROJECT=$PROJECT_ID,DATABASE_ID=$DB_ID,COLLECTION_NAME=$COLLECTION
   ```

3. **Schedule execution** (using Cloud Scheduler):
   ```bash
   gcloud scheduler jobs create http edm-scraper-daily \
     --schedule="0 0 * * *" \
     --uri="https://run.googleapis.com/..." \
     --http-method=POST
   ```

### Manual Deployment

Build and run locally:
```bash
# Build the binary
go build -v -o edmEventsScraperJob ./cmd

# Run the scraper
./edmEventsScraperJob
```

## 🐛 Troubleshooting

### Common Issues

**Missing dependencies error**:
```bash
go mod download
go mod tidy
```

**Firestore connection issues**:
- Verify `GOOGLE_CLOUD_PROJECT` and `DATABASE_ID` are set
- Check service account permissions
- Ensure Firestore API is enabled in GCP

**Test failures**:
- Clear test cache: `go clean -testcache`
- Verify date formats match expected values
- Check that mock server responses use correct JSON field names (`post_title` not `PostTitle`)

**Pre-commit hook failures**:
- Ensure Go modules are downloaded: `go mod download`
- The hooks run in an isolated environment, so dependencies must be available
- Use `go-test-mod` instead of `go-test-pkg` for module-level testing

## 📝 Important Notes

### Database Strategy

The scraper uses a **full replace strategy**:
1. Scrapes all events from all venues
2. Deletes entire Firestore collection
3. Inserts all newly scraped events

This ensures no stale data but requires careful error handling.

### Date Handling

Different venues use different date formats:
- Wynn: `YYYYMMDD`
- Tao Group: `MM/DD/YYYY HH:MM AM/PM`
- Zouk/LIV: Various pagination-based formats

All dates are standardized to RFC3339 before storage.

### Filtering Logic

Two levels of filtering:
1. **Temporal**: Removes past events via `isPastDate()`
2. **Venue-specific**: Removes non-EDM events (restaurants, festivals) via `filterUnwantedEvents()`

Note: The `filterEvent()` function performs case-insensitive matching by lowercasing both the venue name and filter strings.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-venue`
3. Write tests for new functionality (table-driven tests preferred)
4. Ensure tests pass: `go test ./cmd`
5. Ensure coverage ≥ 90%: `go test -cover ./cmd`
6. Run pre-commit hooks: `pre-commit run --all-files`
7. Submit a pull request

## 📄 License

See LICENSE file for details.

## 📧 Contact

For questions or issues, please open an issue on GitHub.

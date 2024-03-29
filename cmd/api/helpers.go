package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to JSON, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include. We loop
	// through the header map and add each header to the http.ResponseWriter header map.
	// Note that it's OK if the provided header map is nil. Go doesn't throw an error
	// if you try to range over (or generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func extractEventDate(url string) string {
	regexPattern := regexp.MustCompile(`\d+`)
	extractedData := regexPattern.FindStringSubmatch(url)
	extractedDigits := extractedData[0]
	extractedDate := extractedDigits[len(extractedDigits)-8:]
	return extractedDate
}

func filterUnwantedEvents(edmEvents []EdmEvent, unWantedEvents []string) []EdmEvent {
	var filteredEdmEvents []EdmEvent

	for _, edmEvent := range edmEvents {
		isWantedClubName := filterEvent(edmEvent.ClubName, unWantedEvents)
		if isWantedClubName {
			filteredEdmEvents = append(filteredEdmEvents, edmEvent)
		}
	}
	return filteredEdmEvents
}

func filterEvent(str string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(strings.ToLower(str), substring) {
			return false
		}
	}
	return true
}

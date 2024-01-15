package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_getHakassanGroupEdmEvents(t *testing.T) {
	t.Run("Happy Path, return a correct response and Status OK", func(t *testing.T) {
		// Create a test server
		testServer := mockServerResponses(http.StatusOK, "This is a good response")
		defer testServer.Close()

		// Call the function with the test server URL
		response, err := getHakassanGroupEdmEvents(testServer.URL)
		if err != nil {
			t.Fatalf("error calling getHakassanGroupEdmEvents: %v", err)
		}
		defer response.Body.Close()

		// Check the response status code
		if response.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, response.StatusCode)
		}

	})

	t.Run("Negative Path, Got a bad response", func(t *testing.T) {
		// Create a test server
		testServer := mockServerResponses(http.StatusBadGateway, "This is a bad response")
		defer testServer.Close()

		// Call the function with the test server URL
		response, err := getHakassanGroupEdmEvents(testServer.URL)

		fmt.Println(err)
		got := fmt.Sprintf("%s", err)
		want := "status received was not a 200"

		// Check the response status code
		if response.StatusCode != http.StatusBadGateway {
			t.Errorf("expected status code %d, got %d", http.StatusBadGateway, response.StatusCode)
		}

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}

	})

	t.Run("Negative Path, Got a Good response, no data", func(t *testing.T) {
		testServer := mockServerResponses(http.StatusOK, "")
		defer testServer.Close()

		_, err := getHakassanGroupEdmEvents("")

		fmt.Println(err)
		got := fmt.Sprintf("%s", err)
		want := "error sending request: Get \"\": unsupported protocol scheme \"\""

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}

	})

}

func mockServerResponses(httpStatus int, response string) *httptest.Server {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(httpStatus)
		w.Write([]byte(response))

	}))
	return mockServer
}

package main

import (
	"log"
	"testing"
)

func Test_application_deleteAllDocumentsInLasVegasEdmEventsCollection(t *testing.T) {
	type fields struct {
		config   config
		logger   *log.Logger
		dbConfig dbConfiguration
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &application{
				config:   tt.fields.config,
				logger:   tt.fields.logger,
				dbConfig: tt.fields.dbConfig,
			}
			app.deleteAllDocumentsInLasVegasEdmEventsCollection()
		})
	}
}

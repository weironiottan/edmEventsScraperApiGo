package main

import (
	"errors"
	"log"
	"strings"
	"testing"
	"time"
)

// fakeSnippetModel records calls and can be told to fail at a given step
type fakeSnippetModel struct {
	calls       []string
	inserted    []EdmEvent
	lastUpdated time.Time
	failOn      string
}

func (f *fakeSnippetModel) step(name string) error {
	f.calls = append(f.calls, name)
	if f.failOn == name {
		return errors.New(name + " failed")
	}
	return nil
}

func (f *fakeSnippetModel) DeleteMany(edmEvents []EdmEvent) error {
	return f.step("delete")
}

func (f *fakeSnippetModel) InsertMany(edmEvents []EdmEvent) error {
	if err := f.step("insert"); err != nil {
		return err
	}
	f.inserted = edmEvents
	return nil
}

func (f *fakeSnippetModel) SetLastUpdated(lastUpdated time.Time) error {
	if err := f.step("setLastUpdated"); err != nil {
		return err
	}
	f.lastUpdated = lastUpdated
	return nil
}

func TestReplaceEdmEvents(t *testing.T) {
	scrapedAt := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	edmEvents := []EdmEvent{{Id: "1", ClubName: "omnia nightclub", ArtistName: "steve aoki"}}

	t.Run("replaces the events then records the scrape time", func(t *testing.T) {
		model := &fakeSnippetModel{}
		app := &application{logger: log.Default(), dbSnippets: model}

		if err := app.replaceEdmEvents(edmEvents, scrapedAt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got := strings.Join(model.calls, ","); got != "delete,insert,setLastUpdated" {
			t.Errorf("calls = %s, want delete,insert,setLastUpdated", got)
		}
		if len(model.inserted) != 1 {
			t.Errorf("inserted %d events, want 1", len(model.inserted))
		}
		if !model.lastUpdated.Equal(scrapedAt) {
			t.Errorf("lastUpdated = %v, want %v", model.lastUpdated, scrapedAt)
		}
	})

	for _, step := range []string{"delete", "insert"} {
		t.Run("does not record the scrape time when "+step+" fails", func(t *testing.T) {
			model := &fakeSnippetModel{failOn: step}
			app := &application{logger: log.Default(), dbSnippets: model}

			err := app.replaceEdmEvents(edmEvents, scrapedAt)

			if err == nil || !strings.Contains(err.Error(), step+" failed") {
				t.Fatalf("err = %v, want it to mention %q", err, step+" failed")
			}
			if !model.lastUpdated.IsZero() {
				t.Errorf("lastUpdated was set to %v after a failed %s", model.lastUpdated, step)
			}
		})
	}

	t.Run("reports a failure to record the scrape time", func(t *testing.T) {
		model := &fakeSnippetModel{failOn: "setLastUpdated"}
		app := &application{logger: log.Default(), dbSnippets: model}

		err := app.replaceEdmEvents(edmEvents, scrapedAt)

		if err == nil || !strings.Contains(err.Error(), "recording scrape time") {
			t.Fatalf("err = %v, want a scrape time error", err)
		}
	})
}

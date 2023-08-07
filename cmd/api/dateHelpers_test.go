package main

import (
	"testing"
)

func TestFormatDateFrom_YYYYMMDD_toRFC3339(t *testing.T) {
	t.Run("Date should be formatted to this format: 2023-01-01T00:00:00Z", func(t *testing.T) {
		got, _ := formatDateFrom_YYYYMMDD_toRFC3339("19881130")
		want := "1988-11-30T00:00:00Z"

		assertCorrectMessage(t, got, want)

	})

	t.Run("Date is not correctly formatted and should throw an error", func(t *testing.T) {
		_, err := formatDateFrom_YYYYMMDD_toRFC3339("11301988")
		want := "parsing time \"11301988\": month out of range"

		assertCorrectMessage(t, err.Error(), want)

	})
}

func TestFormatDateFrom_YYYY_MM_DD_toRFC3339(t *testing.T) {
	t.Run("Date should be formatted to this format: 2023-01-01T00:00:00Z", func(t *testing.T) {
		got, _ := formatDateFrom_YYYY_MM_DD_toRFC3339("1988-11-30")
		want := "1988-11-30T00:00:00Z"

		assertCorrectMessage(t, got, want)

	})

	t.Run("Date is not correctly formatted and should throw an error", func(t *testing.T) {
		_, err := formatDateFrom_YYYY_MM_DD_toRFC3339("1988-1130")
		want := "parsing time \"1988-1130\" as \"2006-01-02\": cannot parse \"30\" as \"-\""

		assertCorrectMessage(t, err.Error(), want)

	})
}

func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

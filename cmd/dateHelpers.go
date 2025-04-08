package main

import (
	"fmt"
	"time"
)

func formatDateFrom_YYYYMMDD_toRFC3339(dateToFormat string) (string, error) {
	//	For Reference RFC3339 looks like this: 2023-01-01T00:00:00Z
	dateFormat := "20060102"

	// Parse the string into a time value
	t, err := time.Parse(dateFormat, dateToFormat)
	if err != nil {
		fmt.Println("Error while parsing the date:", err)
		return "", err
	}

	formattedDate := t.Format(time.RFC3339)

	return formattedDate, nil
}

func formatDateFrom_YYYY_MM_DD_toRFC3339(dateToFormat string) (string, error) {
	//	For Reference RFC3339 looks like this: 2023-01-01T00:00:00Z
	dateFormat := "2006-01-02"

	// Parse the string into a time value
	t, err := time.Parse(dateFormat, dateToFormat)
	if err != nil {
		fmt.Println("Error while parsing the date:", err)
		return "", err
	}

	formattedDate := t.Format(time.RFC3339)

	return formattedDate, nil
}

func formatDateFrom_MM_DD_YYYY_toRFC3339(dateToFormat string) (string, error) {
	//	For Reference RFC3339 looks like this: 2023-01-01T00:00:00Z
	dateFormat := "01/02/2006"

	// Parse the string into a time value
	t, err := time.Parse(dateFormat, dateToFormat)
	if err != nil {
		fmt.Println("Error while parsing the date:", err)
		return "", err
	}

	formattedDate := t.Format(time.RFC3339)

	return formattedDate, nil
}

func formatDateFrom_YYYY_M_D_toRFC3339(dateToFormat string) (string, error) {
	//	For Reference RFC3339 looks like this: 2023-01-01T00:00:00Z
	dateFormat := "2006-1-2"

	// Parse the string into a time value
	t, err := time.Parse(dateFormat, dateToFormat)
	if err != nil {
		fmt.Println("Error while parsing the date:", err)
		return "", err
	}

	formattedDate := t.Format(time.RFC3339)

	return formattedDate, nil
}

func isPastDate(date string) (bool, error) {
	dateFormat := "2006-01-02T15:04:05Z"
	t, err := time.Parse(dateFormat, date)
	now := time.Now()

	eventDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, t.Location())

	if err != nil {
		return false, err
	}

	if eventDate.Before(currentDate) {
		return true, nil
	}

	return false, nil
}

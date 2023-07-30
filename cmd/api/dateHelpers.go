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

	formatedDate := t.Format(time.RFC3339)

	return formatedDate, nil
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

	formatedDate := t.Format(time.RFC3339)

	return formatedDate, nil
}

package main

import "github.com/google/uuid"

func getGUID() string {
	id := uuid.New()
	return id.String()
}

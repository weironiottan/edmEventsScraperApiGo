package main

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"time"
)

func scheduledTaskToGrabEdmEventsEvery24hrs() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("02:00").Do(func() {
		fmt.Println("***********")
		fmt.Println("Starting Scheduled Task")

		getEdmEventsFromAllLasVegas()

		fmt.Println("Finished Scheduled Task")
		fmt.Println("***********")
	})
	s.StartAsync()
}

package main

func getEdmEventsFromAllLasVegas() []EdmEvent {
	hakassanGroupEdmEvents := scrapeHakkasanGroupEdmEvents()
	wynnEdmEvents := scrapeWynnForEdmEvents()
	zoukEdmEvents := scrapeZoukEdmEvents()
	allEdmEvents := append(hakassanGroupEdmEvents, wynnEdmEvents...)
	allEdmEvents = append(allEdmEvents, zoukEdmEvents...)
	return allEdmEvents

}

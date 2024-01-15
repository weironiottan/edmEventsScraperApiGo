package main

func getEdmEventsFromAllLasVegas() []EdmEvent {
	wynnEdmEvents := scrapeWynnForEdmEvents()
	zoukEdmEvents := scrapeZoukEdmEvents()
	taoGroupHospitalityEdmEvents := scrapeTaoGroupHospitalityEdmEvents()
	livEdmEvents := scrapeLivForEdmEvents()
	allEdmEvents := append(zoukEdmEvents, wynnEdmEvents...)
	allEdmEvents = append(allEdmEvents, taoGroupHospitalityEdmEvents...)
	allEdmEvents = append(allEdmEvents, livEdmEvents...)
	return allEdmEvents

}

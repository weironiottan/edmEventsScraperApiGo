package main

func getEdmEventsFromAllLasVegas() []EdmEvent {
	wynnEdmEvents := scrapeWynnForEdmEvents()
	zoukEdmEvents := scrapeZoukEdmEvents()
	taoGroupHospitalityEdmEvents := scrapeTaoGroupHospitalityEdmEvents()
	allEdmEvents := append(zoukEdmEvents, wynnEdmEvents...)
	allEdmEvents = append(allEdmEvents, taoGroupHospitalityEdmEvents...)
	return allEdmEvents

}

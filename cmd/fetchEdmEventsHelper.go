package main

func getEdmEventsFromAllLasVegas(ScrapingURLs ScrapingURLs) []EdmEvent {
	wynnEdmEvents := scrapeWynnForEdmEvents()
	zoukEdmEvents := scrapeZoukEdmEvents()
	taoGroupHospitalityEdmEvents := scrapeTaoGroupHospitalityEdmEvents(ScrapingURLs.TaoGroupHospitality)
	livEdmEvents := scrapeLivForEdmEvents()
	allEdmEvents := append(zoukEdmEvents, wynnEdmEvents...)
	allEdmEvents = append(allEdmEvents, taoGroupHospitalityEdmEvents...)
	allEdmEvents = append(allEdmEvents, livEdmEvents...)
	return allEdmEvents

}

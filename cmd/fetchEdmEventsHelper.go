package main

func getEdmEventsFromAllLasVegas(ScrapingURLs ScrapingURLs) []EdmEvent {
	wynnEdmEvents := scrapeWynnForEdmEvents(ScrapingURLs.Wynn)
	zoukEdmEvents := scrapeZoukEdmEvents(ScrapingURLs.Zouk)
	taoGroupHospitalityEdmEvents := scrapeTaoGroupHospitalityEdmEvents(ScrapingURLs.TaoGroupHospitality)
	livEdmEvents := scrapeLivForEdmEvents(ScrapingURLs.Liv)
	allEdmEvents := append(zoukEdmEvents, wynnEdmEvents...)
	allEdmEvents = append(allEdmEvents, taoGroupHospitalityEdmEvents...)
	allEdmEvents = append(allEdmEvents, livEdmEvents...)
	return allEdmEvents

}

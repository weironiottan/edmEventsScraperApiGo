package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	// Embeds the timezone database so America/Los_Angeles loads in minimal containers
	_ "time/tzdata"

	"github.com/PuerkitoBio/goquery"
)

/*
Update: 10/14/2026

NOTE: We scrape LIV's events from Speakeasy (speakeasygo.com), not livnightclub.com.

livnightclub.com is behind a Cloudflare bot challenge ("Just a moment...") that returns a 403
to requests from Cloud Run, so the scraper got 0 events in prod. Speakeasy is LIV's ticketing
partner (every "Buy Tickets" link on livnightclub.com goes there) and lists the same events
with the same event IDs (EVE-XXXXXX).

The Speakeasy discovery page is server rendered (Next.js) and embeds the search results as JSON
in the __NEXT_DATA__ script tag. It returns 12 events at a time, sorted by date, so we page by
moving the startDate filter forward to the date of the last event we got.

We don't call Speakeasy's JSON API (production.speakeasygo.com) directly, its responses are
encrypted and it's behind Cloudflare too.
----
*/

const (
	livSpeakeasyVenueName = "liv las vegas" // Excludes LIV Beach, which is a dayclub
	livSpeakeasyImageUrl  = "https://media.speakeasygo.com/"
	livSpeakeasyEventUrl  = "https://speakeasygo.com/liv-nightclub-las-vegas/event?eid="
	livMaxPages           = 30
)

func scrapeLivForEdmEvents(scrapingUrl string) []EdmEvent {
	edmEvents := []EdmEvent{}

	lasVegas, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println("Error while scraping LIV: could not load timezone:", err)
		return edmEvents
	}

	// Speakeasy returns nothing if startDate is before its "today", which is in UTC.
	// UTC is always on or ahead of the Vegas date, so start from there.
	startDate := time.Now().UTC()
	endDate := startDate.AddDate(1, 0, 0)
	seenEventIds := map[string]bool{}

	for page := 0; page < livMaxPages; page++ {
		pageUrl := formatLivSpeakeasyURL(scrapingUrl, startDate, endDate)

		pageProps, err := getLivSpeakeasyEvents(pageUrl)
		if err != nil {
			fmt.Println("Error while scraping LIV:", err)
			break
		}

		newEvents := 0
		var lastEventDate time.Time
		for _, speakeasyEvent := range pageProps.Events {
			eventDate, err := time.ParseInLocation("2006-01-02 15:04:05", speakeasyEvent.StartDateTime, time.UTC)
			if err != nil {
				fmt.Println("Error while parsing the date:", err)
				continue
			}
			// Speakeasy times are UTC, a 10:30pm show in Vegas is the next day in UTC
			eventDate = eventDate.In(lasVegas)
			lastEventDate = eventDate

			if seenEventIds[speakeasyEvent.UniqueId] {
				continue
			}
			seenEventIds[speakeasyEvent.UniqueId] = true
			newEvents++

			if strings.ToLower(strings.TrimSpace(speakeasyEvent.Venue.Name)) != livSpeakeasyVenueName {
				continue
			}

			edmEvent, err := toLivEdmEvent(speakeasyEvent, eventDate)
			if err != nil {
				fmt.Println("Error while parsing the date:", err)
				continue
			}

			isPastDate, err := isPastDate(edmEvent.EventDate)
			if err != nil {
				fmt.Println("Error while parsing the date:", err)
				continue
			}
			if !isPastDate {
				edmEvents = append(edmEvents, edmEvent)
			}
		}

		if !pageProps.EventsHasMore || len(pageProps.Events) == 0 {
			break
		}

		// Next page starts on the last event's date, the overlap is deduped by event ID.
		// If a whole page was repeats (more than 12 events on one day), skip ahead a day.
		nextStartDate := lastEventDate
		if newEvents == 0 || !nextStartDate.After(startDate) {
			nextStartDate = startDate.AddDate(0, 0, 1)
		}
		startDate = nextStartDate
	}

	// Placeholder listings LIV posts before the lineup is announced, or for closed nights
	edmEvents = filterUnwantedArtists(edmEvents, []string{"special guest", "private event"})

	fmt.Println(len(edmEvents), "LIV edmEvents")
	fmt.Println("Scraping Completed for LIV!!!")
	return edmEvents
}

func toLivEdmEvent(speakeasyEvent LivSpeakeasyEvent, eventDate time.Time) (EdmEvent, error) {
	edmEvent := EdmEvent{}

	formattedDate, err := formatDateFrom_YYYY_MM_DD_toRFC3339(eventDate.Format("2006-01-02"))
	if err != nil {
		return edmEvent, err
	}

	edmEvent.Id = getGUID()
	edmEvent.ArtistName = strings.ToLower(strings.TrimSpace(speakeasyEvent.Title))
	edmEvent.ClubName = strings.ToLower(strings.TrimSpace(speakeasyEvent.Venue.Name))
	edmEvent.EventDate = formattedDate
	edmEvent.TicketUrl = livSpeakeasyEventUrl + url.QueryEscape(speakeasyEvent.UniqueId)
	if speakeasyEvent.Flyer.Key != "" {
		edmEvent.ArtistImageUrl = livSpeakeasyImageUrl + speakeasyEvent.Flyer.Key
	}

	return edmEvent, nil
}

func getLivSpeakeasyEvents(pageUrl string) (LivSpeakeasyPageProps, error) {
	var pageProps LivSpeakeasyPageProps

	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, pageUrl, nil)
	if err != nil {
		return pageProps, fmt.Errorf("error creating request: %w", err)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	response, err := client.Do(request)
	if err != nil {
		return pageProps, fmt.Errorf("error sending request to %s: %w", pageUrl, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return pageProps, fmt.Errorf("error reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return pageProps, fmt.Errorf("status %d from %s (content-type %q): %s",
			response.StatusCode, pageUrl, response.Header.Get("Content-Type"), truncate(string(body), 500))
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return pageProps, fmt.Errorf("error parsing HTML from %s: %w", pageUrl, err)
	}

	nextData := doc.Find("script#__NEXT_DATA__").Text()
	if nextData == "" {
		return pageProps, fmt.Errorf("no __NEXT_DATA__ found on %s, the page layout may have changed", pageUrl)
	}

	var livSpeakeasyNextData LivSpeakeasyNextData
	if err := json.Unmarshal([]byte(nextData), &livSpeakeasyNextData); err != nil {
		return pageProps, fmt.Errorf("error decoding __NEXT_DATA__ from %s: %w", pageUrl, err)
	}

	return livSpeakeasyNextData.Props.PageProps, nil
}

func formatLivSpeakeasyURL(scrapingUrl string, startDate time.Time, endDate time.Time) string {
	return fmt.Sprintf("%s&startDate=%s&endDate=%s", scrapingUrl, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

type LivSpeakeasyNextData struct {
	Props struct {
		PageProps LivSpeakeasyPageProps `json:"pageProps"`
	} `json:"props"`
}

type LivSpeakeasyPageProps struct {
	Events        []LivSpeakeasyEvent `json:"events"`
	EventsHasMore bool                `json:"eventsHasMore"`
}

type LivSpeakeasyEvent struct {
	UniqueId      string `json:"uniqueId"`
	Title         string `json:"title"`
	StartDateTime string `json:"startDateTime"`
	Flyer         struct {
		Key string `json:"key"`
	} `json:"flyer"`
	Venue struct {
		Name     string `json:"name"`
		City     string `json:"city"`
		Timezone string `json:"timezone"`
	} `json:"venue"`
}

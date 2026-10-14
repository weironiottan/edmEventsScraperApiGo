package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

/*
Update: 10/13/2026

LIV moved to a new site. The old admin-ajax pagination endpoint is legacy (its event URLs
301 redirect to the new site). The new events page is server rendered and contains
every upcoming event, with the data we need in data-liv-* attributes on each card,
so there's no pagination anymore.
----
*/

// Only the nightclub, LIV Beach (livlasvegasbeach) is a dayclub
const livNightclubVenueKey = "livlasvegas"

func scrapeLivForEdmEvents(scrapingUrl string) []EdmEvent {
	edmEvents := []EdmEvent{}

	htmlContent, err := getLivEventsPage(scrapingUrl)
	if err != nil {
		fmt.Println("Error while scraping LIV:", err)
		fmt.Println(len(edmEvents), "LIV edmEvents")
		return edmEvents
	}

	edmEvents = parseLivEventsPage(htmlContent, scrapingUrl)
	// Placeholder listings LIV posts before the lineup is announced, or for closed nights
	edmEvents = filterUnwantedArtists(edmEvents, []string{"special guest", "private event"})

	fmt.Println(len(edmEvents), "LIV edmEvents")
	fmt.Println("Scraping Completed for LIV!!!")
	return edmEvents
}

func getLivEventsPage(scrapingUrl string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, scrapingUrl, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("error sending request to %s: %w", scrapingUrl, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d from %s (content-type %q): %s",
			response.StatusCode, scrapingUrl, response.Header.Get("Content-Type"), truncate(string(body), 500))
	}

	return string(body), nil
}

func parseLivEventsPage(htmlContent string, pageUrl string) []EdmEvent {
	edmEvents := []EdmEvent{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		fmt.Println("Error while parsing LIV HTML:", err)
		return edmEvents
	}

	cards := doc.Find("[data-liv-event-id]")
	if cards.Length() == 0 {
		fmt.Println("Error while scraping LIV: no event cards found, the page layout may have changed")
		return edmEvents
	}

	seenEventIds := map[string]bool{}

	cards.Each(func(i int, selection *goquery.Selection) {
		venueKey, _ := selection.Attr("data-liv-venue-key")
		if venueKey != livNightclubVenueKey {
			return
		}

		// The page can show the same event in more than one section
		eventId, _ := selection.Attr("data-liv-event-id")
		if seenEventIds[eventId] {
			return
		}
		seenEventIds[eventId] = true

		artistName, _ := selection.Attr("data-liv-event-name")
		clubName, _ := selection.Attr("data-liv-venue-name")
		eventDate, _ := selection.Attr("data-liv-event-date")
		href, _ := selection.Attr("href")
		artistImageUrl, _ := selection.Find("img").Attr("src")

		formattedDate, err := formatDateFrom_YYYY_MM_DD_toRFC3339(eventDate)
		if err != nil {
			fmt.Println("Error while parsing the date:", err)
			return
		}

		isPastDate, err := isPastDate(formattedDate)
		if err != nil {
			fmt.Println("Error while parsing the date:", err)
			return
		}
		if isPastDate {
			return
		}

		edmEvent := EdmEvent{}
		edmEvent.Id = getGUID()
		edmEvent.ArtistName = strings.ToLower(strings.TrimSpace(artistName))
		edmEvent.ClubName = strings.ToLower(strings.TrimSpace(clubName))
		edmEvent.EventDate = formattedDate
		edmEvent.TicketUrl = resolveUrl(pageUrl, href)
		edmEvent.ArtistImageUrl = artistImageUrl
		edmEvents = append(edmEvents, edmEvent)
	})

	return edmEvents
}

// Event links on the page are relative (/las-vegas/event/...)
func resolveUrl(baseUrl string, href string) string {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

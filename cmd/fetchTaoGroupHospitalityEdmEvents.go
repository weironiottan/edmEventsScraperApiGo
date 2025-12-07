package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func scrapeTaoGroupHospitalityEdmEvents(scrappingUrl string) []EdmEvent {
	pageNumber := 1

	var taoGroupHospitalityEdmEvents TaoGroupHospitalityEdmEvents
	edmEvents := []EdmEvent{}

	for {
		taoGroupHospitalityUrl := formatPaginatedURL(scrappingUrl, pageNumber)
		fmt.Println("Visting ", taoGroupHospitalityUrl)
		response, err := getTaoGroupHospitalityEdmEvents(taoGroupHospitalityUrl)
		if err != nil {
			fmt.Println("Got at the end of the paginated response, breaking out of the loop")
			break
		}
		pageNumber++

		// Read the response body
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
		}

		err = json.Unmarshal(body, &taoGroupHospitalityEdmEvents)
		if err != nil {
			fmt.Println("Error:", err)
		}

		defer response.Body.Close()

		for _, taoGroupHospitalityEvent := range taoGroupHospitalityEdmEvents {
			// Skip if venue array is empty
			if len(taoGroupHospitalityEvent.ACF.EventVenue) == 0 {
				continue
			}

			edmEvent := EdmEvent{}
			edmEvent.Id = getGUID()
			edmEvent.ArtistName = strings.ToLower(taoGroupHospitalityEvent.ACF.EventTitle.DisplayTitle)
			formattedClubName := filterOutLasVegasFromTitle(taoGroupHospitalityEvent.ACF.EventVenue[0].PostTitle)
			edmEvent.ClubName = formattedClubName
			formattedDate := filterOutTimeFromDate(taoGroupHospitalityEvent.ACF.EventStartDate)
			formattedDate, err := formatDateFrom_MM_DD_YYYY_toRFC3339(formattedDate)

			if err != nil {
				fmt.Println("Error while parsing the date:", err)
				continue
			}

			edmEvent.EventDate = formattedDate
			edmEvent.TicketUrl = taoGroupHospitalityEvent.Link

			isPastDate, err := isPastDate(formattedDate)
			if err != nil {
				fmt.Println("Error while parsing the date:", err)
				continue
			}

			if !isPastDate {
				edmEvents = append(edmEvents, edmEvent)
			}

		}
	}
	fmt.Println(len(edmEvents), "NOT filtered events")
	edmEvents = filterUnwantedEvents(edmEvents, []string{"lavo italian restaurant las vegas", "lavo italian restaurant"})
	fmt.Println(len(edmEvents), "filtered events")

	fmt.Println("Scraping Completed for Tao Group Hospitality")
	return edmEvents
}

func getTaoGroupHospitalityEdmEvents(url string) (*http.Response, error) {
	client := &http.Client{}
	request, _ := http.NewRequest(http.MethodGet, url, nil)

	response, err := client.Do(request)

	if err != nil {
		return response, fmt.Errorf("error sending request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return response, fmt.Errorf("status recieved was not a 200")
	}

	return response, nil

}

func filterOutTimeFromDate(eventDateTime string) string {
	eventDate := strings.Split(eventDateTime, " ")
	return eventDate[0]
}

func filterOutLasVegasFromTitle(venueTitle string) string {
	venueTitle = strings.ToLower(venueTitle)
	regexPattern := `\s-\slas vegas`
	re := regexp.MustCompile(regexPattern)
	formattedVenueTitle := re.ReplaceAllString(venueTitle, "")
	formattedVenueTitle = strings.TrimSpace(formattedVenueTitle)
	return formattedVenueTitle
}

func formatPaginatedURL(scrappingUrl string, pageNumber int) string {
	formattedURL := fmt.Sprintf(scrappingUrl+"page=%v&per_page=300", pageNumber)
	return formattedURL
}

// TaoGroupHospitalityEdmEvents represents the event data from Tao Group Hospitality
type TaoGroupHospitalityEdmEvents []struct {
	ID              int         `json:"id"`
	Date            string      `json:"date"`
	DateGMT         string      `json:"date_gmt"`
	GUID            GUID        `json:"guid"`
	Modified        string      `json:"modified"`
	ModifiedGMT     string      `json:"modified_gmt"`
	Slug            string      `json:"slug"`
	Status          string      `json:"status"`
	Type            string      `json:"type"`
	Link            string      `json:"link"`
	Title           Title       `json:"title"`
	Content         Content     `json:"content"`
	FeaturedMedia   int         `json:"featured_media"`
	Template        string      `json:"template"`
	EventType       []int       `json:"event_type"`
	EventHoliday    []int       `json:"event_holiday"`
	EventVenue      []int       `json:"event_venue"`
	EventCity       []int       `json:"event_city"`
	EventArtist     []int       `json:"event_artist"`
	EventMisc       []int       `json:"event_misc"`
	EventNightDay   []int       `json:"event_night_day"`
	EventRestaurant []int       `json:"event_restaurant"`
	ClassList       []string    `json:"class_list"`
	ACF             ACF         `json:"acf"`
	Links           []Link      `json:"links"`
	ContentfulID    string      `json:"contentful_id"`
	StartEpoch      string      `json:"start_epoch"`
	StartYMD        string      `json:"start_ymd"`
	EndEpoch        string      `json:"end_epoch"`
	YoastHead       string      `json:"yoast_head,omitempty" bson:"-"`
	YoastHeadJSON   interface{} `json:"yoast_head_json,omitempty" bson:"-"`
	FeaturedImgURL  string      `json:"featured_img_url"`
	APILinks        APILinks    `json:"_links"`
}

// GUID represents the globally unique identifier structure
type GUID struct {
	Rendered string `json:"rendered"`
}

// Title represents a title with rendered content
type Title struct {
	Rendered string `json:"rendered"`
}

// Content represents content with rendered HTML and protection status
type Content struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected"`
}

// ACF represents the Advanced Custom Fields data
type ACF struct {
	EventTitle       EventTitle   `json:"event_title"`
	EventStartDate   string       `json:"event_start_date"`
	EventEndDate     string       `json:"event_end_date"`
	TimeZone         string       `json:"time_zone"`
	EventDescription string       `json:"event_description"`
	EventVenue       []EventVenue `json:"event_venue"`
}

// EventTitle represents the title information for an event
type EventTitle struct {
	DisplayTitle string `json:"display_title"`
	Badge        string `json:"badge"`
}

// EventVenue represents venue information
type EventVenue struct {
	ID                  int      `json:"ID"`
	PostAuthor          string   `json:"post_author"`
	PostDate            string   `json:"post_date"`
	PostDateGMT         string   `json:"post_date_gmt"`
	PostContent         string   `json:"post_content"`
	PostTitle           string   `json:"post_title"`
	PostExcerpt         string   `json:"post_excerpt"`
	PostStatus          string   `json:"post_status"`
	CommentStatus       string   `json:"comment_status"`
	PingStatus          string   `json:"ping_status"`
	PostPassword        string   `json:"post_password"`
	PostName            string   `json:"post_name"`
	ToPing              string   `json:"to_ping"`
	Pinged              string   `json:"pinged"`
	PostModified        string   `json:"post_modified"`
	PostModifiedGMT     string   `json:"post_modified_gmt"`
	PostContentFiltered string   `json:"post_content_filtered"`
	PostParent          int      `json:"post_parent"`
	GUID                string   `json:"guid"`
	MenuOrder           int      `json:"menu_order"`
	PostType            string   `json:"post_type"`
	PostMimeType        string   `json:"post_mime_type"`
	CommentCount        string   `json:"comment_count"`
	Filter              string   `json:"filter"`
	VenueACF            VenueACF `json:"acf"`
}

// VenueACF represents the Advanced Custom Fields for a venue
type VenueACF struct {
	ApplyRedirect bool     `json:"apply_redirect"`
	RedirectURL   string   `json:"redirect_url"`
	VenueName     string   `json:"venue_name"`
	VenueTheme    string   `json:"venue_theme"`
	Navbar        Navbar   `json:"navbar"`
	Location      Location `json:"location"`
	Contact       Contact  `json:"contact"`
	// Simplified for brevity - add more fields as needed
}

// Navbar represents the navigation bar structure
type Navbar struct {
	LogoType  string      `json:"logo_type"`
	LogoImage bool        `json:"logo_image"`
	LogoSVG   string      `json:"logo_svg"`
	Links     interface{} `json:"links"`
	MainCTA   interface{} `json:"main_cta"`
	Options   string      `json:"options"`
}

// NavLink represents a navigation link with different possible layouts
type NavLink struct {
	ACFLayout    string    `json:"acf_fc_layout"`
	SectionTitle string    `json:"section_title"`
	Options      string    `json:"options"`
	Link         *LinkInfo `json:"link,omitempty"`
	Trigger      string    `json:"trigger,omitempty"`
	Links        []struct {
		Link LinkInfo `json:"link"`
	} `json:"links,omitempty"`
}

// LinkInfo represents link information
type LinkInfo struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Target string `json:"target"`
}

// Location represents a physical location
type Location struct {
	Address string `json:"address"`
	Country string `json:"country"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}

// Contact represents contact information
type Contact struct {
	PhoneNumber string `json:"phone_number"`
}

// Link represents a link with title, URL, and target
type Link struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Target string `json:"target"`
}

// OgImage represents Open Graph image metadata
// Only keeping this for the ToEdmEvent() method
type OgImage struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
	Type   string `json:"type"`
}

// APILinks represents the _links section in the API response
type APILinks struct {
	Self          []APILink           `json:"self"`
	Collection    []APILink           `json:"collection"`
	About         []APILink           `json:"about"`
	FeaturedMedia []APILinkEmbeddable `json:"wp:featuredmedia"`
	Attachment    []APILink           `json:"wp:attachment"`
	Terms         []APILinkTaxonomy   `json:"wp:term"`
	Curies        []APICurie          `json:"curies"`
}

// APILink represents a basic API link
type APILink struct {
	Href        string              `json:"href"`
	TargetHints *APILinkTargetHints `json:"targetHints,omitempty"`
}

// APILinkTargetHints represents hints about a link target
type APILinkTargetHints struct {
	Allow []string `json:"allow"`
}

// APILinkEmbeddable represents an API link that can be embedded
type APILinkEmbeddable struct {
	Embeddable bool   `json:"embeddable"`
	Href       string `json:"href"`
}

// APILinkTaxonomy represents a taxonomy link
type APILinkTaxonomy struct {
	Taxonomy   string `json:"taxonomy"`
	Embeddable bool   `json:"embeddable"`
	Href       string `json:"href"`
}

// APICurie represents a CURIE (Compact URI)
type APICurie struct {
	Name      string `json:"name"`
	Href      string `json:"href"`
	Templated bool   `json:"templated"`
}

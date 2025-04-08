package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func (app *application) fetchTaoGroupHospitalityEdmEvents(w http.ResponseWriter, r *http.Request) {
	edmEvents := scrapeTaoGroupHospitalityEdmEvents()
	err := app.writeJSON(w, http.StatusOK, edmEvents, nil)

	if err != nil {
		app.logger.Print(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}

func scrapeTaoGroupHospitalityEdmEvents() []EdmEvent {
	pageNumber := 1

	var taoGroupHospitalityEdmEvents TaoGroupHospitalityEdmEvents
	edmEvents := []EdmEvent{}

	for {
		taoGroupHospitalityUrl := fmt.Sprintf("https://taogroup.com/wp-json/wp/v2/events?event_city%%5B%%5D=81&filter%%5Bmeta_compare%%5D=%%3E%%3D&filter%%5Bmeta_key%%5D=event_start_date&filter%%5Bmeta_value%%5D=1720422000000&filter%%5Border%%5D=asc&filter%%5Borderby%%5D=meta_value&page=%v&per_page=300", pageNumber)
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

			edmEvent := EdmEvent{}
			edmEvent.Id = getGUID()
			edmEvent.ArtistName = strings.ToLower(taoGroupHospitalityEvent.ACF.EventTitle.DisplayTitle)
			formattedClubName := filterOutLasVegasFromTitle(taoGroupHospitalityEvent.ACF.EventVenue[0].PostTitle)
			edmEvent.ClubName = formattedClubName
			formattedDate := filterOutTimeFromDate(taoGroupHospitalityEvent.ACF.EventStartDate)
			formattedDate, err := formatDateFrom_MM_DD_YYYY_toRFC3339(formattedDate)

			if err != nil {
				fmt.Println("Error while parsing the date:", err)
			}

			edmEvent.EventDate = formattedDate
			edmEvent.TicketUrl = taoGroupHospitalityEvent.Link

			isPastDate, err := isPastDate(formattedDate)
			if err != nil {
				fmt.Println("Error while parsing the date:", err)
			}

			if !isPastDate {
				edmEvents = append(edmEvents, edmEvent)
			}

		}
	}
	fmt.Println(len(edmEvents), "NOT filtered events")
	edmEvents = filterUnwantedEvents(edmEvents, []string{"lavo italian restaurant las vegas"})
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

//type TaoGroupHospitalityEdmEvents []struct {
//	Links struct {
//		About []struct {
//			Href string `json:"href"`
//		} `json:"about"`
//		Collection []struct {
//			Href string `json:"href"`
//		} `json:"collection"`
//		Curies []struct {
//			Href      string `json:"href"`
//			Name      string `json:"name"`
//			Templated bool   `json:"templated"`
//		} `json:"curies"`
//		Self []struct {
//			Href string `json:"href"`
//		} `json:"self"`
//		Wp_attachment []struct {
//			Href string `json:"href"`
//		} `json:"wp:attachment"`
//		Wp_featuredmedia []struct {
//			Embeddable bool   `json:"embeddable"`
//			Href       string `json:"href"`
//		} `json:"wp:featuredmedia"`
//		Wp_term []struct {
//			Embeddable bool   `json:"embeddable"`
//			Href       string `json:"href"`
//			Taxonomy   string `json:"taxonomy"`
//		} `json:"wp:term"`
//	} `json:"_links"`
//	Acf struct {
//		ContentfulID     string `json:"contentful_id"`
//		EndEpoch         string `json:"end_epoch"`
//		EventDescription string `json:"event_description"`
//		EventEndDate     string `json:"event_end_date"`
//		EventStartDate   string `json:"event_start_date"`
//		EventTitle       struct {
//			Badge        string `json:"badge"`
//			DisplayTitle string `json:"display_title"`
//		} `json:"event_title"`
//		EventVenue []struct {
//			ID                  int64  `json:"ID"`
//			CommentCount        string `json:"comment_count"`
//			CommentStatus       string `json:"comment_status"`
//			Filter              string `json:"filter"`
//			GUID                string `json:"guid"`
//			MenuOrder           int64  `json:"menu_order"`
//			PingStatus          string `json:"ping_status"`
//			Pinged              string `json:"pinged"`
//			PostAuthor          string `json:"post_author"`
//			PostContent         string `json:"post_content"`
//			PostContentFiltered string `json:"post_content_filtered"`
//			PostDate            string `json:"post_date"`
//			PostDateGmt         string `json:"post_date_gmt"`
//			PostExcerpt         string `json:"post_excerpt"`
//			PostMimeType        string `json:"post_mime_type"`
//			PostModified        string `json:"post_modified"`
//			PostModifiedGmt     string `json:"post_modified_gmt"`
//			PostName            string `json:"post_name"`
//			PostParent          int64  `json:"post_parent"`
//			PostPassword        string `json:"post_password"`
//			PostStatus          string `json:"post_status"`
//			PostTitle           string `json:"post_title"`
//			PostType            string `json:"post_type"`
//			ToPing              string `json:"to_ping"`
//		} `json:"event_venue"`
//		Links []struct {
//			Link struct {
//				Target interface{} `json:"target"`
//				Title  string      `json:"title"`
//				URL    string      `json:"url"`
//			} `json:"link"`
//		} `json:"links"`
//		StartEpoch string `json:"start_epoch"`
//		StartYmd   string `json:"start_ymd"`
//		TimeZone   string `json:"time_zone"`
//	} `json:"acf"`
//	Content struct {
//		Protected bool   `json:"protected"`
//		Rendered  string `json:"rendered"`
//	} `json:"content"`
//	Date            string        `json:"date"`
//	DateGmt         string        `json:"date_gmt"`
//	EventArtist     []interface{} `json:"event_artist"`
//	EventCity       []int64       `json:"event_city"`
//	EventHoliday    []int64       `json:"event_holiday"`
//	EventMisc       []interface{} `json:"event_misc"`
//	EventNightDay   []interface{} `json:"event_night_day"`
//	EventRestaurant []interface{} `json:"event_restaurant"`
//	EventType       []interface{} `json:"event_type"`
//	EventVenue      []int64       `json:"event_venue"`
//	FeaturedImgURL  string        `json:"featured_img_url"`
//	FeaturedMedia   int64         `json:"featured_media"`
//	GUID            struct {
//		Rendered string `json:"rendered"`
//	} `json:"guid"`
//	ID          int64  `json:"id"`
//	Link        string `json:"link"`
//	Modified    string `json:"modified"`
//	ModifiedGmt string `json:"modified_gmt"`
//	Slug        string `json:"slug"`
//	Status      string `json:"status"`
//	Template    string `json:"template"`
//	Title       struct {
//		Rendered string `json:"rendered"`
//	} `json:"title"`
//	Type          string `json:"type"`
//	YoastHead     string `json:"yoast_head"`
//	YoastHeadJSON struct {
//		ArticleModifiedTime string `json:"article_modified_time"`
//		Canonical           string `json:"canonical"`
//		OgImage             []struct {
//			Height int64  `json:"height"`
//			Type   string `json:"type"`
//			URL    string `json:"url"`
//			Width  int64  `json:"width"`
//		} `json:"og_image"`
//		OgLocale   string `json:"og_locale"`
//		OgSiteName string `json:"og_site_name"`
//		OgTitle    string `json:"og_title"`
//		OgType     string `json:"og_type"`
//		OgURL      string `json:"og_url"`
//		Robots     struct {
//			Follow            string `json:"follow"`
//			Index             string `json:"index"`
//			Max_image_preview string `json:"max-image-preview"`
//			Max_snippet       string `json:"max-snippet"`
//			Max_video_preview string `json:"max-video-preview"`
//		} `json:"robots"`
//		Schema struct {
//			Context string `json:"@context"`
//			Graph   []struct {
//				ID         string `json:"@id"`
//				Type       string `json:"@type"`
//				Breadcrumb struct {
//					ID string `json:"@id"`
//				} `json:"breadcrumb"`
//				DateModified  string `json:"dateModified"`
//				DatePublished string `json:"datePublished"`
//				Description   string `json:"description"`
//				Image         struct {
//					ID string `json:"@id"`
//				} `json:"image"`
//				InLanguage string `json:"inLanguage"`
//				IsPartOf   struct {
//					ID string `json:"@id"`
//				} `json:"isPartOf"`
//				ItemListElement []struct {
//					Type     string `json:"@type"`
//					Item     string `json:"item"`
//					Name     string `json:"name"`
//					Position int64  `json:"position"`
//				} `json:"itemListElement"`
//				Logo struct {
//					ID         string `json:"@id"`
//					Type       string `json:"@type"`
//					Caption    string `json:"caption"`
//					ContentURL string `json:"contentUrl"`
//					Height     int64  `json:"height"`
//					InLanguage string `json:"inLanguage"`
//					URL        string `json:"url"`
//					Width      int64  `json:"width"`
//				} `json:"logo"`
//				Name            string `json:"name"`
//				PotentialAction []struct {
//					Type string `json:"@type"`
//					//Query_input string      `json:"query-input"`
//					Target interface{} `json:"target"`
//				} `json:"potentialAction"`
//				Publisher struct {
//					ID string `json:"@id"`
//				} `json:"publisher"`
//				SameAs []string `json:"sameAs"`
//				URL    string   `json:"url"`
//			} `json:"@graph"`
//		} `json:"schema"`
//		Title       string `json:"title"`
//		TwitterCard string `json:"twitter_card"`
//	} `json:"yoast_head_json"`
//}

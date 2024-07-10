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
	hasEventItems := true
	pageNumber := 1

	var taoGroupHospitalityEdmEvents TaoGroupHospitalityEdmEvents
	edmEvents := []EdmEvent{}

	for hasEventItems {
		taoGroupHospitalityUrl := fmt.Sprintf("https://taogroup.com/wp-json/wp/v2/events?event_city%%5B%%5D=81&filter%%5Bmeta_compare%%5D=%%3E%%3D&filter%%5Bmeta_key%%5D=event_start_date&filter%%5Bmeta_value%%5D=1720422000000&filter%%5Border%%5D=asc&filter%%5Borderby%%5D=meta_value&page=%v&per_page=500", pageNumber)
		fmt.Println("Visting ", taoGroupHospitalityUrl)
		response, err := getTaoGroupHospitalityEdmEvents(taoGroupHospitalityUrl)
		if err != nil {
			hasEventItems = false
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
			edmEvent.ArtistName = strings.ToLower(taoGroupHospitalityEvent.Acf.EventTitle.DisplayTitle)
			formattedClubName := filterOutLasVegasFromTitle(taoGroupHospitalityEvent.Acf.EventVenue[0].PostTitle)
			edmEvent.ClubName = formattedClubName
			formattedDate := filterOutTimeFromDate(taoGroupHospitalityEvent.Acf.EventStartDate)
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

type TaoGroupHospitalityEdmEvents []struct {
	Links struct {
		About []struct {
			Href string `json:"href"`
		} `json:"about"`
		Collection []struct {
			Href string `json:"href"`
		} `json:"collection"`
		Curies []struct {
			Href      string `json:"href"`
			Name      string `json:"name"`
			Templated bool   `json:"templated"`
		} `json:"curies"`
		Self []struct {
			Href string `json:"href"`
		} `json:"self"`
		Wp_attachment []struct {
			Href string `json:"href"`
		} `json:"wp:attachment"`
		Wp_featuredmedia []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
		} `json:"wp:featuredmedia"`
		Wp_term []struct {
			Embeddable bool   `json:"embeddable"`
			Href       string `json:"href"`
			Taxonomy   string `json:"taxonomy"`
		} `json:"wp:term"`
	} `json:"_links"`
	Acf struct {
		ContentfulID     string `json:"contentful_id"`
		EndEpoch         string `json:"end_epoch"`
		EventDescription string `json:"event_description"`
		EventEndDate     string `json:"event_end_date"`
		EventStartDate   string `json:"event_start_date"`
		EventTitle       struct {
			Badge        string `json:"badge"`
			DisplayTitle string `json:"display_title"`
		} `json:"event_title"`
		EventVenue []struct {
			ID                  int64  `json:"ID"`
			CommentCount        string `json:"comment_count"`
			CommentStatus       string `json:"comment_status"`
			Filter              string `json:"filter"`
			GUID                string `json:"guid"`
			MenuOrder           int64  `json:"menu_order"`
			PingStatus          string `json:"ping_status"`
			Pinged              string `json:"pinged"`
			PostAuthor          string `json:"post_author"`
			PostContent         string `json:"post_content"`
			PostContentFiltered string `json:"post_content_filtered"`
			PostDate            string `json:"post_date"`
			PostDateGmt         string `json:"post_date_gmt"`
			PostExcerpt         string `json:"post_excerpt"`
			PostMimeType        string `json:"post_mime_type"`
			PostModified        string `json:"post_modified"`
			PostModifiedGmt     string `json:"post_modified_gmt"`
			PostName            string `json:"post_name"`
			PostParent          int64  `json:"post_parent"`
			PostPassword        string `json:"post_password"`
			PostStatus          string `json:"post_status"`
			PostTitle           string `json:"post_title"`
			PostType            string `json:"post_type"`
			ToPing              string `json:"to_ping"`
		} `json:"event_venue"`
		Links []struct {
			Link struct {
				Target interface{} `json:"target"`
				Title  string      `json:"title"`
				URL    string      `json:"url"`
			} `json:"link"`
		} `json:"links"`
		StartEpoch string `json:"start_epoch"`
		StartYmd   string `json:"start_ymd"`
		TimeZone   string `json:"time_zone"`
	} `json:"acf"`
	Content struct {
		Protected bool   `json:"protected"`
		Rendered  string `json:"rendered"`
	} `json:"content"`
	Date            string        `json:"date"`
	DateGmt         string        `json:"date_gmt"`
	EventArtist     []interface{} `json:"event_artist"`
	EventCity       []int64       `json:"event_city"`
	EventHoliday    []int64       `json:"event_holiday"`
	EventMisc       []interface{} `json:"event_misc"`
	EventNightDay   []interface{} `json:"event_night_day"`
	EventRestaurant []interface{} `json:"event_restaurant"`
	EventType       []interface{} `json:"event_type"`
	EventVenue      []int64       `json:"event_venue"`
	FeaturedImgURL  string        `json:"featured_img_url"`
	FeaturedMedia   int64         `json:"featured_media"`
	GUID            struct {
		Rendered string `json:"rendered"`
	} `json:"guid"`
	ID          int64  `json:"id"`
	Link        string `json:"link"`
	Modified    string `json:"modified"`
	ModifiedGmt string `json:"modified_gmt"`
	Slug        string `json:"slug"`
	Status      string `json:"status"`
	Template    string `json:"template"`
	Title       struct {
		Rendered string `json:"rendered"`
	} `json:"title"`
	Type          string `json:"type"`
	YoastHead     string `json:"yoast_head"`
	YoastHeadJSON struct {
		ArticleModifiedTime string `json:"article_modified_time"`
		Canonical           string `json:"canonical"`
		OgImage             []struct {
			Height int64  `json:"height"`
			Type   string `json:"type"`
			URL    string `json:"url"`
			Width  int64  `json:"width"`
		} `json:"og_image"`
		OgLocale   string `json:"og_locale"`
		OgSiteName string `json:"og_site_name"`
		OgTitle    string `json:"og_title"`
		OgType     string `json:"og_type"`
		OgURL      string `json:"og_url"`
		Robots     struct {
			Follow            string `json:"follow"`
			Index             string `json:"index"`
			Max_image_preview string `json:"max-image-preview"`
			Max_snippet       string `json:"max-snippet"`
			Max_video_preview string `json:"max-video-preview"`
		} `json:"robots"`
		Schema struct {
			Context string `json:"@context"`
			Graph   []struct {
				ID         string `json:"@id"`
				Type       string `json:"@type"`
				Breadcrumb struct {
					ID string `json:"@id"`
				} `json:"breadcrumb"`
				DateModified  string `json:"dateModified"`
				DatePublished string `json:"datePublished"`
				Description   string `json:"description"`
				Image         struct {
					ID string `json:"@id"`
				} `json:"image"`
				InLanguage string `json:"inLanguage"`
				IsPartOf   struct {
					ID string `json:"@id"`
				} `json:"isPartOf"`
				ItemListElement []struct {
					Type     string `json:"@type"`
					Item     string `json:"item"`
					Name     string `json:"name"`
					Position int64  `json:"position"`
				} `json:"itemListElement"`
				Logo struct {
					ID         string `json:"@id"`
					Type       string `json:"@type"`
					Caption    string `json:"caption"`
					ContentURL string `json:"contentUrl"`
					Height     int64  `json:"height"`
					InLanguage string `json:"inLanguage"`
					URL        string `json:"url"`
					Width      int64  `json:"width"`
				} `json:"logo"`
				Name            string `json:"name"`
				PotentialAction []struct {
					Type        string      `json:"@type"`
					Query_input string      `json:"query-input"`
					Target      interface{} `json:"target"`
				} `json:"potentialAction"`
				Publisher struct {
					ID string `json:"@id"`
				} `json:"publisher"`
				SameAs []string `json:"sameAs"`
				URL    string   `json:"url"`
			} `json:"@graph"`
		} `json:"schema"`
		Title       string `json:"title"`
		TwitterCard string `json:"twitter_card"`
	} `json:"yoast_head_json"`
}

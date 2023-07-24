package main

type EdmEvent struct {
	ClubName       string `json:"ClubName,omitempty"`
	ArtistName     string `json:"ArtistName,omitempty"`
	EventDate      string `json:"EventDate,omitempty"`
	TicketUrl      string `json:"TicketUrl,omitempty"`
	ArtistImageUrl string `json:"ArtistImageUrl,omitempty"`
}

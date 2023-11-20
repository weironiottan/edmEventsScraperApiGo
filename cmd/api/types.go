package main

type EdmEvent struct {
	Id             string `json:"Id,omitempty"`
	ClubName       string `json:"ClubName,omitempty"`
	ArtistName     string `json:"ArtistName,omitempty"`
	EventDate      string `json:"EventDate,omitempty"`
	TicketUrl      string `json:"TicketUrl,omitempty"`
	ArtistImageUrl string `json:"ArtistImageUrl,omitempty"`
}

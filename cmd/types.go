package main

type EdmEvent struct {
	Id             string `json:"id,omitempty"`
	ClubName       string `json:"clubname,omitempty"`
	ArtistName     string `json:"artistname,omitempty"`
	EventDate      string `json:"eventdate,omitempty"`
	TicketUrl      string `json:"ticketurl,omitempty"`
	ArtistImageUrl string `json:"artistimageurl,omitempty"`
}

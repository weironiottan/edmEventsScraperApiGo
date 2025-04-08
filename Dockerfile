FROM golang:1.23
LABEL authors="christiangabrielsson"

WORKDIR /app

COPY go.mod ./

# Uncomment the next line if you have a go.sum file
# COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o /edmEventsScraperJob ./cmd

CMD ["/edmEventsScraperJob"]
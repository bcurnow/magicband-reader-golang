package rfidsecuritysvc

import (
	"time"
)

type Media struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"desc"`
}

type Permission struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"desc"`
}

type Guest struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Sound     *Sound `json:"sound"`
	Color     *Color `json:"color"`
}

type Sound struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	LastUpdateTimestamp time.Time `json:"last_update_timestamp"`
	Content             string    `json:"content"`
}

type Color struct {
	Int  int    `json:"int"`
	Hex  string `json:hex`
	Html string `json:html`
}

type MediaConfig struct {
	Media      *Media      `json:"media"`
	Permission *Permission `json:"permission"`
	Guest      *Guest      `json:"guest"`
	Sound      *Sound      `json:"sound"`
	Color      *Color      `json:"color"`
}

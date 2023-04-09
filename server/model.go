package main

import (
	"time"
)

type Event struct {
	Id         string     `json:"id" db:"id"`
	Title      string     `json:"title" db:"title"`
	Start      time.Time  `json:"start" db:"start"`
	End        time.Time  `json:"end" db:"end"`
	Attendees  []string   `json:"attendees"`
	Created    time.Time  `json:"created" db:"created"`
	Owner      string     `json:"owner" db:"owner"`
	Channel    *string    `json:"channel" db:"channel"`
	Processed  *time.Time `json:"-" db:"processed"`
	Recurrent  bool       `json:"-" db:"recurrent"`
	Recurrence string     `json:"recurrence" db:"recurrence"`
	Color      *string    `json:"color" db:"color"`
}

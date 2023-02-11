package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type RecurrenceItem []int

func (r *RecurrenceItem) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &r)
		return nil
	case string:
		json.Unmarshal([]byte(v), &r)
		return nil
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %T", v))
	}
}

func (r *RecurrenceItem) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type Event struct {
	Id         string          `json:"id" db:"id"`
	Title      string          `json:"title" db:"title"`
	Start      time.Time       `json:"start" db:"start"`
	End        time.Time       `json:"end" db:"end"`
	Attendees  []string        `json:"attendees"`
	Created    time.Time       `json:"created" db:"created"`
	Owner      string          `json:"owner" db:"owner"`
	Channel    *string         `json:"channel" db:"channel"`
	Processed  *time.Time      `json:"-" db:"processed"`
	Recurrent  bool            `json:"-" db:"recurrent"`
	Recurrence *RecurrenceItem `json:"recurrence" db:"recurrence"`
}

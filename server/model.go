package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventVisibility string

const (
	VisibilityPrivate EventVisibility = "private"
	VisibilityChannel EventVisibility = "channel"
	VisibilityTeam    EventVisibility = "team"
)

// UnmarshalJSON customizes the JSON unmarshaling of SubscriptionStatus.
func (e *EventVisibility) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case string(VisibilityPrivate), string(VisibilityChannel), string(VisibilityTeam):
		*e = EventVisibility(s)
		return nil
	default:
		return fmt.Errorf("invalid SubscriptionStatus: %s", s)
	}
}

func (s *EventVisibility) Scan(value interface{}) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("SubscriptionStatus must be a string or []byte, got %T", value)
	}

	switch strValue {
	case string(VisibilityChannel), string(VisibilityPrivate), string(VisibilityTeam):
		*s = EventVisibility(strValue)
		return nil
	default:
		return fmt.Errorf("invalid SubscriptionStatus: %s", strValue)
	}
}

type Event struct {
	Id          string          `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description" db:"description"`
	Start       time.Time       `json:"start" db:"dt_start"`
	End         time.Time       `json:"end" db:"dt_end"`
	Attendees   []string        `json:"attendees"`
	Created     time.Time       `json:"created" db:"created"`
	Owner       string          `json:"owner" db:"owner"`
	Team        string          `json:"team" db:"team"`
	Channel     *string         `json:"channel" db:"channel"`
	Processed   *time.Time      `json:"-" db:"processed"`
	Recurrent   bool            `json:"-" db:"recurrent"`
	Recurrence  string          `json:"recurrence" db:"recurrence"`
	Color       *string         `json:"color" db:"color"`
	Visibility  EventVisibility `json:"visibility" db:"visibility"`
	Alert       string          `json:"alert" db:"alert"`
	AlertTime   *time.Time      `json:"alertTime" db:"alert_time"`
}

type UserSettings struct {
	BusinessStartTime     string `json:"businessStartTime"`
	BusinessEndTime       string `json:"businessEndTime"`
	IsOpenCalendarLeftBar bool   `json:"isOpenCalendarLeftBar" db:"is_open_calendar_left_bar"`
	FirstDayOfWeek        int    `json:"firstDayOfWeek" db:"first_day_of_week"`
	BusinessDays          []int  `json:"businessDays"`
	HideNonWorkingDays    bool   `json:"hideNonWorkingDays" db:"hide_non_working_days"`
}

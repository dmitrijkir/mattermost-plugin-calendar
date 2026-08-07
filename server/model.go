package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventVisibility string
type EventAlert string
type EventType string
type EventMention string

const (
	EventAlertNone            EventAlert = ""
	EventAlertAtStartTime     EventAlert = "at_start_time"
	EventAlert5MinutesBefore  EventAlert = "5_minutes_before"
	EventAlert15MinutesBefore EventAlert = "15_minutes_before"
	EventAlert30MinutesBefore EventAlert = "30_minutes_before"
	EventAlert1HourBefore     EventAlert = "1_hour_before"
	EventAlert2HoursBefore    EventAlert = "2_hours_before"
	EventAlert1DayBefore      EventAlert = "1_day_before"
	EventAlert2DaysBefore     EventAlert = "2_days_before"
	EventAlert1WeekBefore     EventAlert = "1_week_before"

	VisibilityPrivate EventVisibility = "private"
	VisibilityChannel EventVisibility = "channel"
	VisibilityTeam    EventVisibility = "team"

	EventTypeCall    EventType = "call"
	EventTypeMeeting EventType = "event"

	// Stored without the leading "@" so the DB holds a plain token, the same
	// way alerts store "5_minutes_before" rather than display text.
	MentionNone    EventMention = ""
	MentionHere    EventMention = "here"
	MentionChannel EventMention = "channel"
	MentionAll     EventMention = "all"
)

var EventAlertDurationMap = map[EventAlert]time.Duration{
	EventAlertNone:            time.Duration(0),
	EventAlertAtStartTime:     time.Duration(0),
	EventAlert5MinutesBefore:  5 * time.Minute,
	EventAlert15MinutesBefore: 15 * time.Minute,
	EventAlert30MinutesBefore: 30 * time.Minute,
	EventAlert1HourBefore:     time.Hour,
	EventAlert2HoursBefore:    2 * time.Hour,
	EventAlert1DayBefore:      24 * time.Hour,
	EventAlert2DaysBefore:     2 * 24 * time.Hour,
	EventAlert1WeekBefore:     7 * 24 * time.Hour,
}

var EventAlertTitleMap = map[EventAlert]string{
	EventAlertNone:            "None",
	EventAlertAtStartTime:     "At start time",
	EventAlert5MinutesBefore:  "5 minutes before",
	EventAlert15MinutesBefore: "15 minutes before",
	EventAlert30MinutesBefore: "30 minutes before",
	EventAlert1HourBefore:     "1 hour before",
	EventAlert2HoursBefore:    "2 hours before",
	EventAlert1DayBefore:      "1 day before",
	EventAlert2DaysBefore:     "2 days before",
	EventAlert1WeekBefore:     "1 week before",
}

// UnmarshalJSON custom EventAlert unmarshaling
func (e *EventAlert) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case string(EventAlertNone), string(EventAlertAtStartTime), string(EventAlert5MinutesBefore), string(EventAlert15MinutesBefore), string(EventAlert30MinutesBefore), string(EventAlert1HourBefore), string(EventAlert2HoursBefore), string(EventAlert1DayBefore), string(EventAlert2DaysBefore), string(EventAlert1WeekBefore):
		*e = EventAlert(s)
		return nil
	default:
		return fmt.Errorf("invalid Alert set: %s", s)
	}
}

func (e *EventAlert) Scan(value interface{}) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("EventAlert must be a string or []byte, got %T", value)
	}

	switch strValue {
	case string(EventAlertNone), string(EventAlertAtStartTime), string(EventAlert5MinutesBefore), string(EventAlert15MinutesBefore), string(EventAlert30MinutesBefore), string(EventAlert1HourBefore), string(EventAlert2HoursBefore), string(EventAlert1DayBefore), string(EventAlert2DaysBefore), string(EventAlert1WeekBefore):
		*e = EventAlert(strValue)
		return nil
	default:
		return fmt.Errorf("invalid Alert set: %s", strValue)
	}
}

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

func (e *EventVisibility) Scan(value interface{}) error {
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
		*e = EventVisibility(strValue)
		return nil
	default:
		return fmt.Errorf("invalid SubscriptionStatus: %s", strValue)
	}
}

// UnmarshalJSON customizes the JSON unmarshaling of EventType.
func (e *EventType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case string(EventTypeCall), string(EventTypeMeeting):
		*e = EventType(s)
		return nil
	case "":
		// left for the handler to default, same as an omitted field
		*e = EventType(s)
		return nil
	default:
		return fmt.Errorf("invalid EventType: %s", s)
	}
}

func (e *EventType) Scan(value interface{}) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("EventType must be a string or []byte, got %T", value)
	}

	switch strValue {
	case string(EventTypeCall), string(EventTypeMeeting):
		*e = EventType(strValue)
		return nil
	case "":
		// rows written before the column had a default
		*e = EventTypeCall
		return nil
	default:
		return fmt.Errorf("invalid EventType: %s", strValue)
	}
}

// UnmarshalJSON customizes the JSON unmarshaling of EventMention.
func (e *EventMention) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch s {
	case string(MentionNone), string(MentionHere), string(MentionChannel), string(MentionAll):
		*e = EventMention(s)
		return nil
	default:
		return fmt.Errorf("invalid EventMention: %s", s)
	}
}

func (e *EventMention) Scan(value interface{}) error {
	// rows written before the column existed come back as NULL
	if value == nil {
		*e = MentionNone
		return nil
	}

	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		return fmt.Errorf("EventMention must be a string or []byte, got %T", value)
	}

	switch strValue {
	case string(MentionNone), string(MentionHere), string(MentionChannel), string(MentionAll):
		*e = EventMention(strValue)
		return nil
	default:
		return fmt.Errorf("invalid EventMention: %s", strValue)
	}
}

// AsText renders the mention the way it has to appear in a post to actually
// notify people; an empty mention renders as an empty string.
func (e EventMention) AsText() string {
	if e == MentionNone {
		return ""
	}
	return "@" + string(e)
}

type Event struct {
	Id          string          `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Description string          `json:"description" db:"description"`
	Start       time.Time       `json:"start" db:"dt_start"`
	End         time.Time       `json:"end" db:"dt_end"`
	Attendees   []string        `json:"attendees"`
	Created     time.Time       `json:"created" db:"created"`
	Updated     time.Time       `json:"updated" db:"updated"`
	Owner       string          `json:"owner" db:"owner"`
	Team        string          `json:"team" db:"team"`
	Channel     *string         `json:"channel" db:"channel"`
	Processed   *time.Time      `json:"-" db:"processed"`
	Recurrent   bool            `json:"-" db:"recurrent"`
	Recurrence  string          `json:"recurrence" db:"recurrence"`
	Color       *string         `json:"color" db:"color"`
	Visibility  EventVisibility `json:"visibility" db:"visibility"`
	Alert       EventAlert      `json:"alert" db:"alert"`
	AlertTime   *time.Time      `json:"alertTime" db:"alert_time"`
	Type        EventType       `json:"type" db:"type"`
	MeetingLink *string         `json:"meetingLink" db:"meeting_link"`

	// Which channel-wide mention to put in the reminder post. Only has an
	// effect for events that post to a channel.
	Mention EventMention `json:"mention" db:"mention"`

	// Start/End store an exclusive range for all-day events (End is midnight
	// of the day after the last day), matching FullCalendar's own convention
	// and RFC 5545's VALUE=DATE semantics, so iCal/CalDAV export needs no
	// special-casing beyond picking the all-day property setters.
	AllDay bool `json:"allDay" db:"all_day"`
}

type UserSettings struct {
	BusinessStartTime     string `json:"businessStartTime"`
	BusinessEndTime       string `json:"businessEndTime"`
	IsOpenCalendarLeftBar bool   `json:"isOpenCalendarLeftBar" db:"is_open_calendar_left_bar"`
	FirstDayOfWeek        int    `json:"firstDayOfWeek" db:"first_day_of_week"`
	BusinessDays          []int  `json:"businessDays"`
	HideNonWorkingDays    bool   `json:"hideNonWorkingDays" db:"hide_non_working_days"`
	CallColor             string `json:"callColor" db:"call_color"`
	EventColor            string `json:"eventColor" db:"event_color"`
	JitsiBaseURL          string `json:"jitsiBaseUrl"`
}

package main

import (
	"time"
)

type Event struct {
	Id          string     `json:"id" db:"id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Start       time.Time  `json:"start" db:"dt_start"`
	End         time.Time  `json:"end" db:"dt_end"`
	Attendees   []string   `json:"attendees"`
	Created     time.Time  `json:"created" db:"created"`
	Owner       string     `json:"owner" db:"owner"`
	Channel     *string    `json:"channel" db:"channel"`
	Processed   *time.Time `json:"-" db:"processed"`
	Recurrent   bool       `json:"-" db:"recurrent"`
	Recurrence  string     `json:"recurrence" db:"recurrence"`
	Color       *string    `json:"color" db:"color"`
}

type UserSettings struct {
	BusinessStartTime     string `json:"businessStartTime"`
	BusinessEndTime       string `json:"businessEndTime"`
	IsOpenCalendarLeftBar bool   `json:"isOpenCalendarLeftBar" db:"is_open_calendar_left_bar"`
	FirstDayOfWeek        int    `json:"firstDayOfWeek" db:"first_day_of_week"`
	BusinessDays          []int  `json:"businessDays"`
	HideNonWorkingDays    bool   `json:"hideNonWorkingDays" db:"hide_non_working_days"`
}

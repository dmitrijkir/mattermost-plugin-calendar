package main

import "time"

// How far back the background job re-checks for one-off events whose exact
// minute-tick was never observed (restart, reload, a slow tick). Bounded so a
// restart never replays old history.
const missedTickWindow = 10 * time.Minute

const (
	PluginId            = "com.dmkir.calendar"
	EventDateTimeLayout = "2006-01-02T15:04:05"
	BusinessTimeLayout  = "15:04"
	DefaultColor        = "#D0D0D0"
	DefaultSlotTime     = 15
)

const (
	POSTGRES = "postgres"
	MYSQL    = "mysql"
)

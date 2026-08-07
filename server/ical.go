package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	sq "github.com/Masterminds/squirrel"
	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/teambition/rrule-go"
)

// ICalToken represents a user's iCal subscription token
type ICalToken struct {
	Token         string     `json:"token" db:"token"`
	UserID        string     `json:"user_id" db:"user_id"`
	Created       time.Time  `json:"created" db:"created"`
	LastUsed      *time.Time `json:"last_used" db:"last_used"`
	CalendarColor *string    `json:"calendar_color" db:"calendar_color"`
}

// ICalTokenResponse is the response format for iCal token API
type ICalTokenResponse struct {
	Token     string `json:"token,omitempty"`
	URL       string `json:"url,omitempty"`
	CalDAVURL string `json:"caldavUrl,omitempty"`
	Enabled   bool   `json:"enabled"`
}

var (
	ICalTokenNotFound = &model.AppError{
		Id:         "ical_token_not_found",
		Message:    "iCal token not found",
		StatusCode: 404,
		Where:      PluginId,
	}

	CantCreateICalToken = &model.AppError{
		Id:         "cant_create_ical_token",
		Message:    "Can't create iCal token",
		StatusCode: 500,
		Where:      PluginId,
	}

	CantRevokeICalToken = &model.AppError{
		Id:         "cant_revoke_ical_token",
		Message:    "Can't revoke iCal token",
		StatusCode: 500,
		Where:      PluginId,
	}

	InvalidICalToken = &model.AppError{
		Id:         "invalid_ical_token",
		Message:    "Invalid iCal token",
		StatusCode: 401,
		Where:      PluginId,
	}
)

// generateSecureToken generates a cryptographically secure 64-character hex token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetICalToken returns the existing iCal token for the authenticated user
func (p *Plugin) GetICalToken(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)
	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, args, sqlErr := queryBuilder.ToSql()
	if sqlErr != nil {
		p.API.LogError("GetICalToken: SQL build error: " + sqlErr.Error())
		errorResponse(w, SomethingWentWrong)
		return
	}

	var token ICalToken
	err2 := p.DB.Get(&token, querySql, args...)
	if err2 != nil {
		// Token doesn't exist, return empty response indicating not enabled
		apiResponse(w, &ICalTokenResponse{
			Enabled: false,
		})
		return
	}

	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	icalURL := ""
	caldavURL := ""
	if siteURL != nil && *siteURL != "" {
		icalURL = *siteURL + "/plugins/" + PluginId + "/ical/feed/" + token.Token
		caldavURL = *siteURL + "/plugins/" + PluginId + "/caldav/" + token.Token + "/"
	}

	apiResponse(w, &ICalTokenResponse{
		Token:     token.Token,
		URL:       icalURL,
		CalDAVURL: caldavURL,
		Enabled:   true,
	})
}

// GenerateICalToken creates a new iCal token for the authenticated user
func (p *Plugin) GenerateICalToken(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)
	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	// Generate a new secure token
	newToken, tokenErr := generateSecureToken()
	if tokenErr != nil {
		p.API.LogError("GenerateICalToken: can't generate token: " + tokenErr.Error())
		errorResponse(w, CantCreateICalToken)
		return
	}

	// Delete existing token if any
	deleteBuilder := sq.Delete("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())
	deleteSql, deleteArgs, _ := deleteBuilder.ToSql()
	_, _ = p.DB.Exec(deleteSql, deleteArgs...)

	// Insert new token
	insertBuilder := sq.Insert("calendar_ical_tokens").
		Columns("token", "user_id", "created").
		Values(newToken, session.UserId, time.Now().UTC()).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	insertSql, insertArgs, sqlErr := insertBuilder.ToSql()
	if sqlErr != nil {
		p.API.LogError("GenerateICalToken: SQL build error: " + sqlErr.Error())
		errorResponse(w, CantCreateICalToken)
		return
	}

	_, insertErr := p.DB.Exec(insertSql, insertArgs...)
	if insertErr != nil {
		p.API.LogError("GenerateICalToken: can't insert token: " + insertErr.Error())
		errorResponse(w, CantCreateICalToken)
		return
	}

	siteURL := p.API.GetConfig().ServiceSettings.SiteURL
	icalURL := ""
	caldavURL := ""
	if siteURL != nil && *siteURL != "" {
		icalURL = *siteURL + "/plugins/" + PluginId + "/ical/feed/" + newToken
		caldavURL = *siteURL + "/plugins/" + PluginId + "/caldav/" + newToken + "/"
	}

	apiResponse(w, &ICalTokenResponse{
		Token:     newToken,
		URL:       icalURL,
		CalDAVURL: caldavURL,
		Enabled:   true,
	})
}

// RevokeICalToken removes the iCal token for the authenticated user
func (p *Plugin) RevokeICalToken(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)
	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	deleteBuilder := sq.Delete("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	deleteSql, deleteArgs, sqlErr := deleteBuilder.ToSql()
	if sqlErr != nil {
		p.API.LogError("RevokeICalToken: SQL build error: " + sqlErr.Error())
		errorResponse(w, CantRevokeICalToken)
		return
	}

	_, deleteErr := p.DB.Exec(deleteSql, deleteArgs...)
	if deleteErr != nil {
		p.API.LogError("RevokeICalToken: can't delete token: " + deleteErr.Error())
		errorResponse(w, CantRevokeICalToken)
		return
	}

	apiResponse(w, &ICalTokenResponse{
		Enabled: false,
	})
}

// ServeICalFeed serves the iCal feed for external calendar applications
// This endpoint is public and authenticated via the token in the URL
func (p *Plugin) ServeICalFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" || len(token) != 64 {
		errorResponse(w, InvalidICalToken)
		return
	}

	// Look up the token and get the user ID
	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": token}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, args, sqlErr := queryBuilder.ToSql()
	if sqlErr != nil {
		p.API.LogError("ServeICalFeed: SQL build error: " + sqlErr.Error())
		errorResponse(w, SomethingWentWrong)
		return
	}

	var icalToken ICalToken
	err := p.DB.Get(&icalToken, querySql, args...)
	if err != nil {
		p.API.LogError("ServeICalFeed: token not found: " + err.Error())
		errorResponse(w, InvalidICalToken)
		return
	}

	// Update last_used timestamp
	go func() {
		updateBuilder := sq.Update("calendar_ical_tokens").
			Set("last_used", time.Now().UTC()).
			Where(sq.Eq{"token": token}).
			PlaceholderFormat(p.GetDBPlaceholderFormat())
		updateSql, updateArgs, _ := updateBuilder.ToSql()
		_, _ = p.DB.Exec(updateSql, updateArgs...)
	}()

	// Get user to verify they still exist
	user, userErr := p.API.GetUser(icalToken.UserID)
	if userErr != nil {
		p.API.LogError("ServeICalFeed: user not found: " + userErr.Error())
		errorResponse(w, InvalidICalToken)
		return
	}

	// Get events for the user (past 30 days to future 365 days)
	now := time.Now().UTC()
	start := now.AddDate(0, -1, 0) // 1 month ago
	end := now.AddDate(1, 0, 0)    // 1 year from now

	userLoc := p.GetUserLocation(user)
	events, eventsErr := p.GetUserEventsForICalUTC(icalToken.UserID, userLoc, start, end)
	if eventsErr != nil {
		p.API.LogError("ServeICalFeed: can't get events: " + eventsErr.Error())
		errorResponse(w, SomethingWentWrong)
		return
	}

	// Generate iCalendar content
	icalContent := p.generateICalendar(events, user)

	// Set proper headers for iCalendar file
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"calendar.ics\"")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(icalContent))
}

// GetUserEventsForICalUTC returns user events without expanding RRULE (keeps recurrence rules intact)
func (p *Plugin) GetUserEventsForICalUTC(
	userId string,
	userLocation *time.Location,
	start, end time.Time,
) ([]Event, *model.AppError) {
	var events []Event

	conditions := sq.And{
		sq.Or{
			sq.Eq{"cm.member": userId},
			sq.Eq{"ce.owner": userId},
			sq.NotEq{"ce.visibility": string(VisibilityPrivate)},
		},
		sq.Or{
			sq.And{
				sq.GtOrEq{"ce.dt_start": start},
				sq.LtOrEq{"ce.dt_start": end},
			},
			sq.Eq{"ce.recurrent": true},
		},
	}

	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.description",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.team",
			"ce.visibility",
			"ce.alert",
			"ce.alert_time",
			"ce.type",
			"ce.meeting_link",
			"ce.all_day",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(conditions).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, args, err := queryBuilder.ToSql()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, SomethingWentWrong
	}

	rows, errSelect := p.DB.Queryx(querySql, args...)
	if errSelect != nil {
		p.API.LogError(errSelect.Error())
		return nil, SomethingWentWrong
	}
	defer rows.Close()

	userTeams, _ := p.GetUserTeams(userId)
	userChannels, _ := p.GetUserChannels(userId)

	addedEvent := map[string]bool{}
	for rows.Next() {
		var eventDb Event
		errScan := rows.StructScan(&eventDb)
		if errScan != nil {
			p.API.LogError("Can't scan row to struct: " + errScan.Error())
			continue
		}

		if addedEvent[eventDb.Id] {
			continue
		}

		if eventDb.Visibility == VisibilityChannel && eventDb.Channel == nil {
			continue
		}

		if eventDb.Visibility == VisibilityChannel && !contains(userChannels, *eventDb.Channel) {
			continue
		}

		if eventDb.Visibility == VisibilityTeam && !contains(userTeams, eventDb.Team) {
			continue
		}

		// For iCal, we don't expand recurrent events - we include RRULE
		events = append(events, eventDb)
		addedEvent[eventDb.Id] = true
	}

	return events, nil
}

// generateICalendar converts events to iCalendar format
func (p *Plugin) generateICalendar(events []Event, user *model.User) string {
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//Mattermost Calendar Plugin//EN")
	cal.SetVersion("2.0")
	cal.SetCalscale("GREGORIAN")
	cal.SetName("Mattermost Calendar")
	cal.SetXWRCalName("Mattermost Calendar")

	for _, event := range events {
		icsEvent := cal.AddEvent(event.Id)
		icsEvent.SetDtStampTime(event.Created)
		icsEvent.SetCreatedTime(event.Created)
		icsEvent.SetModifiedAt(event.Created)
		if event.AllDay {
			// Start/End are already stored as an exclusive day range, matching
			// what VALUE=DATE expects
			icsEvent.SetAllDayStartAt(event.Start)
			icsEvent.SetAllDayEndAt(event.End)
		} else {
			icsEvent.SetStartAt(event.Start)
			icsEvent.SetEndAt(event.End)
		}
		icsEvent.SetSummary(event.Title)

		if event.Description != "" {
			icsEvent.SetDescription(event.Description)
		}

		// LOCATION is what most calendar clients surface as the "join" target,
		// URL keeps it available to the ones that prefer that property.
		if event.MeetingLink != nil && *event.MeetingLink != "" {
			icsEvent.SetLocation(*event.MeetingLink)
			icsEvent.SetURL(*event.MeetingLink)
		}

		// Add organizer
		if user != nil {
			icsEvent.SetOrganizer(user.Email, ics.WithCN(user.GetDisplayName("")))
		}

		// Add RRULE for recurrent events
		if event.Recurrent && event.Recurrence != "" {
			// Parse and validate the RRULE
			_, rruleErr := rrule.StrToRRule(event.Recurrence)
			if rruleErr == nil {
				// Add RRULE property - remove "RRULE:" prefix if present
				rruleStr := event.Recurrence
				if len(rruleStr) > 6 && rruleStr[:6] == "RRULE:" {
					rruleStr = rruleStr[6:]
				}
				icsEvent.AddRrule(rruleStr)
			}
		}

		// Add alarm if configured
		if event.Alert != EventAlertNone {
			if alertDuration, ok := EventAlertDurationMap[event.Alert]; ok && alertDuration > 0 {
				alarm := icsEvent.AddAlarm()
				alarm.SetAction(ics.ActionDisplay)
				alarm.SetTrigger("-PT" + formatDurationForICS(alertDuration))
				alarm.SetProperty(ics.ComponentPropertyDescription, "Reminder: "+event.Title)
			}
		}

		// Set status
		icsEvent.SetStatus(ics.ObjectStatusConfirmed)
	}

	return cal.Serialize()
}

// formatDurationForICS formats a duration as ISO 8601 duration for iCalendar
func formatDurationForICS(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		hours = hours % 24
		if hours > 0 {
			return intToString(days) + "DT" + intToString(hours) + "H"
		}
		return intToString(days) + "D"
	}

	if hours > 0 && minutes > 0 {
		return intToString(hours) + "H" + intToString(minutes) + "M"
	} else if hours > 0 {
		return intToString(hours) + "H"
	}
	return intToString(minutes) + "M"
}

// intToString converts an int to string without using strconv
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + intToString(-n)
	}
	result := ""
	for n > 0 {
		digit := n % 10
		result = string(rune('0'+digit)) + result
		n /= 10
	}
	return result
}

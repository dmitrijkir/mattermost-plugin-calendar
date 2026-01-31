package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/emersion/go-ical"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/model"
)

// CalDAVBackend handles CalDAV operations for Mattermost Calendar
type CalDAVBackend struct {
	plugin   *Plugin
	userID   string
	token    string
	basePath string
}

// NewCalDAVBackend creates a new CalDAV backend for a specific user
func NewCalDAVBackend(plugin *Plugin, userID, token string) *CalDAVBackend {
	return &CalDAVBackend{
		plugin:   plugin,
		userID:   userID,
		token:    token,
		basePath: "/plugins/" + PluginId + "/caldav/" + token,
	}
}

// ServeCalDAV handles CalDAV requests
func (p *Plugin) ServeCalDAV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" || len(token) != 64 {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// CalDAV clients require Basic Auth - we accept any credentials since real auth is via token
	username, password, ok := r.BasicAuth()
	if !ok || username == "" || password == "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="CalDAV"`)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Look up the token and get the user ID
	queryBuilder := sq.Select("token", "user_id", "created", "last_used").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": token}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, args, sqlErr := queryBuilder.ToSql()
	if sqlErr != nil {
		p.API.LogError("ServeCalDAV: SQL build error: " + sqlErr.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var icalToken ICalToken
	err := p.DB.Get(&icalToken, querySql, args...)
	if err != nil {
		p.API.LogError("ServeCalDAV: token not found: " + err.Error())
		http.Error(w, "Invalid token", http.StatusUnauthorized)
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

	// Verify user exists
	_, userErr := p.API.GetUser(icalToken.UserID)
	if userErr != nil {
		p.API.LogError("ServeCalDAV: user not found: " + userErr.Error())
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Create backend and handle request
	backend := NewCalDAVBackend(p, icalToken.UserID, token)
	backend.ServeHTTP(w, r)
}

// ServeHTTP handles CalDAV HTTP requests
func (b *CalDAVBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.plugin.API.LogInfo("CalDAV request", "method", r.Method, "path", r.URL.Path)

	// Set DAV header for all responses (required by Apple Calendar)
	w.Header().Set("DAV", "1, 2, 3, calendar-access")

	switch r.Method {
	case "OPTIONS":
		b.handleOptions(w, r)
	case "PROPFIND":
		b.handlePropfind(w, r)
	case "REPORT":
		b.handleReport(w, r)
	case "GET":
		b.handleGet(w, r)
	case "PUT":
		b.handlePut(w, r)
	case "DELETE":
		b.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (b *CalDAVBackend) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS, PROPFIND, REPORT, GET, PUT, DELETE")
	w.Header().Set("DAV", "1, 2, calendar-access")
	w.WriteHeader(http.StatusNoContent)
}

func (b *CalDAVBackend) handlePropfind(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Determine what we're querying
	isRoot := strings.HasSuffix(path, b.token+"/") || strings.HasSuffix(path, b.token)
	isCalendar := strings.Contains(path, "/calendar")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	var response string
	if isCalendar {
		response = b.calendarPropfindResponse()
	} else if isRoot {
		response = b.rootPropfindResponse()
	} else {
		response = b.rootPropfindResponse()
	}

	w.Write([]byte(response))
}

func (b *CalDAVBackend) rootPropfindResponse() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav" xmlns:cs="http://calendarserver.org/ns/">
  <d:response>
    <d:href>%s/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype>
          <d:collection/>
        </d:resourcetype>
        <d:current-user-principal>
          <d:href>%s/</d:href>
        </d:current-user-principal>
        <d:principal-URL>
          <d:href>%s/</d:href>
        </d:principal-URL>
        <cal:calendar-home-set>
          <d:href>%s/</d:href>
        </cal:calendar-home-set>
        <cal:calendar-user-address-set>
          <d:href>mailto:user@localhost</d:href>
        </cal:calendar-user-address-set>
        <d:displayname>Mattermost User</d:displayname>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
  <d:response>
    <d:href>%s/calendar/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype>
          <d:collection/>
          <cal:calendar/>
        </d:resourcetype>
        <d:displayname>Mattermost Calendar</d:displayname>
        <cal:supported-calendar-component-set>
          <cal:comp name="VEVENT"/>
        </cal:supported-calendar-component-set>
        <cs:getctag>%d</cs:getctag>
        <d:sync-token>%d</d:sync-token>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`, b.basePath, b.basePath, b.basePath, b.basePath, b.basePath, time.Now().Unix(), time.Now().Unix())
}

func (b *CalDAVBackend) calendarPropfindResponse() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav" xmlns:cs="http://calendarserver.org/ns/" xmlns:ical="http://apple.com/ns/ical/">
  <d:response>
    <d:href>%s/calendar/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype>
          <d:collection/>
          <cal:calendar/>
        </d:resourcetype>
        <d:displayname>Mattermost Calendar</d:displayname>
        <cal:supported-calendar-component-set>
          <cal:comp name="VEVENT"/>
        </cal:supported-calendar-component-set>
        <cs:getctag>%d</cs:getctag>
        <d:sync-token>%d</d:sync-token>
        <ical:calendar-color>#1E90FFFF</ical:calendar-color>
        <ical:calendar-order>1</ical:calendar-order>
        <d:current-user-privilege-set>
          <d:privilege><d:read/></d:privilege>
          <d:privilege><d:write/></d:privilege>
          <d:privilege><d:write-content/></d:privilege>
        </d:current-user-privilege-set>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`, b.basePath, time.Now().Unix(), time.Now().Unix())
}

func (b *CalDAVBackend) handleReport(w http.ResponseWriter, r *http.Request) {
	user, userErr := b.plugin.API.GetUser(b.userID)
	if userErr != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Get events
	now := time.Now().UTC()
	start := now.AddDate(-1, 0, 0)
	end := now.AddDate(2, 0, 0)

	userLoc := b.plugin.GetUserLocation(user)
	events, eventsErr := b.plugin.GetUserEventsForICalUTC(b.userID, userLoc, start, end)
	if eventsErr != nil {
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav">`)

	for _, event := range events {
		icalData := b.eventToICalendarString(&event, user)
		buf.WriteString(fmt.Sprintf(`
  <d:response>
    <d:href>%s/calendar/%s.ics</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"%d"</d:getetag>
        <cal:calendar-data>%s</cal:calendar-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>`, b.basePath, event.Id, event.Created.UnixNano(), xmlEscape(icalData)))
	}

	buf.WriteString(`</d:multistatus>`)
	w.Write(buf.Bytes())
}

func (b *CalDAVBackend) handleGet(w http.ResponseWriter, r *http.Request) {
	eventID := b.extractEventID(r.URL.Path)
	if eventID == "" {
		http.Error(w, "Invalid event path", http.StatusBadRequest)
		return
	}

	event, err := b.getEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	user, _ := b.plugin.API.GetUser(b.userID)
	icalData := b.eventToICalendarString(event, user)

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, event.Created.UnixNano()))
	w.Write([]byte(icalData))
}

func (b *CalDAVBackend) handlePut(w http.ResponseWriter, r *http.Request) {
	eventID := b.extractEventID(r.URL.Path)
	if eventID == "" {
		// Generate new ID for new events
		eventID = uuid.New().String()
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	cal, err := ical.NewDecoder(bytes.NewReader(body)).Decode()
	if err != nil {
		http.Error(w, "Failed to parse iCalendar data", http.StatusBadRequest)
		return
	}

	event, err := b.icalendarToEvent(cal, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if event exists
	existingEvent, _ := b.getEventByID(eventID)

	if existingEvent != nil {
		err = b.updateEvent(event)
	} else {
		event.Id = eventID
		event.Owner = b.userID
		event.Created = time.Now().UTC()
		err = b.createEvent(event)
	}

	if err != nil {
		b.plugin.API.LogError("CalDAV PUT error: " + err.Error())
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, event.Created.UnixNano()))
	w.WriteHeader(http.StatusCreated)
}

func (b *CalDAVBackend) handleDelete(w http.ResponseWriter, r *http.Request) {
	eventID := b.extractEventID(r.URL.Path)
	if eventID == "" {
		http.Error(w, "Invalid event path", http.StatusBadRequest)
		return
	}

	// Check if user owns the event
	event, err := b.getEventByID(eventID)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	if event.Owner != b.userID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	deleteBuilder := sq.Delete("calendar_events").
		Where(sq.Eq{"id": eventID}).
		PlaceholderFormat(b.plugin.GetDBPlaceholderFormat())

	deleteSql, deleteArgs, _ := deleteBuilder.ToSql()
	_, dbErr := b.plugin.DB.Exec(deleteSql, deleteArgs...)
	if dbErr != nil {
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (b *CalDAVBackend) extractEventID(path string) string {
	path = strings.TrimSuffix(path, ".ics")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	eventID := parts[len(parts)-1]
	if eventID == "calendar" || eventID == "" {
		return ""
	}
	return eventID
}

func (b *CalDAVBackend) getEventByID(eventID string) (*Event, error) {
	queryBuilder := sq.Select(
		"id", "title", "description", "dt_start", "dt_end",
		"created", "owner", "channel", "recurrent", "recurrence",
		"color", "team", "visibility", "alert", "alert_time",
	).
		From("calendar_events").
		Where(sq.Eq{"id": eventID}).
		PlaceholderFormat(b.plugin.GetDBPlaceholderFormat())

	querySql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("SQL build error: %w", err)
	}

	var event Event
	err = b.plugin.DB.Get(&event, querySql, args...)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	return &event, nil
}

func (b *CalDAVBackend) eventToICalendarString(event *Event, user *model.User) string {
	var buf bytes.Buffer
	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//Mattermost Calendar Plugin//EN\r\n")
	buf.WriteString("BEGIN:VEVENT\r\n")
	buf.WriteString(fmt.Sprintf("UID:%s\r\n", event.Id))
	buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", event.Created.UTC().Format("20060102T150405Z")))
	buf.WriteString(fmt.Sprintf("DTSTART:%s\r\n", event.Start.UTC().Format("20060102T150405Z")))
	buf.WriteString(fmt.Sprintf("DTEND:%s\r\n", event.End.UTC().Format("20060102T150405Z")))
	buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", event.Title))

	if event.Description != "" {
		buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", event.Description))
	}

	if user != nil && user.Email != "" {
		buf.WriteString(fmt.Sprintf("ORGANIZER:mailto:%s\r\n", user.Email))
	}

	if event.Recurrent && event.Recurrence != "" {
		rrule := event.Recurrence
		if !strings.HasPrefix(rrule, "RRULE:") {
			rrule = "RRULE:" + rrule
		}
		buf.WriteString(rrule + "\r\n")
	}

	buf.WriteString("STATUS:CONFIRMED\r\n")
	buf.WriteString("END:VEVENT\r\n")
	buf.WriteString("END:VCALENDAR\r\n")

	return buf.String()
}

func (b *CalDAVBackend) icalendarToEvent(cal *ical.Calendar, eventID string) (*Event, error) {
	var vevent *ical.Component
	for _, child := range cal.Children {
		if child.Name == ical.CompEvent {
			vevent = child
			break
		}
	}

	if vevent == nil {
		return nil, fmt.Errorf("no VEVENT found in calendar")
	}

	event := &Event{
		Id:         eventID,
		Visibility: VisibilityPrivate,
	}

	if summary := vevent.Props.Get(ical.PropSummary); summary != nil {
		event.Title = summary.Value
	}

	if desc := vevent.Props.Get(ical.PropDescription); desc != nil {
		event.Description = desc.Value
	}

	if dtstart := vevent.Props.Get(ical.PropDateTimeStart); dtstart != nil {
		start, err := dtstart.DateTime(time.UTC)
		if err == nil {
			event.Start = start
		}
	}

	if dtend := vevent.Props.Get(ical.PropDateTimeEnd); dtend != nil {
		end, err := dtend.DateTime(time.UTC)
		if err == nil {
			event.End = end
		}
	}

	if rrule := vevent.Props.Get(ical.PropRecurrenceRule); rrule != nil {
		event.Recurrence = "RRULE:" + rrule.Value
		event.Recurrent = true
	}

	if event.Id == "" {
		event.Id = uuid.New().String()
	}

	return event, nil
}

func (b *CalDAVBackend) createEvent(event *Event) error {
	if event.Team == "" {
		teams, _ := b.plugin.API.GetTeamsForUser(b.userID)
		if len(teams) > 0 {
			event.Team = teams[0].Id
		}
	}

	queryBuilder := sq.Insert("calendar_events").
		Columns(
			"id", "title", "description", "dt_start", "dt_end",
			"created", "owner", "channel", "recurrent", "recurrence",
			"color", "visibility", "team", "alert", "alert_time",
		).
		Values(
			event.Id, event.Title, event.Description, event.Start, event.End,
			event.Created, event.Owner, event.Channel, event.Recurrent, event.Recurrence,
			event.Color, event.Visibility, event.Team, event.Alert, event.AlertTime,
		).
		PlaceholderFormat(b.plugin.GetDBPlaceholderFormat())

	insertSql, insertArgs, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("SQL build error: %w", err)
	}

	_, dbErr := b.plugin.DB.Exec(insertSql, insertArgs...)
	if dbErr != nil {
		return fmt.Errorf("insert error: %w", dbErr)
	}

	return nil
}

func (b *CalDAVBackend) updateEvent(event *Event) error {
	updateFields := map[string]interface{}{
		"title":       event.Title,
		"description": event.Description,
		"dt_start":    event.Start,
		"dt_end":      event.End,
		"recurrence":  event.Recurrence,
		"recurrent":   event.Recurrent,
	}

	updateQueryBuilder := sq.Update("calendar_events").
		SetMap(updateFields).
		Where(sq.Eq{"id": event.Id}).
		PlaceholderFormat(b.plugin.GetDBPlaceholderFormat())

	updateSql, updateArgs, err := updateQueryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("SQL build error: %w", err)
	}

	_, dbErr := b.plugin.DB.Exec(updateSql, updateArgs...)
	if dbErr != nil {
		return fmt.Errorf("update error: %w", dbErr)
	}

	return nil
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

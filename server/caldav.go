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
	ics "github.com/arran4/golang-ical"
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

	// Log all incoming requests for debugging
	p.API.LogInfo("CalDAV incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"host", r.Host,
		"user-agent", r.Header.Get("User-Agent"),
		"hasAuth", r.Header.Get("Authorization") != "")

	if token == "" || len(token) != 64 {
		p.API.LogError("CalDAV invalid token", "token_length", len(token))
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Set DAV headers early for discovery (Apple Calendar needs this)
	w.Header().Set("DAV", "1, 2, 3, calendar-access, addressbook")
	w.Header().Set("Server", "Mattermost-Calendar-CalDAV/1.0")

	// Allow OPTIONS without authentication for CalDAV discovery (required by Apple Calendar)
	if r.Method == "OPTIONS" {
		w.Header().Set("Allow", "OPTIONS, PROPFIND, PROPPATCH, REPORT, GET, PUT, DELETE, MKCALENDAR")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authentication is via token in URL - no Basic Auth required
	// This allows Apple Calendar to work over HTTP (it refuses to send passwords over HTTP)
	// The 64-char token in the URL is the secret that authenticates the user

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
	case "PROPPATCH":
		b.handleProppatch(w, r)
	case "REPORT":
		b.handleReport(w, r)
	case "GET":
		b.handleGet(w, r)
	case "PUT":
		b.handlePut(w, r)
	case "DELETE":
		b.handleDelete(w, r)
	case "MKCALENDAR":
		// We don't support creating additional calendars
		http.Error(w, "Calendar already exists", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (b *CalDAVBackend) handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "OPTIONS, PROPFIND, PROPPATCH, REPORT, GET, PUT, DELETE")
	w.Header().Set("DAV", "1, 2, 3, calendar-access")
	w.WriteHeader(http.StatusOK)
}

func (b *CalDAVBackend) handleProppatch(w http.ResponseWriter, r *http.Request) {
	// Apple Calendar may try to set calendar properties (color, display name, etc.)
	// We accept the request but don't actually store the changes
	// Return success to keep the client happy
	b.plugin.API.LogInfo("CalDAV PROPPATCH", "path", r.URL.Path)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<d:multistatus xmlns:d="DAV:">
  <d:response>
    <d:href>%s</d:href>
    <d:propstat>
      <d:prop/>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>
</d:multistatus>`, r.URL.Path)

	w.Write([]byte(response))
}

func (b *CalDAVBackend) handlePropfind(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	depth := r.Header.Get("Depth")

	// Read request body to see what properties are requested
	body, _ := io.ReadAll(r.Body)
	b.plugin.API.LogInfo("CalDAV PROPFIND", "path", path, "depth", depth, "body", string(body))

	// Determine what we're querying
	isRoot := strings.HasSuffix(path, b.token+"/") || strings.HasSuffix(path, b.token)
	isCalendar := strings.Contains(path, "/calendar")

	b.plugin.API.LogInfo("CalDAV PROPFIND routing", "isRoot", isRoot, "isCalendar", isCalendar, "depth", depth)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusMultiStatus)

	var response string
	if isCalendar && depth == "1" {
		// Depth 1 on calendar - return calendar info + list of events
		response = b.calendarPropfindWithEvents()
	} else if isCalendar {
		response = b.calendarPropfindResponse()
	} else if isRoot && depth == "1" {
		// Depth 1 on root - return root + calendar collection (Apple Calendar needs this)
		response = b.rootPropfindWithCalendars()
	} else if isRoot {
		response = b.rootPropfindResponse()
	} else {
		response = b.rootPropfindResponse()
	}

	b.plugin.API.LogInfo("CalDAV PROPFIND response length", "length", len(response))
	w.Write([]byte(response))
}

func (b *CalDAVBackend) rootPropfindResponse() string {
	// Return only info about the principal/root - not the calendar
	// Apple Calendar expects current-user-principal to point to a principal resource
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:CS="http://calendarserver.org/ns/">
  <D:response>
    <D:href>%s/</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype>
          <D:collection/>
          <D:principal/>
        </D:resourcetype>
        <D:current-user-principal>
          <D:href>%s/</D:href>
        </D:current-user-principal>
        <D:principal-URL>
          <D:href>%s/</D:href>
        </D:principal-URL>
        <C:calendar-home-set>
          <D:href>%s/</D:href>
        </C:calendar-home-set>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`, b.basePath, b.basePath, b.basePath, b.basePath)
}

func (b *CalDAVBackend) rootPropfindWithCalendars() string {
	// Return root + calendar collection for Depth: 1 requests (Apple Calendar needs this)
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:CS="http://calendarserver.org/ns/" xmlns:I="http://apple.com/ns/ical/">
  <D:response>
    <D:href>%s/</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype>
          <D:collection/>
        </D:resourcetype>
        <D:current-user-principal>
          <D:href>%s/</D:href>
        </D:current-user-principal>
        <D:displayname>Mattermost Calendar</D:displayname>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
  <D:response>
    <D:href>%s/calendar/</D:href>
    <D:propstat>
      <D:prop>
        <D:resourcetype>
          <D:collection/>
          <C:calendar/>
        </D:resourcetype>
        <D:displayname>Mattermost Calendar</D:displayname>
        <I:calendar-color>#1E90FFFF</I:calendar-color>
        <I:calendar-order>1</I:calendar-order>
        <C:supported-calendar-component-set>
          <C:comp name="VEVENT"/>
        </C:supported-calendar-component-set>
        <CS:getctag>%d</CS:getctag>
        <D:current-user-privilege-set>
          <D:privilege><D:read/></D:privilege>
          <D:privilege><D:write/></D:privilege>
          <D:privilege><D:write-content/></D:privilege>
        </D:current-user-privilege-set>
      </D:prop>
      <D:status>HTTP/1.1 200 OK</D:status>
    </D:propstat>
  </D:response>
</D:multistatus>`, b.basePath, b.basePath, b.basePath, time.Now().Unix())
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

func (b *CalDAVBackend) calendarPropfindWithEvents() string {
	user, _ := b.plugin.API.GetUser(b.userID)

	// Get events
	now := time.Now().UTC()
	start := now.AddDate(-1, 0, 0)
	end := now.AddDate(2, 0, 0)

	var userLoc *time.Location
	if user != nil {
		userLoc = b.plugin.GetUserLocation(user)
	}
	events, _ := b.plugin.GetUserEventsForICalUTC(b.userID, userLoc, start, end)

	// Log events for debugging
	var eventIDs []string
	for _, e := range events {
		eventIDs = append(eventIDs, e.Id)
	}
	b.plugin.API.LogInfo("CalDAV PROPFIND events", "count", len(events), "ids", strings.Join(eventIDs, ","))

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	buf.WriteString(`<d:multistatus xmlns:d="DAV:" xmlns:cal="urn:ietf:params:xml:ns:caldav" xmlns:cs="http://calendarserver.org/ns/" xmlns:ical="http://apple.com/ns/ical/">`)

	// First, the calendar collection itself
	buf.WriteString(fmt.Sprintf(`
  <d:response>
    <d:href>%s/calendar/</d:href>
    <d:propstat>
      <d:prop>
        <d:resourcetype>
          <d:collection/>
          <cal:calendar/>
        </d:resourcetype>
        <d:displayname>Mattermost Calendar</d:displayname>
        <cs:getctag>%d</cs:getctag>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>`, b.basePath, time.Now().Unix()))

	// Then each event
	for _, event := range events {
		buf.WriteString(fmt.Sprintf(`
  <d:response>
    <d:href>%s/calendar/%s.ics</d:href>
    <d:propstat>
      <d:prop>
        <d:getetag>"%s"</d:getetag>
        <d:getcontenttype>text/calendar; charset=utf-8</d:getcontenttype>
        <d:resourcetype/>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>`, b.basePath, event.Id, eventETag(&event)))
	}

	buf.WriteString(`</d:multistatus>`)
	return buf.String()
}

func (b *CalDAVBackend) handleReport(w http.ResponseWriter, r *http.Request) {
	user, userErr := b.plugin.API.GetUser(b.userID)
	if userErr != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Read and parse the REPORT body to see what's being requested
	body, _ := io.ReadAll(r.Body)
	b.plugin.API.LogInfo("CalDAV REPORT body", "body", string(body))

	// Check if this is a calendar-multiget (requesting specific events)
	var requestedEventIDs []string
	if strings.Contains(string(body), "calendar-multiget") {
		// Parse hrefs from the request using regex
		// Looking for href tags like <D:href>...</D:href>
		bodyStr := string(body)
		// Find all href values - they end with .ics
		idx := 0
		for {
			// Find href opening tag (case insensitive)
			hrefStart := strings.Index(strings.ToLower(bodyStr[idx:]), "<d:href>")
			if hrefStart == -1 {
				hrefStart = strings.Index(strings.ToLower(bodyStr[idx:]), "<href>")
			}
			if hrefStart == -1 {
				break
			}
			hrefStart += idx

			// Find the closing tag
			var hrefEnd int
			if strings.Contains(bodyStr[hrefStart:hrefStart+10], "<D:href>") || strings.Contains(bodyStr[hrefStart:hrefStart+10], "<d:href>") {
				hrefEnd = strings.Index(bodyStr[hrefStart:], "</D:href>")
				if hrefEnd == -1 {
					hrefEnd = strings.Index(bodyStr[hrefStart:], "</d:href>")
				}
			} else {
				hrefEnd = strings.Index(bodyStr[hrefStart:], "</href>")
			}
			if hrefEnd == -1 {
				break
			}
			hrefEnd += hrefStart

			// Extract href content
			tagEnd := strings.Index(bodyStr[hrefStart:], ">") + hrefStart + 1
			href := bodyStr[tagEnd:hrefEnd]

			// Extract event ID from href (last part before .ics)
			if strings.HasSuffix(href, ".ics") {
				parts := strings.Split(href, "/")
				if len(parts) > 0 {
					eventID := strings.TrimSuffix(parts[len(parts)-1], ".ics")
					requestedEventIDs = append(requestedEventIDs, eventID)
				}
			}

			idx = hrefEnd + 1
		}
		b.plugin.API.LogInfo("CalDAV REPORT multiget", "requestedEventIDs", strings.Join(requestedEventIDs, ","))
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

	// If specific event IDs were requested, filter events
	if len(requestedEventIDs) > 0 {
		var filteredEvents []Event
		for _, event := range events {
			for _, reqID := range requestedEventIDs {
				if event.Id == reqID {
					filteredEvents = append(filteredEvents, event)
					break
				}
			}
		}
		b.plugin.API.LogInfo("CalDAV REPORT filtered", "requested", len(requestedEventIDs), "found", len(filteredEvents))
		events = filteredEvents
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
        <d:getetag>"%s"</d:getetag>
        <cal:calendar-data>%s</cal:calendar-data>
      </d:prop>
      <d:status>HTTP/1.1 200 OK</d:status>
    </d:propstat>
  </d:response>`, b.basePath, event.Id, eventETag(&event), xmlEscape(icalData)))
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
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, eventETag(event)))
	w.Write([]byte(icalData))
}

func (b *CalDAVBackend) handlePut(w http.ResponseWriter, r *http.Request) {
	urlEventID := b.extractEventID(r.URL.Path)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	cal, err := ics.ParseCalendar(bytes.NewReader(body))
	if err != nil {
		b.plugin.API.LogError("CalDAV PUT: failed to parse iCal", "error", err.Error(), "body", string(body))
		http.Error(w, "Failed to parse iCalendar data", http.StatusBadRequest)
		return
	}

	// Extract UID from iCal data
	icalUID := ""
	events := cal.Events()
	if len(events) > 0 {
		if uid := events[0].GetProperty(ics.ComponentPropertyUniqueId); uid != nil {
			icalUID = uid.Value
		}
	}

	// Determine event ID: prefer URL ID, fallback to iCal UID, generate new if neither
	eventID := urlEventID
	if eventID == "" {
		eventID = icalUID
	}
	if eventID == "" {
		eventID = uuid.New().String()
	}

	b.plugin.API.LogInfo("CalDAV PUT", "urlEventID", urlEventID, "icalUID", icalUID, "eventID", eventID)

	event, err := b.icalendarToEvent(cal, eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if event exists by URL ID first, then by iCal UID
	existingEvent, _ := b.getEventByID(eventID)
	if existingEvent == nil && icalUID != "" && icalUID != eventID {
		existingEvent, _ = b.getEventByID(icalUID)
		if existingEvent != nil {
			eventID = icalUID
			event.Id = icalUID
		}
	}

	isUpdate := existingEvent != nil
	b.plugin.API.LogInfo("CalDAV PUT decision", "isUpdate", isUpdate, "finalEventID", eventID)

	if isUpdate {
		// Preserve owner and created time from existing event
		event.Owner = existingEvent.Owner
		event.Created = existingEvent.Created

		// Check if user owns the event
		if existingEvent.Owner != b.userID {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		err = b.updateEvent(event)
		b.plugin.API.LogInfo("CalDAV PUT update", "eventID", eventID, "title", event.Title, "error", fmt.Sprintf("%v", err))
	} else {
		event.Id = eventID
		event.Owner = b.userID
		event.Created = time.Now().UTC()
		err = b.createEvent(event)
		b.plugin.API.LogInfo("CalDAV PUT create", "eventID", eventID, "title", event.Title, "error", fmt.Sprintf("%v", err))
	}

	if err != nil {
		b.plugin.API.LogError("CalDAV PUT error: " + err.Error())
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	// Get the saved event to return correct ETag
	savedEvent, _ := b.getEventByID(eventID)
	if savedEvent != nil {
		etag := eventETag(savedEvent)
		w.Header().Set("ETag", fmt.Sprintf(`"%s"`, etag))
		b.plugin.API.LogInfo("CalDAV PUT response", "eventID", eventID, "etag", etag, "isUpdate", isUpdate)
	}

	if isUpdate {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.Header().Set("Location", fmt.Sprintf("%s/calendar/%s.ics", b.basePath, eventID))
		w.WriteHeader(http.StatusCreated)
	}
}

func (b *CalDAVBackend) handleDelete(w http.ResponseWriter, r *http.Request) {
	eventID := b.extractEventID(r.URL.Path)
	b.plugin.API.LogInfo("CalDAV DELETE", "eventID", eventID, "path", r.URL.Path)

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
	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId("-//Mattermost Calendar Plugin//EN")
	cal.SetVersion("2.0")

	icsEvent := cal.AddEvent(event.Id)
	icsEvent.SetDtStampTime(event.Created)
	icsEvent.SetCreatedTime(event.Created)
	icsEvent.SetStartAt(event.Start)
	icsEvent.SetEndAt(event.End)
	icsEvent.SetSummary(event.Title)

	if event.Description != "" {
		icsEvent.SetDescription(event.Description)
	}

	if user != nil && user.Email != "" {
		icsEvent.SetOrganizer(user.Email, ics.WithCN(user.GetDisplayName("")))
	}

	if event.Recurrent && event.Recurrence != "" {
		rruleStr := event.Recurrence
		if strings.HasPrefix(rruleStr, "RRULE:") {
			rruleStr = rruleStr[6:]
		}
		icsEvent.AddRrule(rruleStr)
	}

	icsEvent.SetStatus(ics.ObjectStatusConfirmed)

	return cal.Serialize()
}

func (b *CalDAVBackend) icalendarToEvent(cal *ics.Calendar, eventID string) (*Event, error) {
	events := cal.Events()
	if len(events) == 0 {
		return nil, fmt.Errorf("no VEVENT found in calendar")
	}

	vevent := events[0]

	event := &Event{
		Id:         eventID,
		Visibility: VisibilityPrivate,
	}

	if summary := vevent.GetProperty(ics.ComponentPropertySummary); summary != nil {
		event.Title = summary.Value
	}

	if desc := vevent.GetProperty(ics.ComponentPropertyDescription); desc != nil {
		event.Description = desc.Value
	}

	if dtstart := vevent.GetProperty(ics.ComponentPropertyDtStart); dtstart != nil {
		start, err := time.Parse("20060102T150405Z", dtstart.Value)
		if err != nil {
			// Try parsing with timezone
			start, err = time.Parse("20060102T150405", dtstart.Value)
			if err == nil {
				start = start.UTC()
			}
		}
		if err == nil {
			event.Start = start
		}
	}

	if dtend := vevent.GetProperty(ics.ComponentPropertyDtEnd); dtend != nil {
		end, err := time.Parse("20060102T150405Z", dtend.Value)
		if err != nil {
			end, err = time.Parse("20060102T150405", dtend.Value)
			if err == nil {
				end = end.UTC()
			}
		}
		if err == nil {
			event.End = end
		}
	}

	if rrule := vevent.GetProperty(ics.ComponentPropertyRrule); rrule != nil {
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

	now := time.Now().UTC()
	event.Created = now
	event.Updated = now

	queryBuilder := sq.Insert("calendar_events").
		Columns(
			"id", "title", "description", "dt_start", "dt_end",
			"created", "updated", "owner", "channel", "recurrent", "recurrence",
			"color", "visibility", "team", "alert", "alert_time",
		).
		Values(
			event.Id, event.Title, event.Description, event.Start, event.End,
			event.Created, event.Updated, event.Owner, event.Channel, event.Recurrent, event.Recurrence,
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
	event.Updated = time.Now().UTC()

	updateFields := map[string]interface{}{
		"title":       event.Title,
		"description": event.Description,
		"dt_start":    event.Start,
		"dt_end":      event.End,
		"recurrence":  event.Recurrence,
		"recurrent":   event.Recurrent,
		"updated":     event.Updated,
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

// eventETag generates a deterministic ETag based on event's Updated timestamp
func eventETag(event *Event) string {
	// Use Updated timestamp for ETag - changes every time event is modified
	return fmt.Sprintf("%x", event.Updated.UnixNano())
}

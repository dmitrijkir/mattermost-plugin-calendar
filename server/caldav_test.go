package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	ics "github.com/arran4/golang-ical"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCalDAVBackend_extractEventID(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	tests := []struct {
		path     string
		expected string
	}{
		{"/plugins/com.dmkir.calendar/caldav/token/calendar/event-123.ics", "event-123"},
		{"/plugins/com.dmkir.calendar/caldav/token/calendar/abc-def-ghi.ics", "abc-def-ghi"},
		{"/calendar/event-123.ics", "event-123"},
		{"/calendar/", ""},
		{"/calendar", ""},
		{"event-123.ics", "event-123"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := backend.extractEventID(tt.path)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestEventETag(t *testing.T) {
	assert := assert.New(t)

	now := time.Now().UTC()
	event := &Event{
		Id:          "event-123",
		Title:       "Test Event",
		Description: "Test description",
		Start:       now,
		End:         now.Add(time.Hour),
		Updated:     now,
		Recurrent:   false,
		Recurrence:  "",
	}

	etag1 := eventETag(event)
	assert.NotEmpty(etag1)

	// Same event should produce same ETag
	etag2 := eventETag(event)
	assert.Equal(etag1, etag2)

	// Changing Updated timestamp should produce different ETag
	event.Updated = now.Add(time.Second)
	etag3 := eventETag(event)
	assert.NotEqual(etag1, etag3)

	// Different Updated time should produce different ETag
	event.Updated = now.Add(time.Minute)
	etag4 := eventETag(event)
	assert.NotEqual(etag1, etag4)
	assert.NotEqual(etag3, etag4)
}

func TestCalDAVBackend_eventToICalendarString(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	user := &model.User{
		Id:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	now := time.Now().UTC()
	event := &Event{
		Id:          "event-123",
		Title:       "Test Event",
		Description: "Test description",
		Start:       now,
		End:         now.Add(time.Hour),
		Created:     now,
		Owner:       user.Id,
		Recurrent:   false,
	}

	icalStr := backend.eventToICalendarString(event, user)
	assert.Contains(icalStr, "BEGIN:VCALENDAR")
	assert.Contains(icalStr, "END:VCALENDAR")
	assert.Contains(icalStr, "BEGIN:VEVENT")
	assert.Contains(icalStr, "END:VEVENT")
	assert.Contains(icalStr, "UID:event-123")
	assert.Contains(icalStr, "SUMMARY:Test Event")
	assert.Contains(icalStr, "DESCRIPTION:Test description")
	assert.Contains(icalStr, "ORGANIZER;CN=testuser:mailto:test@example.com")
}

func TestCalDAVBackend_eventToICalendarString_AllDay(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	event := &Event{
		Id:      "event-allday",
		Title:   "All Day Event",
		Start:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		End:     time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
		Created: time.Now().UTC(),
		AllDay:  true,
	}

	icalStr := backend.eventToICalendarString(event, nil)
	assert.Contains(icalStr, "DTSTART;VALUE=DATE:20240115")
	assert.Contains(icalStr, "DTEND;VALUE=DATE:20240116")
	assert.NotContains(icalStr, "DTSTART:20240115T")
}

func TestCalDAVBackend_eventToICalendarString_Recurring(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	now := time.Now().UTC()
	event := &Event{
		Id:         "event-123",
		Title:      "Weekly Meeting",
		Start:      now,
		End:        now.Add(time.Hour),
		Created:    now,
		Owner:      "user-123",
		Recurrent:  true,
		Recurrence: "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
	}

	icalStr := backend.eventToICalendarString(event, nil)
	assert.Contains(icalStr, "RRULE:FREQ=WEEKLY")
	assert.Contains(icalStr, "BYDAY=MO")
}

func TestCalDAVBackend_icalendarToEvent(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	// Create a simple iCalendar using arran4/golang-ical
	cal := ics.NewCalendar()
	cal.SetVersion("2.0")
	cal.SetProductId("-//Test//EN")

	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	vevent := cal.AddEvent("test-uid")
	vevent.SetSummary("Test Event")
	vevent.SetDescription("Test description")
	vevent.SetStartAt(start)
	vevent.SetEndAt(end)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.Equal("event-123", event.Id)
	assert.Equal("Test Event", event.Title)
	assert.Equal("Test description", event.Description)
	assert.Equal(start, event.Start)
	assert.Equal(end, event.End)
	// no CalDAV client offers a visibility picker, so events created from the
	// phone/desktop must default to team-visible rather than private, or
	// nobody else would ever see them
	assert.Equal(VisibilityTeam, event.Visibility)
	assert.False(event.AllDay)
}

func TestCalDAVBackend_icalendarToEvent_AllDay(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	// RFC 5545 VALUE=DATE all-day events carry no "T" time separator
	icalData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-uid
DTSTART;VALUE=DATE:20240115
DTEND;VALUE=DATE:20240116
SUMMARY:Test All Day Event
END:VEVENT
END:VCALENDAR`

	cal, err := ics.ParseCalendar(strings.NewReader(icalData))
	assert.Nil(err)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.True(event.AllDay)
	assert.Equal(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), event.Start)
	assert.Equal(time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC), event.End)
}

func TestCalDAVBackend_icalendarToEvent_WithRRULE(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	cal := ics.NewCalendar()
	start := time.Now().UTC().Truncate(time.Second)

	vevent := cal.AddEvent("test-uid")
	vevent.SetSummary("Recurring Event")
	vevent.SetStartAt(start)
	vevent.SetEndAt(start.Add(time.Hour))
	vevent.AddRrule("FREQ=DAILY;COUNT=10")

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.True(event.Recurrent)
	assert.True(strings.HasPrefix(event.Recurrence, "RRULE:"))
	assert.Contains(event.Recurrence, "FREQ=DAILY")
	assert.Contains(event.Recurrence, "COUNT=10")
}

func TestCalDAVBackend_icalendarToEvent_NoVEVENT(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	cal := ics.NewCalendar()
	cal.SetVersion("2.0")

	_, err := backend.icalendarToEvent(cal, "event-123")
	assert.NotNil(err)
	assert.Contains(err.Error(), "no VEVENT")
}

func TestCalDAVBackend_icalendarToEvent_WithTimezone(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	// Parse iCalendar with TZID parameter (like Apple Calendar sends)
	icalData := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Apple Inc.//macOS//EN
BEGIN:VEVENT
UID:test-uid
DTSTART;TZID=Europe/Moscow:20240115T110000
DTEND;TZID=Europe/Moscow:20240115T120000
SUMMARY:Test Event with Timezone
END:VEVENT
END:VCALENDAR`

	cal, err := ics.ParseCalendar(strings.NewReader(icalData))
	assert.Nil(err)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.Equal("Test Event with Timezone", event.Title)

	// Moscow is UTC+3, so 11:00 Moscow = 08:00 UTC
	expectedStartUTC := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	expectedEndUTC := time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)

	assert.Equal(expectedStartUTC, event.Start)
	assert.Equal(expectedEndUTC, event.End)
}

func TestCalDAVBackend_icalendarToEvent_UTCTime(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token", "#1E90FFFF")

	// Parse iCalendar with UTC time (ends with Z)
	icalData := `BEGIN:VCALENDAR
VERSION:2.0
BEGIN:VEVENT
UID:test-uid
DTSTART:20240115T110000Z
DTEND:20240115T120000Z
SUMMARY:Test Event UTC
END:VEVENT
END:VCALENDAR`

	cal, err := ics.ParseCalendar(strings.NewReader(icalData))
	assert.Nil(err)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)

	// Time should remain as specified (already UTC)
	expectedStart := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	assert.Equal(expectedStart, event.Start)
	assert.Equal(expectedEnd, event.End)
}

func TestServeCalDAV_InvalidToken(t *testing.T) {
	assert := assert.New(t)

	api := plugintest.API{}
	api.On("LogDebug", "Plugin HTTP request", "method", "GET", "path", "/caldav/short/", "user-agent", "").Return()
	api.On("LogInfo", "CalDAV incoming request", "method", "GET", "path", "/caldav/short/", "host", "example.com", "user-agent", "", "hasAuth", false).Return()
	api.On("LogError", "CalDAV invalid token", "token_length", 5).Return()

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
	}
	calPlugin.router = calPlugin.InitAPI()

	// Test with short token
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/caldav/short/", nil)

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)
}

func TestServeCalDAV_NoBasicAuth(t *testing.T) {
	// Authentication is now via token in URL, not Basic Auth
	// This test verifies that without Basic Auth, the request still proceeds to token lookup
	assert := assert.New(t)

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	api := plugintest.API{}
	api.On("LogDebug", "Plugin HTTP request", "method", "GET", "path", "/caldav/"+tokenValue+"/", "user-agent", "").Return()
	api.On("LogInfo", "CalDAV incoming request", "method", "GET", "path", "/caldav/"+tokenValue+"/", "host", "example.com", "user-agent", "", "hasAuth", false).Return()
	api.On("LogError", "ServeCalDAV: token not found: sql: no rows in result set").Return()

	// DB mocks - token not found
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query for token - return empty result
	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(tokenValue).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used", "calendar_color"}))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/caldav/"+tokenValue+"/", nil)
	// No Basic Auth header - should still work, auth is via token in URL

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	// Token not found in DB, so returns 401
	assert.Equal(http.StatusUnauthorized, result.StatusCode)

	api.AssertExpectations(t)
}

func TestServeCalDAV_TokenNotFound(t *testing.T) {
	assert := assert.New(t)

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	api := plugintest.API{}
	api.On("LogDebug", "Plugin HTTP request", "method", "GET", "path", "/caldav/"+tokenValue+"/", "user-agent", "").Return()
	api.On("LogInfo", "CalDAV incoming request", "method", "GET", "path", "/caldav/"+tokenValue+"/", "host", "example.com", "user-agent", "", "hasAuth", true).Return()
	api.On("LogError", "ServeCalDAV: token not found: sql: no rows in result set").Return()

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query for token - return empty result
	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(tokenValue).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used", "calendar_color"}))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/caldav/"+tokenValue+"/", nil)
	r.SetBasicAuth("user", "pass")

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)

	api.AssertExpectations(t)
}

func TestServeCalDAV_PROPFIND(t *testing.T) {
	assert := assert.New(t)

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	api := plugintest.API{}
	// Use mock.Anything for log calls to avoid brittleness
	api.On("LogDebug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	api.On("LogInfo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return().Maybe()
	api.On("GetUser", "test-user").Return(&model.User{
		Id:       "test-user",
		Username: "testuser",
		Email:    "test@example.com",
		Timezone: map[string]string{
			"useAutomaticTimezone": "false",
			"manualTimezone":       "UTC",
		},
	}, nil)
	propfindBody := `<?xml version="1.0" encoding="UTF-8"?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:current-user-principal/>
  </D:prop>
</D:propfind>`

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	createdTime := time.Now().UTC()

	// Mock query for token
	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)

	defaultColor := "#1E90FFFF"
	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(tokenValue).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used", "calendar_color"}).
			AddRow(tokenValue, "test-user", createdTime, nil, defaultColor))

	// Mock update last_used
	updateBuilder := sq.Update("calendar_ical_tokens").
		Set("last_used", sqlmock.AnyArg()).
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()
	dbMock.ExpectExec(regexp.QuoteMeta(updateSql)).
		WithArgs(sqlmock.AnyArg(), tokenValue).
		WillReturnResult(sqlmock.NewResult(0, 1))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PROPFIND", "/caldav/"+tokenValue+"/", strings.NewReader(propfindBody))
	r.Header.Set("Content-Type", "application/xml")
	r.Header.Set("Depth", "0")
	r.SetBasicAuth("user", "pass")

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	bodyBytes, _ := io.ReadAll(result.Body)
	bodyStr := string(bodyBytes)

	// CalDAV PROPFIND should return 207 Multi-Status
	assert.Equal(http.StatusMultiStatus, result.StatusCode)
	assert.Contains(bodyStr, "multistatus")

	api.AssertExpectations(t)
}

func TestGetICalToken_WithCalDAVURL(t *testing.T) {
	assert := assert.New(t)

	ctx := &plugin.Context{
		SessionId: "session-id",
	}

	api := plugintest.API{}
	api.On("LogDebug", "Plugin HTTP request", "method", "GET", "path", "/ical/token", "user-agent", "").Return()
	session := &model.Session{
		UserId: "test-user",
	}
	api.On("GetSession", ctx.SessionId).Return(session, nil)

	siteURL := "https://mattermost.example.com"
	config := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}
	api.On("GetConfig").Return(config)

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"
	createdTime := time.Now().UTC()

	// Mock query for token - return existing token
	queryBuilder := sq.Select("token", "user_id", "created", "last_used", "calendar_color").
		From("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used", "calendar_color"}).
			AddRow(tokenValue, session.UserId, createdTime, nil, "#1E90FFFF"))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ical/token", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)

	assert.Equal(http.StatusOK, result.StatusCode)

	bodyStr := string(bodyBytes)
	assert.Contains(bodyStr, `"caldavUrl":`)
	assert.Contains(bodyStr, "/caldav/"+tokenValue+"/")

	api.AssertExpectations(t)
}

func TestGenerateICalToken_WithCalDAVURL(t *testing.T) {
	assert := assert.New(t)

	ctx := &plugin.Context{
		SessionId: "session-id",
	}

	api := plugintest.API{}
	api.On("LogDebug", "Plugin HTTP request", "method", "POST", "path", "/ical/token", "user-agent", "").Return()
	session := &model.Session{
		UserId: "test-user",
	}
	api.On("GetSession", ctx.SessionId).Return(session, nil)

	siteURL := "https://mattermost.example.com"
	config := &model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}
	api.On("GetConfig").Return(config)

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock delete query (delete existing token)
	deleteBuilder := sq.Delete("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(sq.Dollar)
	deleteSql, _, _ := deleteBuilder.ToSql()
	dbMock.ExpectExec(regexp.QuoteMeta(deleteSql)).
		WithArgs(session.UserId).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Mock insert query
	dbMock.ExpectExec(regexp.QuoteMeta("INSERT INTO calendar_ical_tokens")).
		WithArgs(sqlmock.AnyArg(), session.UserId, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/ical/token", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)

	assert.Equal(http.StatusOK, result.StatusCode)

	bodyStr := string(bodyBytes)
	assert.Contains(bodyStr, `"enabled":true`)
	assert.Contains(bodyStr, `"token":`)
	assert.Contains(bodyStr, `"url":`)
	assert.Contains(bodyStr, `"caldavUrl":`)
	assert.Contains(bodyStr, "/caldav/")

	api.AssertExpectations(t)
}

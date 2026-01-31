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
	"github.com/emersion/go-ical"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

func TestCalDAVBackend_extractEventID(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

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
		Recurrent:   false,
		Recurrence:  "",
	}

	etag1 := eventETag(event)
	assert.NotEmpty(etag1)

	// Same event should produce same ETag
	etag2 := eventETag(event)
	assert.Equal(etag1, etag2)

	// Modified event should produce different ETag
	event.Title = "Modified Title"
	etag3 := eventETag(event)
	assert.NotEqual(etag1, etag3)

	// Changing time should change ETag
	event.Title = "Test Event"
	event.Start = now.Add(time.Minute)
	etag4 := eventETag(event)
	assert.NotEqual(etag1, etag4)
}

func TestCalDAVBackend_eventToICalendarString(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

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
	assert.Contains(icalStr, "ORGANIZER:mailto:test@example.com")
}

func TestCalDAVBackend_eventToICalendarString_Recurring(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

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
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

	// Create a simple iCalendar
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")
	cal.Props.SetText(ical.PropProductID, "-//Test//EN")

	vevent := ical.NewComponent(ical.CompEvent)
	vevent.Props.SetText(ical.PropUID, "test-uid")
	vevent.Props.SetText(ical.PropSummary, "Test Event")
	vevent.Props.SetText(ical.PropDescription, "Test description")

	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)
	vevent.Props.SetDateTime(ical.PropDateTimeStart, start)
	vevent.Props.SetDateTime(ical.PropDateTimeEnd, end)

	cal.Children = append(cal.Children, vevent)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.Equal("event-123", event.Id)
	assert.Equal("Test Event", event.Title)
	assert.Equal("Test description", event.Description)
	assert.Equal(start, event.Start)
	assert.Equal(end, event.End)
	assert.Equal(VisibilityPrivate, event.Visibility)
}

func TestCalDAVBackend_icalendarToEvent_WithRRULE(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

	cal := ical.NewCalendar()
	vevent := ical.NewComponent(ical.CompEvent)
	vevent.Props.SetText(ical.PropSummary, "Recurring Event")
	vevent.Props.SetText(ical.PropRecurrenceRule, "FREQ=DAILY;COUNT=10")

	start := time.Now().UTC()
	vevent.Props.SetDateTime(ical.PropDateTimeStart, start)
	vevent.Props.SetDateTime(ical.PropDateTimeEnd, start.Add(time.Hour))

	cal.Children = append(cal.Children, vevent)

	event, err := backend.icalendarToEvent(cal, "event-123")
	assert.Nil(err)
	assert.True(event.Recurrent)
	// go-ical library may escape semicolons, so check for presence of RRULE prefix and key parts
	assert.True(strings.HasPrefix(event.Recurrence, "RRULE:"))
	assert.Contains(event.Recurrence, "FREQ=DAILY")
	assert.Contains(event.Recurrence, "COUNT=10")
}

func TestCalDAVBackend_icalendarToEvent_NoVEVENT(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropVersion, "2.0")

	_, err := backend.icalendarToEvent(cal, "event-123")
	assert.NotNil(err)
	assert.Contains(err.Error(), "no VEVENT")
}

func TestServeCalDAV_InvalidToken(t *testing.T) {
	assert := assert.New(t)

	calPlugin := Plugin{}
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
	assert := assert.New(t)

	calPlugin := Plugin{}
	calPlugin.router = calPlugin.InitAPI()

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/caldav/"+tokenValue+"/", nil)
	// No Basic Auth header

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)
	assert.Contains(result.Header.Get("WWW-Authenticate"), "Basic")
}

func TestServeCalDAV_TokenNotFound(t *testing.T) {
	assert := assert.New(t)

	api := plugintest.API{}
	api.On("LogError", "ServeCalDAV: token not found: sql: no rows in result set").Return()

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"

	// Mock query for token - return empty result
	queryBuilder := sq.Select("token", "user_id", "created", "last_used").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(tokenValue).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used"}))

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

	api := plugintest.API{}
	api.On("GetUser", "test-user").Return(&model.User{
		Id:       "test-user",
		Username: "testuser",
		Email:    "test@example.com",
		Timezone: map[string]string{
			"useAutomaticTimezone": "false",
			"manualTimezone":       "UTC",
		},
	}, nil)
	api.On("LogInfo", "CalDAV request", "method", "PROPFIND", "path", "/caldav/abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234/").Return()

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	tokenValue := "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234"
	createdTime := time.Now().UTC()

	// Mock query for token
	queryBuilder := sq.Select("token", "user_id", "created", "last_used").
		From("calendar_ical_tokens").
		Where(sq.Eq{"token": tokenValue}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(tokenValue).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used"}).
			AddRow(tokenValue, "test-user", createdTime, nil))

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

	propfindBody := `<?xml version="1.0" encoding="UTF-8"?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:current-user-principal/>
  </D:prop>
</D:propfind>`

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
	queryBuilder := sq.Select("token", "user_id", "created", "last_used").
		From("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used"}).
			AddRow(tokenValue, session.UserId, createdTime, nil))

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

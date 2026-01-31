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

func TestCalDAVBackend_CurrentUserPrincipal(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token-1234567890123456789012345678901234567890123456789012")

	principal, err := backend.CurrentUserPrincipal(nil)
	assert.Nil(err)
	assert.Contains(principal, "/caldav/")
	assert.Contains(principal, "test-token-1234567890123456789012345678901234567890123456789012")
}

func TestCalDAVBackend_CalendarHomeSetPath(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token-1234567890123456789012345678901234567890123456789012")

	path, err := backend.CalendarHomeSetPath(nil)
	assert.Nil(err)
	assert.Contains(path, "/caldav/")
	assert.Contains(path, "test-token-1234567890123456789012345678901234567890123456789012")
}

func TestCalDAVBackend_ListCalendars(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token-1234567890123456789012345678901234567890123456789012")

	calendars, err := backend.ListCalendars(nil)
	assert.Nil(err)
	assert.Equal(1, len(calendars))
	assert.Equal("Mattermost Calendar", calendars[0].Name)
	assert.Contains(calendars[0].Path, "/calendar/")
	assert.Contains(calendars[0].SupportedComponentSet, "VEVENT")
}

func TestCalDAVBackend_GetCalendar(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token-1234567890123456789012345678901234567890123456789012")

	// Test valid calendar path
	calendar, err := backend.GetCalendar(nil, "/plugins/com.dmkir.calendar/caldav/token/calendar/")
	assert.Nil(err)
	assert.Equal("Mattermost Calendar", calendar.Name)

	// Test invalid path
	_, err = backend.GetCalendar(nil, "/invalid/path")
	assert.NotNil(err)
}

func TestCalDAVBackend_CreateCalendar(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token-1234567890123456789012345678901234567890123456789012")

	err := backend.CreateCalendar(nil, nil)
	assert.NotNil(err, "Creating additional calendars should not be supported")
}

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

func TestCalDAVBackend_generateETag(t *testing.T) {
	assert := assert.New(t)

	calPlugin := &Plugin{}
	backend := NewCalDAVBackend(calPlugin, "user-123", "test-token")

	created := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	event := &Event{
		Id:      "event-123",
		Created: created,
	}

	etag := backend.generateETag(event)
	assert.Contains(etag, `"`)
	assert.NotEmpty(etag)

	// Same event should produce same ETag
	etag2 := backend.generateETag(event)
	assert.Equal(etag, etag2)
}

func TestCalDAVBackend_eventToICalendar(t *testing.T) {
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

	cal := backend.eventToICalendar(event, user)
	assert.NotNil(cal)

	// Verify calendar properties
	version := cal.Props.Get(ical.PropVersion)
	assert.NotNil(version)
	assert.Equal("2.0", version.Value)

	// Verify VEVENT component
	assert.Equal(1, len(cal.Children))
	vevent := cal.Children[0]
	assert.Equal(ical.CompEvent, vevent.Name)

	// Verify event properties
	uid := vevent.Props.Get(ical.PropUID)
	assert.NotNil(uid)
	assert.Equal("event-123", uid.Value)

	summary := vevent.Props.Get(ical.PropSummary)
	assert.NotNil(summary)
	assert.Equal("Test Event", summary.Value)

	desc := vevent.Props.Get(ical.PropDescription)
	assert.NotNil(desc)
	assert.Equal("Test description", desc.Value)
}

func TestCalDAVBackend_eventToICalendar_Recurring(t *testing.T) {
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

	cal := backend.eventToICalendar(event, nil)
	assert.NotNil(cal)

	vevent := cal.Children[0]
	rrule := vevent.Props.Get(ical.PropRecurrenceRule)
	assert.NotNil(rrule)
	// go-ical library escapes semicolons, so we check for the unescaped content
	assert.Contains(rrule.Value, "FREQ=WEEKLY")
	assert.Contains(rrule.Value, "INTERVAL=1")
	assert.Contains(rrule.Value, "BYDAY=MO")
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

func TestCalDAVBackend_ListCalendarObjects(t *testing.T) {
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
	api.On("GetTeamsForUser", "test-user").Return([]*model.Team{
		{Id: "team-1", DisplayName: "Test Team"},
	}, nil)

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock events query - return some events
	now := time.Now().UTC()
	dbMock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "title", "description", "dt_start", "dt_end",
			"created", "owner", "channel", "recurrent", "recurrence",
			"color", "team", "visibility", "alert", "alert_time",
		}).
			AddRow("event-1", "Test Event", "Description", now, now.Add(time.Hour),
				now, "test-user", nil, false, "",
				"#D0D0D0", "team-1", "private", "", nil))

	// Mock channel query
	dbMock.ExpectQuery("SELECT ChannelId").
		WillReturnRows(sqlmock.NewRows([]string{"ChannelId"}))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}

	backend := NewCalDAVBackend(&calPlugin, "test-user", "test-token")
	objects, err := backend.ListCalendarObjects(nil, "/calendar/", nil)

	assert.Nil(err)
	assert.Equal(1, len(objects))
	assert.Equal("event-1", backend.extractEventID(objects[0].Path))
	assert.NotNil(objects[0].Data)

	api.AssertExpectations(t)
}

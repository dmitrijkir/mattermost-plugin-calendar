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
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSecureToken(t *testing.T) {
	assert := assert.New(t)

	token1, err1 := generateSecureToken()
	assert.Nil(err1)
	assert.Equal(64, len(token1), "Token should be 64 characters (32 bytes hex encoded)")

	token2, err2 := generateSecureToken()
	assert.Nil(err2)
	assert.Equal(64, len(token2))

	// Tokens should be unique
	assert.NotEqual(token1, token2, "Two generated tokens should be different")

	// Token should be valid hex
	for _, c := range token1 {
		assert.True((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"Token should only contain hex characters")
	}
}

func TestFormatDurationForICS(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"5 minutes", 5 * time.Minute, "5M"},
		{"15 minutes", 15 * time.Minute, "15M"},
		{"30 minutes", 30 * time.Minute, "30M"},
		{"1 hour", 1 * time.Hour, "1H"},
		{"2 hours", 2 * time.Hour, "2H"},
		{"1 hour 30 minutes", 90 * time.Minute, "1H30M"},
		{"1 day", 24 * time.Hour, "1D"},
		{"2 days", 48 * time.Hour, "2D"},
		{"1 day 2 hours", 26 * time.Hour, "1DT2H"},
		{"1 week", 7 * 24 * time.Hour, "7D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDurationForICS(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetICalToken_NotFound(t *testing.T) {
	assert := assert.New(t)

	ctx := &plugin.Context{
		SessionId: "session-id",
	}

	api := plugintest.API{}
	session := &model.Session{
		UserId: "test-user",
	}
	api.On("GetSession", ctx.SessionId).Return(session, nil)

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock query for token - return empty result
	queryBuilder := sq.Select("token", "user_id", "created", "last_used").
		From("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(session.UserId).
		WillReturnRows(sqlmock.NewRows([]string{"token", "user_id", "created", "last_used"}))

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
	assert.JSONEq(`{"data":{"enabled":false}}`, string(bodyBytes))

	api.AssertExpectations(t)
}

func TestGetICalToken_Found(t *testing.T) {
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

	expectedURL := siteURL + "/plugins/" + PluginId + "/ical/feed/" + tokenValue
	expectedJSON := `{"data":{"token":"` + tokenValue + `","url":"` + expectedURL + `","enabled":true}}`
	assert.JSONEq(expectedJSON, string(bodyBytes))

	api.AssertExpectations(t)
}

func TestGenerateICalToken(t *testing.T) {
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

	// Check response contains expected fields
	bodyStr := string(bodyBytes)
	assert.Contains(bodyStr, `"enabled":true`)
	assert.Contains(bodyStr, `"token":`)
	assert.Contains(bodyStr, `"url":`)
	assert.Contains(bodyStr, siteURL)

	api.AssertExpectations(t)
}

func TestRevokeICalToken(t *testing.T) {
	assert := assert.New(t)

	ctx := &plugin.Context{
		SessionId: "session-id",
	}

	api := plugintest.API{}
	session := &model.Session{
		UserId: "test-user",
	}
	api.On("GetSession", ctx.SessionId).Return(session, nil)

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	// Mock delete query
	deleteBuilder := sq.Delete("calendar_ical_tokens").
		Where(sq.Eq{"user_id": session.UserId}).
		PlaceholderFormat(sq.Dollar)
	deleteSql, _, _ := deleteBuilder.ToSql()
	dbMock.ExpectExec(regexp.QuoteMeta(deleteSql)).
		WithArgs(session.UserId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/ical/token", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)

	assert.Equal(http.StatusOK, result.StatusCode)
	assert.JSONEq(`{"data":{"enabled":false}}`, string(bodyBytes))

	api.AssertExpectations(t)
}

func TestServeICalFeed_InvalidToken(t *testing.T) {
	assert := assert.New(t)

	calPlugin := Plugin{}
	calPlugin.router = calPlugin.InitAPI()

	// Test with short token
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ical/feed/short", nil)

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)
}

func TestServeICalFeed_TokenNotFound(t *testing.T) {
	assert := assert.New(t)

	api := plugintest.API{}
	api.On("LogError", "ServeICalFeed: token not found: sql: no rows in result set").Return()

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
	r := httptest.NewRequest(http.MethodGet, "/ical/feed/"+tokenValue, nil)

	calPlugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)

	api.AssertExpectations(t)
}

func TestGenerateICalendar(t *testing.T) {
	assert := assert.New(t)

	api := plugintest.API{}
	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
	}

	user := &model.User{
		Id:       "user-1",
		Username: "testuser",
		Email:    "test@example.com",
	}

	now := time.Now().UTC()
	events := []Event{
		{
			Id:          "event-1",
			Title:       "Test Event 1",
			Description: "Test description",
			Start:       now,
			End:         now.Add(time.Hour),
			Created:     now,
			Owner:       user.Id,
			Recurrent:   false,
		},
		{
			Id:          "event-2",
			Title:       "Recurring Event",
			Description: "",
			Start:       now,
			End:         now.Add(30 * time.Minute),
			Created:     now,
			Owner:       user.Id,
			Recurrent:   true,
			Recurrence:  "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
			Alert:       EventAlert15MinutesBefore,
		},
	}

	icalContent := calPlugin.generateICalendar(events, user)

	// Check iCalendar structure
	assert.Contains(icalContent, "BEGIN:VCALENDAR")
	assert.Contains(icalContent, "END:VCALENDAR")
	assert.Contains(icalContent, "VERSION:2.0")
	assert.Contains(icalContent, "PRODID:-//Mattermost Calendar Plugin//EN")
	assert.Contains(icalContent, "X-WR-CALNAME:Mattermost Calendar")

	// Check events
	assert.Contains(icalContent, "BEGIN:VEVENT")
	assert.Contains(icalContent, "END:VEVENT")
	assert.Contains(icalContent, "SUMMARY:Test Event 1")
	assert.Contains(icalContent, "SUMMARY:Recurring Event")
	assert.Contains(icalContent, "DESCRIPTION:Test description")

	// Check RRULE for recurring event
	assert.Contains(icalContent, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO")

	// Check alarm
	assert.Contains(icalContent, "BEGIN:VALARM")
	assert.Contains(icalContent, "END:VALARM")
	assert.Contains(icalContent, "ACTION:DISPLAY")

	// Check organizer
	assert.Contains(icalContent, "ORGANIZER")
	assert.Contains(icalContent, user.Email)

	// Count events (should have 2)
	eventCount := strings.Count(icalContent, "BEGIN:VEVENT")
	assert.Equal(2, eventCount)
}

func TestGetICalToken_NotAuthorized(t *testing.T) {
	assert := assert.New(t)

	ctx := &plugin.Context{
		SessionId: "invalid-session",
	}

	api := plugintest.API{}
	api.On("GetSession", ctx.SessionId).Return(nil, model.NewAppError("", "", nil, "", http.StatusUnauthorized))
	api.On("LogError", "can't get session").Return()

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API: &api,
		},
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ical/token", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()

	assert.Equal(http.StatusUnauthorized, result.StatusCode)

	api.AssertExpectations(t)
}

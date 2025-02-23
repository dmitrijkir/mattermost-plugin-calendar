package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	plugin.router = plugin.InitAPI()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	plugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)
	bodyString := string(bodyBytes)

	assert.Equal("404 page not found\n", bodyString)
}

func TestGetEvents(t *testing.T) {
	ctx := &plugin.Context{
		AcceptLanguage: "EN",
		IPAddress:      "",
		RequestId:      "",
		SessionId:      "user-id",
		UserAgent:      "test",
	}

	api := plugintest.API{}

	session := &model.Session{
		UserId: "test-user",
	}
	user := &model.User{
		Id: "test-user",
		Timezone: map[string]string{
			"manualTimezone": "Europe/Moscow",
		},
	}
	api.On("GetSession", ctx.SessionId).Return(session, nil)
	api.On("GetUser", session.UserId).Return(user, nil)
	api.On("GetTeamsForUser", "test-user").Return([]*model.Team{}, nil)

	// DB mocks

	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	sqlTimeStart := time.Date(2023, time.February, 26, 21, 0, 0, 0, time.UTC)
	sqlTimeEnd := time.Date(2023, time.March, 05, 21, 0, 0, 0, time.UTC)

	// Create a new select builder
	queryBuilderUsersInChannel := sq.Select().
		Columns("ChannelId").
		From("ChannelMembers").
		Where(sq.Eq{"userid": session.UserId}).
		PlaceholderFormat(sq.Dollar)

	querySqlUsersInChannel, _, _ := queryBuilderUsersInChannel.ToSql()
	expectedQueryUsersInChannel := dbMock.ExpectQuery(
		regexp.QuoteMeta(querySqlUsersInChannel),
	).WithArgs(session.UserId)
	expectedQueryUsersInChannel.WillReturnRows(sqlmock.NewRows([]string{"channelid"}).AddRow("channel-1"))

	// Create a new select builder
	conditions := sq.And{
		sq.Or{
			sq.Eq{"cm.member": session.UserId},
			sq.Eq{"ce.owner": session.UserId},
			sq.NotEq{"ce.visibility": string(VisibilityPrivate)},
		},
		sq.Or{
			sq.And{
				sq.GtOrEq{"ce.dt_start": sqlTimeStart},
				sq.LtOrEq{"ce.dt_start": sqlTimeEnd},
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
			"ce.owner",
			"ce.channel",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.team",
			"ce.visibility",
			"ce.alert",
			"ce.alert_time",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(conditions).PlaceholderFormat(sq.Dollar)

	expectedQuerySql, _, err := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(expectedQuerySql)).
		WithArgs(session.UserId, session.UserId, "private", sqlTimeStart, sqlTimeEnd, true)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"description",
		"dt_start",
		"dt_end",
		"created",
		"owner",
		"channel",
		"recurrent",
		"recurrence",
		"color",
		"team",
		"visibility",
		"alert",
		"alert_time",
	},
	).AddRow("event-1", "test event 1", "", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", false, "", nil, "team1", "private", "", nil).AddRow("event-2", "test event 2", "", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", false, "", "#D0D0D0", "team1", "private", "", nil).AddRow("event-3", "test event 3", "", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0", "team1", "private", "", nil).AddRow("event-3", "test event 3 another user", "", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0", "team1", "private", "", nil).AddRow("event-3", "test event 3 another user", "", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0", "team1", "private", "", nil)

	expectedQuery.WillReturnRows(eventsRow)

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
		DB: dbx,
	}
	calPlugin.router = calPlugin.InitAPI()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/events?start=2023-02-27T00:00:00&end=2023-03-06T00:00:00", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	assert := assert.New(t)
	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)

	expectedResponse := `{"data":[{"id":"event-1","title":"test event 1","description":"",
						"start":"2023-02-27T00:00:00+03:00","end":"2023-03-06T00:00:00+03:00",
						"attendees":null,"created":"2023-03-05T21:00:00Z","owner":"owner_id","team":"team1",
						"channel":"channel-id","recurrence":"","color":"#D0D0D0","visibility":"private","alert":"",
						"alertTime":null},{"id":"event-2","title":"test event 2","description":"",
						"start":"2023-02-27T00:00:00+03:00","end":"2023-03-06T00:00:00+03:00","attendees":null,
						"created":"2023-03-05T21:00:00Z","owner":"owner_id","team":"team1","channel":"channel-id",
						"recurrence":"","color":"#D0D0D0","visibility":"private","alert":"","alertTime":null},
						{"id":"event-3","title":"test event 3","description":"","start":"2023-02-27T00:00:00+03:00",
						"end":"2023-03-06T00:00:00+03:00","attendees":null,"created":"2023-03-05T21:00:00Z",
						"owner":"owner_id","team":"team1","channel":"channel-id",
						"recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE","color":"#D0D0D0",
						"visibility":"private","alert":"","alertTime":null},{"id":"event-3","title":"test event 3",
						"description":"","start":"2023-02-28T00:00:00+03:00","end":"2023-03-07T00:00:00+03:00",
						"attendees":null,"created":"2023-03-05T21:00:00Z","owner":"owner_id","team":"team1",
						"channel":"channel-id","recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE",
						"color":"#D0D0D0","visibility":"private","alert":"","alertTime":null},{"id":"event-3",
						"title":"test event 3","description":"","start":"2023-03-01T00:00:00+03:00",
						"end":"2023-03-08T00:00:00+03:00","attendees":null,"created":"2023-03-05T21:00:00Z",
						"owner":"owner_id","team":"team1","channel":"channel-id",
						"recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE","color":"#D0D0D0",
						"visibility":"private","alert":"","alertTime":null}]}`
	assert.JSONEq(string(bodyBytes), expectedResponse)
	api.AssertExpectations(t)
}

package main

import (
	"github.com/DATA-DOG/go-sqlmock"
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
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	plugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)
	bodyString := string(bodyBytes)

	assert.Equal("", bodyString)
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

	// DB mocks

	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	utcLoc, _ := time.LoadLocation("UTC")
	sqlTimeStart := time.Date(2023, time.February, 26, 21, 0, 0, 0, utcLoc)
	sqlTimeEnd := time.Date(2023, time.March, 05, 21, 0, 0, 0, utcLoc)

	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT ce.id,
			  	   ce.title,
				   ce."start",
				   ce."end",
				   ce.created,
				   ce."owner",
				   ce."channel",
				   ce.recurrent,
				   ce.recurrence,
				   ce.color
		    FROM calendar_events ce
				 FULL JOIN calendar_members cm 
					    ON ce.id = cm."event"
		    WHERE (cm."user" = $1 OR ce."owner" = $2)
				 AND (
					  (ce."start" >= $3 AND ce."start" <= $4) 
						  or ce.recurrent = true
					 )
			`)).WithArgs(session.UserId, session.UserId, sqlTimeStart, sqlTimeEnd)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"start",
		"end",
		"created",
		"owner",
		"channel",
		"recurrent",
		"recurrence",
		"color",
	},
	).AddRow("event-1", "test event 1", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", false, "", nil).AddRow("event-2", "test event 2", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", false, "", "#D0D0D0").AddRow("event-3", "test event 3", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0").AddRow("event-3", "test event 3 another user", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0").AddRow("event-3", "test event 3 another user", sqlTimeStart, sqlTimeEnd, sqlTimeEnd,
		"owner_id", "channel-id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE", "#D0D0D0")

	expectedQuery.WillReturnRows(eventsRow)

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
		DB: dbx,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/events?start=2023-02-27T00:00:00&end=2023-03-06T00:00:00", nil)

	calPlugin.ServeHTTP(ctx, w, r)

	assert := assert.New(t)
	result := w.Result()
	assert.NotNil(result)
	defer result.Body.Close()
	bodyBytes, err := io.ReadAll(result.Body)
	assert.Nil(err)

	expectedResponse := `{"data":[{"id":"event-1","title":"test event 1","start":"2023-02-27T00:00:00+03:00",
						 "end":"2023-03-06T00:00:00+03:00","attendees":null,"created":"2023-03-05T21:00:00Z",
						 "owner":"owner_id","channel":"channel-id","recurrence":"","color":"#D0D0D0"},{"id":"event-2",
						 "title":"test event 2","start":"2023-02-27T00:00:00+03:00","end":"2023-03-06T00:00:00+03:00",
						 "attendees":null,"created":"2023-03-05T21:00:00Z","owner":"owner_id","channel":"channel-id",
						 "recurrence":"","color":"#D0D0D0"},{"id":"event-3","title":"test event 3",
					     "start":"2023-02-27T00:00:00+03:00","end":"2023-03-06T00:00:00+03:00","attendees":null,
						 "created":"2023-03-05T21:00:00Z","owner":"owner_id","channel":"channel-id",
						 "recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE","color":"#D0D0D0"},
						 {"id":"event-3","title":"test event 3",
						 "start":"2023-02-28T00:00:00+03:00","end":"2023-03-07T00:00:00+03:00","attendees":null,
						 "created":"2023-03-05T21:00:00Z","owner":"owner_id","channel":"channel-id",
						 "recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE","color":"#D0D0D0"},
						 {"id":"event-3","title":"test event 3",
						 "start":"2023-03-01T00:00:00+03:00","end":"2023-03-08T00:00:00+03:00","attendees":null,
						 "created":"2023-03-05T21:00:00Z","owner":"owner_id","channel":"channel-id",
						 "recurrence":"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE","color":"#D0D0D0"}]}`
	assert.JSONEq(string(bodyBytes), expectedResponse)
	api.AssertExpectations(t)
}

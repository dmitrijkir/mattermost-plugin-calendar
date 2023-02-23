package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"regexp"
	"testing"
	"time"
)

// Test personal notification
func TestSendGroupOrPersonalEventNotification(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	testEvent := &Event{
		Id:        "efe-fe",
		Title:     "test event for channel",
		Start:     time.Now(),
		End:       time.Now(),
		Attendees: []string{},
		Created:   time.Now(),
		Owner:     "owner-id",
		Channel:   &channelId,
		Processed: nil,
		Recurrent: false,
	}

	foundChannel := &model.Channel{
		Id: channelId,
	}

	postForSend := &model.Post{
		UserId:    botId,
		Message:   testEvent.Title,
		ChannelId: channelId,
	}

	api := plugintest.API{}

	api.On("GetDirectChannel", testEvent.Owner, botId).Return(foundChannel, nil)
	api.On("CreatePost", postForSend).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	background := NewBackgroundJob(pluginT, botId, dbx)
	background.sendGroupOrPersonalEventNotification(testEvent)

}

// Test group notification
func TestSendGroupOrPersonalEventGroupNotification(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	attendees := []string{"first-id", "second-id"}

	testEvent := &Event{
		Id:        "efe-fe",
		Title:     "test event for channel",
		Start:     time.Now(),
		End:       time.Now(),
		Attendees: attendees,
		Created:   time.Now(),
		Owner:     "owner-id",
		Channel:   &channelId,
		Processed: nil,
		Recurrent: false,
	}

	foundChannel := &model.Channel{
		Id: channelId,
	}

	postForSend := &model.Post{
		UserId:    botId,
		Message:   testEvent.Title,
		ChannelId: channelId,
	}

	api := plugintest.API{}

	attendees = append(attendees, testEvent.Owner)
	attendees = append(attendees, botId)

	api.On("GetGroupChannel", attendees).Return(foundChannel, nil)
	api.On("CreatePost", postForSend).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	background := NewBackgroundJob(pluginT, botId, nil)

	background.sendGroupOrPersonalEventNotification(testEvent)
}

// process event with channel
func TestProcessEventWithChannel(t *testing.T) {

	botId := "bot-id"
	api := plugintest.API{}

	postForSendChannel := &model.Post{
		UserId:    botId,
		Message:   "test event",
		ChannelId: "channel-id",
	}

	api.On("CreatePost", postForSendChannel).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	processingTime := time.Now()

	utcLoc, _ := time.LoadLocation("UTC")

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		utcLoc,
	)

	background := NewBackgroundJob(pluginT, botId, dbx)

	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT ce.id,
				   ce.title,
                   ce."start",
                   ce."end",
                   ce.created,
                   ce."owner",
                   ce."channel",
                   cm."user",
                   ce.recurrent,
                   ce.recurrence
			FROM   calendar_events ce
                FULL JOIN calendar_members cm
                       ON ce.id = cm."event"
			WHERE (ce."start" = $1 OR (ce.recurrent = true AND ce."start"::time = $2)) 
			  	   AND (ce."processed" isnull OR ce."processed" != $3)
			`)).WithArgs(sqlQueryTime, sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"start",
		"end",
		"created",
		"owner",
		"channel",
		"user",
		"recurrent",
		"recurrence"},
	).AddRow("qwcw", "test event", sqlQueryTime, sqlQueryTime, sqlQueryTime,
		"owner_id", "channel-id", "user-Id", false, "[]")

	expectedQuery.WillReturnRows(eventsRow)

	dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE PUBLIC.calendar_events
                                           SET processed = ?
                                           WHERE id = ?`)).WithArgs(sqlQueryTime, "qwcw").
		WillReturnResult(sqlmock.NewResult(0, 1))

	background.process(&processingTime)

}

// process event without channel
func TestProcessEventWithoutChannel(t *testing.T) {
	botId := "bot-id"

	api := plugintest.API{}

	postForSendGroup := &model.Post{
		UserId:    botId,
		Message:   "tests event without channel",
		ChannelId: "channel-id",
	}

	foundChannel := &model.Channel{
		Id: "channel-id",
	}

	api.On("GetGroupChannel", []string{"user-id", "owner-id", "bot-id"}).Return(foundChannel, nil)
	api.On("CreatePost", postForSendGroup).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	processingTime := time.Now()

	utcLoc, _ := time.LoadLocation("UTC")

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		utcLoc,
	)

	background := NewBackgroundJob(pluginT, botId, dbx)

	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(`
			SELECT ce.id,
				   ce.title,
                   ce."start",
                   ce."end",
                   ce.created,
                   ce."owner",
                   ce."channel",
                   cm."user",
                   ce.recurrent,
                   ce.recurrence
			FROM   calendar_events ce
                FULL JOIN calendar_members cm
                       ON ce.id = cm."event"
			WHERE (ce."start" = $1 OR (ce.recurrent = true AND ce."start"::time = $2)) 
			  	   AND (ce."processed" isnull OR ce."processed" != $3)`)).WithArgs(sqlQueryTime,
		sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"start",
		"end",
		"created",
		"owner",
		"channel",
		"user",
		"recurrent",
		"recurrence"},
	).AddRow("qwert-2", "tests event without channel", sqlQueryTime, sqlQueryTime, sqlQueryTime,
		"owner-id", nil, "user-id", false, "[]")

	expectedQuery.WillReturnRows(eventsRow)

	dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE PUBLIC.calendar_events
                                           SET processed = ?
                                           WHERE id = ?`)).
		WithArgs(sqlQueryTime, "qwert-2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	background.process(&processingTime)

}

package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"regexp"
	"strings"
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
		ChannelId: channelId,
	}

	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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

	pluginT.SetDB(dbx)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	postForSend.SetProps(background.getMessageProps(testEvent, time.Now()))

	api.On("GetDirectChannel", testEvent.Owner, botId).Return(foundChannel, nil)
	api.On("CreatePost", postForSend).Return(nil, nil)

	background.sendGroupOrPersonalEventNotification(testEvent, time.Now())
	api.AssertExpectations(t)

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
		ChannelId: channelId,
	}

	api := &plugintest.API{}

	attendees = append(attendees, testEvent.Owner)
	attendees = append(attendees, botId)
	api.On("GetUser", "first-id").Return(&model.User{
		Username: "userName",
	}, nil)
	api.On("GetUser", "second-id").Return(&model.User{
		Username: "userName",
	}, nil)
	api.On("GetGroupChannel", attendees).Return(foundChannel, nil)

	pluginT := &Plugin{
		BotId: botId,
		MattermostPlugin: plugin.MattermostPlugin{
			API:    api,
			Driver: nil,
		},
	}

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	postForSend.SetProps(background.getMessageProps(testEvent, time.Now()))

	api.On("CreatePost", postForSend).Return(nil, nil)

	background.sendGroupOrPersonalEventNotification(testEvent, time.Now())

	api.AssertExpectations(t)
}

// process event with channel notification
func TestProcessEventWithChannel(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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

	pluginT.SetDB(dbx)

	processingTime := time.Now().In(time.UTC)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	postForSendChannel := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
	}

	api.On("CreatePost", postForSendChannel).Return(nil, nil)
	api.On("GetUser", "user-Id").Return(&model.User{
		Username: "userName",
	}, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "qwcw",
		"title":   "test event",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "user-Id",
	}).Return(nil, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "qwcw",
		"title":   "test event",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "owner_id",
	}).Return(nil, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:        "qwcw",
		Title:     "test event",
		Attendees: []string{"user-Id"},
	}

	postForSendChannel.SetProps(background.getMessageProps(testEvent, time.Now()))

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(false, sqlQueryTime, true, sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time",
		"type",
		"meeting_link",
		"mention",
	},
	).AddRow("qwcw", "test event", sqlQueryTime, sqlQueryTime, sqlQueryTime,
		sqlQueryTime, "owner_id", channelId, "user-Id", false, "", "team1", "", nil, "call", nil, "")

	expectedQuery.WillReturnRows(eventsRow)

	updateBuilder := sq.Update("calendar_events").
		Set("processed", sqlQueryTime).
		Where(sq.Eq{"id": "qwcw"}).PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()

	expectedQueryUpdate := dbMock.ExpectQuery(regexp.QuoteMeta(updateSql)).WithArgs(sqlQueryTime, "qwcw")
	expectedQueryUpdate.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("qwcw"))

	background.process(processingTime)

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)

}

// the mention picked on the event has to end up as real post text, otherwise
// Mattermost doesn't notify anyone
func TestProcessEventWithChannelUsesSelectedMention(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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

	pluginT.SetDB(dbx)

	processingTime := time.Now().In(time.UTC)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	postForSendChannel := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
		Message:   "@all",
	}

	api.On("CreatePost", postForSendChannel).Return(nil, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "qw-no-att",
		"title":   "test event no attendees",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "owner_id",
	}).Return(nil, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:    "qw-no-att",
		Title: "test event no attendees",
	}

	postForSendChannel.SetProps(background.getMessageProps(testEvent, time.Now()))

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(false, sqlQueryTime, true, sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time",
		"type",
		"meeting_link",
		"mention",
	},
	).AddRow("qw-no-att", "test event no attendees", sqlQueryTime, sqlQueryTime, sqlQueryTime,
		sqlQueryTime, "owner_id", channelId, nil, false, "", "team1", "", nil, "call", nil, "all")

	expectedQuery.WillReturnRows(eventsRow)

	updateBuilder := sq.Update("calendar_events").
		Set("processed", sqlQueryTime).
		Where(sq.Eq{"id": "qw-no-att"}).PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()

	expectedQueryUpdate := dbMock.ExpectQuery(regexp.QuoteMeta(updateSql)).WithArgs(sqlQueryTime, "qw-no-att")
	expectedQueryUpdate.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("qw-no-att"))

	background.process(processingTime)

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)
}

// process recurrent event
func TestProcessEventWithChannelRecurrent(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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

	pluginT.SetDB(dbx)

	processingTime := time.Now().In(time.UTC)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	featureTime := sqlQueryTime.Add(time.Hour * 24 * 4)

	recurrentEventTimeStart := time.Date(
		2023,
		time.February,
		26,
		21,
		0,
		0,
		0,
		time.UTC,
	)
	recurrentEventTimeEnd := time.Date(
		2023,
		time.February,
		26,
		22,
		0,
		0,
		0,
		time.UTC,
	)

	postForSendChannel := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
	}
	api.On(
		"PublishWebSocketEvent",
		"event_occur",
		map[string]interface{}{
			"id":      "rec-ev",
			"title":   "test event recevent",
			"channel": nil,
		},
		&model.WebsocketBroadcast{UserId: "user-Id"},
	)
	api.On(
		"PublishWebSocketEvent",
		"event_occur",
		map[string]interface{}{
			"id":      "rec-ev",
			"title":   "test event recevent",
			"channel": nil,
		},
		&model.WebsocketBroadcast{UserId: "owner_id"},
	)

	api.On("CreatePost", postForSendChannel).Return(nil, nil)
	api.On("GetUser", "user-Id").Return(&model.User{
		Username: "userName",
	}, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:        "qwcw",
		Title:     "test event recevent",
		Attendees: []string{"user-Id"},
	}

	postForSendChannel.SetProps(background.getMessageProps(testEvent, time.Now()))

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(false, sqlQueryTime, true, sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time",
		"type",
		"meeting_link",
		"mention",
	},
	).AddRow(
		"rec-ev", "test event recevent", recurrentEventTimeStart,
		recurrentEventTimeEnd, featureTime,
		featureTime, "owner_id", channelId, "user-Id", true,
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU",
		"team1", "", nil, "call", nil, "",
	)

	expectedQuery.WillReturnRows(eventsRow)

	updateBuilder := sq.Update("calendar_events").
		Set("processed", sqlQueryTime).
		Where(sq.Eq{"id": "rec-ev"}).PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()
	expectedQueryUpdate := dbMock.ExpectQuery(regexp.QuoteMeta(updateSql)).WithArgs(sqlQueryTime, "rec-ev")
	expectedQueryUpdate.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec-ev"))

	background.process(processingTime)
	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)

}

// recurrent event start in processing time
func TestProcessCornerEventWithChannelRecurrent(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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

	pluginT.SetDB(dbx)

	processingTime := time.Now().In(time.UTC)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	featureTime := sqlQueryTime.Add(time.Hour * 2)

	recurrentEventTimeStart := time.Date(
		2023,
		time.February,
		26,
		sqlQueryTime.Hour(),
		sqlQueryTime.Minute(),
		sqlQueryTime.Second(),
		sqlQueryTime.Nanosecond(),
		time.UTC,
	)
	recurrentEventTimeEnd := time.Date(
		2023,
		time.February,
		26,
		featureTime.Hour(),
		featureTime.Minute(),
		featureTime.Second(),
		featureTime.Nanosecond(),
		time.UTC,
	)

	postForSendChannel := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
	}

	api.On("CreatePost", postForSendChannel).Return(nil, nil)
	api.On("GetUser", "user-Id").Return(&model.User{
		Username: "userName",
	}, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "rec-ev",
		"title":   "test event recurrent",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "user-Id",
	}).Return(nil, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "rec-ev",
		"title":   "test event recurrent",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "owner_id",
	}).Return(nil, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:        "qwcw",
		Title:     "test event recurrent",
		Attendees: []string{"user-Id"},
	}

	postForSendChannel.SetProps(background.getMessageProps(testEvent, time.Now()))

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(false, sqlQueryTime, true, sqlQueryTime, sqlQueryTime)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time",
		"type",
		"meeting_link",
		"mention",
	},
	).AddRow(
		"rec-ev", "test event recurrent", recurrentEventTimeStart,
		recurrentEventTimeEnd, recurrentEventTimeStart,
		recurrentEventTimeStart, "owner_id", channelId, "user-Id", true,
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR,SA,SU",
		"team1", "", nil, "call", nil, "",
	)

	expectedQuery.WillReturnRows(eventsRow)

	updateBuilder := sq.Update("calendar_events").
		Set("processed", sqlQueryTime).
		Where(sq.Eq{"id": "rec-ev"}).PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(updateSql)).WithArgs(sqlQueryTime, "rec-ev").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec-ev"))

	background.process(processingTime)

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)

}

// process event without channel
func TestProcessEventWithoutChannel(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"

	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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
	pluginT.SetDB(dbx)

	processingTime := time.Now().In(time.UTC)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	postForSendGroup := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
	}

	foundChannel := &model.Channel{
		Id: channelId,
	}

	api.On("GetGroupChannel", []string{"user-id", "owner-id", "bot-id"}).Return(foundChannel, nil)
	api.On("GetUser", "user-id").Return(&model.User{
		Username: "userName",
	}, nil)

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "qwert-2",
		"title":   "tests event without channel",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "user-id",
	}).Return(nil, nil)
	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      "qwert-2",
		"title":   "tests event without channel",
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: "owner-id",
	}).Return(nil, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:        "",
		Title:     "tests event without channel",
		Attendees: []string{"user-id"},
	}

	postForSendGroup.SetProps(background.getMessageProps(testEvent, time.Now()))

	api.On("CreatePost", postForSendGroup).Return(nil, nil)

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(regexp.QuoteMeta(querySql)).
		WithArgs(
			false,
			sqlQueryTime,
			true,
			sqlQueryTime,
			sqlQueryTime,
		)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time",
		"type",
		"meeting_link",
		"mention",
	},
	).AddRow("qwert-2", "tests event without channel", sqlQueryTime, sqlQueryTime, sqlQueryTime,
		sqlQueryTime, "owner-id", nil, "user-id", false, "", "team1", "", nil, "call", nil, "")

	expectedQuery.WillReturnRows(eventsRow)

	updateBuilder := sq.Update("calendar_events").
		Set("processed", sqlQueryTime).
		Where(sq.Eq{"id": "qwert-2"}).PlaceholderFormat(sq.Dollar)
	updateSql, _, _ := updateBuilder.ToSql()
	dbMock.ExpectQuery(regexp.QuoteMeta(updateSql)).
		WithArgs(sqlQueryTime, "qwert-2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("qwert-2"))

	background.process(processingTime)

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)

}

// The process event isn't in the rule for the day
func TestProcessEventWithChannelRecurrentNotDay(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: botId,
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
	pluginT.SetDB(dbx)

	processingTime := time.Date(
		2023,
		04,
		12,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	sqlQueryTime := time.Date(
		processingTime.Year(),
		processingTime.Month(),
		processingTime.Day(),
		processingTime.Hour(),
		processingTime.Minute(),
		0,
		0,
		time.UTC,
	)

	featureTime := sqlQueryTime.Add(time.Hour * 2)

	recurrentEventTimeStart := time.Date(
		2023,
		time.March,
		11,
		sqlQueryTime.Hour(),
		sqlQueryTime.Minute(),
		sqlQueryTime.Second(),
		sqlQueryTime.Nanosecond(),
		time.UTC,
	)
	recurrentEventTimeEnd := time.Date(
		2023,
		time.March,
		11,
		featureTime.Hour(),
		featureTime.Minute(),
		featureTime.Second(),
		featureTime.Nanosecond(),
		time.UTC,
	)

	postForSendChannel := &model.Post{
		UserId:    botId,
		ChannelId: channelId,
	}

	api.On("GetUser", "user-Id").Return(&model.User{
		Username: "userName",
	}, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	testEvent := &Event{
		Id:        "qwcw",
		Title:     "test event recurrent",
		Attendees: []string{"user-Id"},
	}

	postForSendChannel.SetProps(background.getMessageProps(testEvent, time.Now()))

	recurrentTimeQuery := sq.And{
		sq.Eq{"ce.recurrent": true},
		sq.Eq{"ce.alert_time::time": sqlQueryTime},
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.updated",
			"ce.owner",
			"ce.channel",
			"cm.member",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"ce.alert_time",
			"ce.alert",
			"ce.team",
			"ce.type",
			"ce.meeting_link",
			"ce.mention",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.And{
			sq.Eq{"ce.all_day": false},
			sq.Or{
				sq.Eq{"ce.alert_time": sqlQueryTime},
				recurrentTimeQuery,
			},
			sq.Or{
				sq.Eq{"ce.processed": nil},
				sq.NotEq{"ce.processed": sqlQueryTime},
			},
		}).
		PlaceholderFormat(sq.Dollar)

	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(
		regexp.QuoteMeta(querySql)).
		WithArgs(
			false,
			sqlQueryTime,
			true,
			sqlQueryTime,
			sqlQueryTime,
		)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"title",
		"dt_start",
		"dt_end",
		"created",
		"updated",
		"owner",
		"channel",
		"member",
		"recurrent",
		"recurrence",
		"team",
		"alert",
		"alert_time", "type", "meeting_link", "mention"},
	).AddRow(
		"rec-ev", "test event recurrent",
		recurrentEventTimeStart, recurrentEventTimeEnd, recurrentEventTimeStart,
		recurrentEventTimeStart, "owner_id", channelId, "user-Id", true, "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=SU,MO",
		"team1", "", nil, "call", nil, "",
	)

	expectedQuery.WillReturnRows(eventsRow)

	background.process(processingTime)

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	api.AssertExpectations(t)

}

func TestWSSendNotification(t *testing.T) {
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

	api := plugintest.API{}

	pluginT := &Plugin{
		BotId: "bot-id",
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	api.On("PublishWebSocketEvent", "event_occur", map[string]interface{}{
		"id":      testEvent.Id,
		"title":   testEvent.Title,
		"channel": nil,
	}, &model.WebsocketBroadcast{
		UserId: testEvent.Owner,
	}).Return(nil, nil)

	background := &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: pluginT,
	}

	background.sendWsNotification(testEvent, time.Now())

	api.AssertExpectations(t)
}

// "At start time" is a plain ping, not a "heads-up N before" message, so it
// must not carry an alarm-clock title like the other alerts do.
// Events created before the mention setting existed carry MentionNone, and
// must stay silent rather than suddenly pinging whole channels.
func TestEventMentionAsText(t *testing.T) {
	cases := map[EventMention]string{
		MentionNone:    "",
		MentionHere:    "@here",
		MentionChannel: "@channel",
		MentionAll:     "@all",
	}

	for mention, want := range cases {
		if got := mention.AsText(); got != want {
			t.Errorf("mention %q: expected %q, got %q", string(mention), want, got)
		}
	}
}

func TestGetMessageFromEventAtStartTimeHasNoAlertTitle(t *testing.T) {
	now := time.Now()
	testEvent := &Event{
		Title:     "standup",
		Alert:     EventAlertAtStartTime,
		AlertTime: &now,
	}

	background := &Background{plugin: &Plugin{}}
	message := background.getMessageFromEvent(testEvent, now)

	if strings.Contains(message, ":alarm_clock:") {
		t.Errorf("expected a plain start-time ping, got an alarm-clock message: %q", message)
	}
	if !strings.Contains(message, "standup") {
		t.Errorf("expected the message to contain the event title, got: %q", message)
	}
}

// every other alert keeps the alarm-clock heads-up with its title
func TestGetMessageFromEventWithAlertHasAlarmClockTitle(t *testing.T) {
	now := time.Now()
	testEvent := &Event{
		Title:     "standup",
		Alert:     EventAlert15MinutesBefore,
		AlertTime: &now,
	}

	background := &Background{plugin: &Plugin{}}
	message := background.getMessageFromEvent(testEvent, now)

	if !strings.Contains(message, ":alarm_clock:") {
		t.Errorf("expected an alarm-clock message, got: %q", message)
	}
	if !strings.Contains(message, "15 minutes before") {
		t.Errorf("expected the alert title in the message, got: %q", message)
	}
}

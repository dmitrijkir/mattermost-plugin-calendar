package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
	"time"
)

func TestGetUTCEvents(t *testing.T) {
	api := plugintest.API{}

	api.On("GetTeamsForUser", "test-user").Return([]*model.Team{}, nil)

	session := &model.Session{
		UserId: "test-user",
	}
	userLocation, _ := time.LoadLocation("Europe/Berlin")

	// DB mocks
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	sqlRequestTimeStart := time.Date(2023, time.February, 26, 23, 0, 0, 0, time.UTC)
	sqlRequestTimeEnd := time.Date(2023, time.March, 05, 23, 0, 0, 0, time.UTC)

	conditions := sq.And{
		sq.Or{
			sq.Eq{"cm.member": session.UserId},
			sq.Eq{"ce.owner": session.UserId},
			sq.NotEq{"ce.visibility": string(VisibilityPrivate)},
		},
		sq.Or{
			sq.And{
				sq.GtOrEq{"ce.dt_start": sqlRequestTimeStart},
				sq.LtOrEq{"ce.dt_start": sqlRequestTimeEnd},
			},
			sq.Eq{"ce.recurrent": true},
		},
	}

	// Create a new select builder
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

	querySql, _, err := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(
		regexp.QuoteMeta(querySql),
	).WithArgs(session.UserId, session.UserId, "private", sqlRequestTimeStart, sqlRequestTimeEnd, true)

	sqlEventsRow := sqlmock.NewRows([]string{
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
	})

	//	add events to sqlEventsRow
	// common event
	sqlEventsRow.AddRow(
		"event-1",
		"test event 1",
		"",
		sqlRequestTimeStart,
		sqlRequestTimeStart.Add(time.Minute*30),
		sqlRequestTimeEnd,
		session.UserId,
		"channel-1",
		false,
		"",
		"#000000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)
	// recurrent event, every monday, tuesday, wednesday
	sqlEventsRow.AddRow(
		"event-2",
		"test event 2",
		"",
		sqlRequestTimeStart,
		sqlRequestTimeStart.Add(time.Minute*30),
		sqlRequestTimeEnd,
		session.UserId,
		"channel-1",
		true,
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE",
		"#000000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)

	// 2 events with multiple members, should be mapped to 1 event
	sqlEventsRow.AddRow(
		"event-3",
		"test event 3",
		"",
		sqlRequestTimeStart,
		sqlRequestTimeStart.Add(time.Minute*30),
		sqlRequestTimeEnd,
		session.UserId,
		"channel-1",
		false,
		"",
		"#000000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)
	sqlEventsRow.AddRow(
		"event-3",
		"test event 3",
		"",
		sqlRequestTimeStart,
		sqlRequestTimeStart.Add(time.Minute*30),
		sqlRequestTimeEnd,
		"another-user",
		"channel-1",
		false,
		"",
		"#000000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)

	// recurrent event, every second monday, event must start 2 week earlier
	sqlEventsRow.AddRow(
		"event-4",
		"test event 4",
		"",
		sqlRequestTimeStart.Add(-time.Hour*24*14),
		sqlRequestTimeStart.Add(-time.Hour*24*14).Add(time.Minute*30),
		sqlRequestTimeEnd,
		session.UserId,
		"channel-1",
		true,
		"RRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=MO",
		"#000000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)

	// recurrent event, corner case, start 00:00, and repeat every current week day
	sqlEventsRow.AddRow(
		"event-5",
		"test event 5",
		"",
		time.Date(2023, time.February, 27, 23, 0, 0, 0, time.UTC),
		time.Date(2023, time.February, 27, 24, 0, 0, 0, time.UTC),
		sqlRequestTimeEnd,
		session.UserId,
		"channel-1",
		true,
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TU",
		"#00000",
		"team1",
		VisibilityPrivate,
		"",
		nil,
	)
	//

	expectedQuery.WillReturnRows(sqlEventsRow)

	// Create a new select builder
	queryBuilderUsersInChannel := sq.Select().
		Columns("channelid").
		From("channelmembers").
		Where(sq.Eq{"userid": session.UserId}).
		PlaceholderFormat(sq.Dollar)

	querySqlUsersInChannel, _, _ := queryBuilderUsersInChannel.ToSql()
	expectedQueryUsersInChannel := dbMock.ExpectQuery(
		regexp.QuoteMeta(querySqlUsersInChannel),
	).WithArgs(session.UserId)
	expectedQueryUsersInChannel.WillReturnRows(sqlmock.NewRows([]string{"channelid"}).AddRow("channel-1"))

	calPlugin := Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
		DB: dbx,
	}

	events, eventsErr := calPlugin.GetUserEventsUTC(session.UserId, userLocation, sqlRequestTimeStart, sqlRequestTimeEnd)

	if eventsErr != nil {
		t.Errorf("Error getting events: %s", eventsErr)
	}
	api.AssertExpectations(t)

	assertChecker := assert.New(t)

	assertChecker.Equal(7, len(events))

	// check event-1
	assertChecker.Equal("event-1", events[0].Id)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 0, 0, 0, userLocation),
		events[0].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 30, 0, 0, userLocation),
		events[0].End,
	)

	// check event-2
	assertChecker.Equal("event-2", events[1].Id)

	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 0, 0, 0, userLocation),
		events[1].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 30, 0, 0, userLocation),
		events[1].End,
	)

	assertChecker.Equal(
		time.Date(2023, time.February, 28, 00, 0, 0, 0, userLocation),
		events[2].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 28, 00, 30, 0, 0, userLocation),
		events[2].End,
	)

	assertChecker.Equal(
		time.Date(2023, time.February, 29, 00, 0, 0, 0, userLocation),
		events[3].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 29, 00, 30, 0, 0, userLocation),
		events[3].End,
	)

	// check event-3
	assertChecker.Equal("event-3", events[4].Id)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 0, 0, 0, userLocation),
		events[4].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 30, 0, 0, userLocation),
		events[4].End,
	)

	//	check event-4
	assertChecker.Equal("event-4", events[5].Id)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 0, 0, 0, userLocation),
		events[5].Start,
	)
	assertChecker.Equal(
		time.Date(2023, time.February, 27, 00, 30, 0, 0, userLocation),
		events[5].End,
	)
}

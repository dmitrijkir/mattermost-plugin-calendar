package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"regexp"
	"testing"
)

func TestMigrateLegacy(t *testing.T) {
	db, dbMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dbx := sqlx.NewDb(db, "sqlmock")

	api := plugintest.API{}
	pluginT := &Plugin{
		BotId: "calendar-id",
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	migrator := Migrator{
		plugin: pluginT,
		DB:     dbx,
	}

	expectedQuery := dbMock.ExpectQuery(
		regexp.QuoteMeta(
			"SELECT id, recurrence FROM calendar_events WHERE recurrence LIKE '[%' and recurrent = true",
		),
	)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"recurrence",
	}).AddRow("qwer", "[1]").AddRow("qazx", "[]")

	expectedQuery.WillReturnRows(eventsRow)

	dbMock.ExpectExec(
		regexp.QuoteMeta(
			`UPDATE PUBLIC.calendar_events SET recurrence = ? WHERE id = ?`,
		),
	).WithArgs(
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
		"qwer",
	).WillReturnResult(
		sqlmock.NewResult(0, 0),
	)

	dbMock.ExpectExec(
		regexp.QuoteMeta(
			`UPDATE calendar_events SET recurrence = '' 
				WHERE recurrence = '[]' or recurrence = 'null'`,
		),
	).WillReturnResult(
		sqlmock.NewResult(0, 0),
	)

	migrator.migrateLegacyRecurrentEvents()

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

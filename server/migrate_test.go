package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
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

	queryBuilder := sq.Select().
		Columns("id", "recurrence").
		From("calendar_events").
		Where(sq.Like{"recurrence": "[%"}).
		Where("recurrent = true").
		PlaceholderFormat(sq.Dollar)
	querySql, _, _ := queryBuilder.ToSql()
	expectedQuery := dbMock.ExpectQuery(
		regexp.QuoteMeta(
			querySql,
		),
	)

	eventsRow := sqlmock.NewRows([]string{
		"id",
		"recurrence",
	}).AddRow("qwer", "[1]").AddRow("qazx", "[]")

	expectedQuery.WillReturnRows(eventsRow)

	updateQueryBuilder := sq.Update("calendar_events").
		Set("recurrence", "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO").
		Where(sq.Eq{"id": "qwer"})

	updateQuerySql, _, _ := updateQueryBuilder.ToSql()
	dbMock.ExpectQuery(
		regexp.QuoteMeta(
			updateQuerySql,
		),
	).WithArgs(
		"RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=MO",
		"qwer",
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("qwer"))

	updateEmptyBuilder := sq.Update("calendar_events").
		Set("recurrence", "").
		Where(sq.Or{sq.Eq{"recurrence": "[]"}, sq.Eq{"recurrence": "null"}})

	updateEmptySql, _, _ := updateEmptyBuilder.ToSql()
	dbMock.ExpectQuery(
		regexp.QuoteMeta(
			updateEmptySql,
		),
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("qazx"))

	migrator.migrateLegacyRecurrentEvents()

	if err := dbMock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

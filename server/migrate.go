package main

import (
	"context"
	"database/sql/driver"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/morph"
	"github.com/mattermost/morph/drivers/postgres"
	"github.com/mattermost/morph/sources"
	"github.com/mattermost/morph/sources/embedded"
	"path/filepath"
	"strings"
)

//go:embed migrations
var assets embed.FS

type RecurrenceItem []int

func (r *RecurrenceItem) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &r)
		return nil
	case string:
		json.Unmarshal([]byte(v), &r)
		return nil
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %T", v))
	}
}

func (r *RecurrenceItem) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type Migrator struct {
	DB     *sqlx.DB
	plugin *Plugin
}

func (m *Migrator) createSource() (sources.Source, error) {
	assetsList, err := assets.ReadDir(filepath.Join("migrations", m.DB.DriverName()))

	if err != nil {
		m.plugin.API.LogError(err.Error())
		return nil, err
	}

	assetNamesForDriver := make([]string, len(assetsList))
	for i, entry := range assetsList {
		assetNamesForDriver[i] = entry.Name()
	}
	src, err := embedded.WithInstance(&embedded.AssetSource{
		Names: assetNamesForDriver,
		AssetFunc: func(name string) ([]byte, error) {
			return assets.ReadFile(filepath.Join("migrations", m.DB.DriverName(), name))
		},
	})

	return src, err
}

func (m *Migrator) migrate() *model.AppError {
	driver, err := postgres.WithInstance(m.DB.DB, &postgres.Config{})

	if err != nil {
		return CantMakeMigration
	}

	src, err := m.createSource()

	if err != nil {
		m.plugin.API.LogError(err.Error())
		return CantMakeMigration
	}

	opts := []morph.EngineOption{
		morph.WithLock("mm-calendar-lock-key"),
		morph.SetMigrationTableName("calendar_db_migrations"),
		morph.SetStatementTimeoutInSeconds(100000),
	}
	engine, err := morph.New(context.Background(), driver, src, opts...)

	if err != nil {
		m.plugin.API.LogError(err.Error())
		return CantMakeMigration
	}

	defer engine.Close()

	err = engine.ApplyAll()
	if err != nil {
		m.plugin.API.LogError(err.Error())
		return CantMakeMigration
	}

	return nil
}
func (m *Migrator) migrateLegacyRecurrentEvents() *model.AppError {
	rows, errSelect := m.DB.Queryx(`
			SELECT id, recurrence FROM calendar_events WHERE recurrence LIKE '[%' and recurrent = true
`)
	if errSelect != nil {
		m.plugin.API.LogError(errSelect.Error())
	}

	type EventFromDb struct {
		Id         string          `json:"id" db:"id"`
		Recurrence *RecurrenceItem `json:"recurrence" db:"recurrence"`
	}

	dayOfWeek := map[int]string{
		0: "MO",
		1: "TU",
		2: "WE",
		3: "TH",
		4: "FR",
		5: "SA",
		6: "SU",
	}

	for rows.Next() {
		var eventDb EventFromDb

		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			m.plugin.API.LogError(errSelect.Error())
			continue
		}

		recurrenceDays := []string{}

		for value := range *eventDb.Recurrence {
			recurrenceDays = append(recurrenceDays, dayOfWeek[value])
		}

		if len(recurrenceDays) < 1 {
			continue
		}
		rrule := "RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=" + strings.Join(recurrenceDays, ",")

		_, errUpdate := m.DB.NamedExec(`UPDATE PUBLIC.calendar_events
                                           SET recurrence = :recurrence
                                           WHERE id = :eventId`, map[string]interface{}{
			"recurrence": rrule,
			"eventId":    eventDb.Id,
		})
		if errUpdate != nil {
			m.plugin.API.LogError(errUpdate.Error())
			continue
		}
	}

	// clear empty rows
	_, errUpdate := m.DB.Exec(`UPDATE calendar_events 
									 SET recurrence = '' 
									 WHERE recurrence = '[]' or recurrence = 'null'`)
	if errUpdate != nil {
		m.plugin.API.LogError(errUpdate.Error())
		return CantMakeMigration
	}

	return nil
}

func newMigrator(db *sqlx.DB, plugin *Plugin) *Migrator {
	return &Migrator{
		DB:     db,
		plugin: plugin,
	}
}

package main

import (
	"context"
	"embed"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/morph"
	"github.com/mattermost/morph/drivers/postgres"
	"github.com/mattermost/morph/sources"
	"github.com/mattermost/morph/sources/embedded"
	"path/filepath"
)

//go:embed migrations
var assets embed.FS

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

func newMigrator(db *sqlx.DB, plugin *Plugin) *Migrator {
	return &Migrator{
		DB:     db,
		plugin: plugin,
	}
}

package main

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"net/http"
	"sync"
)

const (
	BotUsername    = "calendar"
	BotDisplayName = "Calendar"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	router *mux.Router

	DB    *sqlx.DB
	BotId string
}

func (p *Plugin) SetDB(db *sqlx.DB) {
	p.DB = db
}

func (p *Plugin) GetDBPlaceholderFormat() sq.PlaceholderFormat {
	if p.DB == nil {
		return sq.Dollar
	}

	switch p.DB.DriverName() {
	case POSTGRES:
		return sq.Dollar
	case MYSQL:
		return sq.Question
	default:
		return sq.Dollar
	}

}
func (p *Plugin) SetBotId(botId string) {
	p.BotId = botId
}

func (p *Plugin) FromContext(ctx context.Context) *plugin.Context {
	return ctx.Value("pluginRequest").(*plugin.Context)
}

func (p *Plugin) OnActivate() error {

	config := p.API.GetUnsanitizedConfig()

	db := initDb(*config.SqlSettings.DriverName, *config.SqlSettings.DataSource)
	p.SetDB(db)

	migrator := newMigrator(db, p)
	if errMigrate := migrator.migrate(); errMigrate != nil {
		return errMigrate
	}

	if errMigrate := migrator.migrateLegacyRecurrentEvents(); errMigrate != nil {
		return errMigrate
	}

	command, err := p.createCalCommand()
	if err != nil {
		return err
	}

	if err = p.API.RegisterCommand(command); err != nil {
		return err
	}

	GetBotsResp, GetBotError := p.API.GetBots(&model.BotGetOptions{
		Page:           0,
		PerPage:        1000,
		OwnerId:        "",
		IncludeDeleted: false,
	})

	if GetBotError != nil {
		p.API.LogError(GetBotError.Error())
		return &model.AppError{
			Message:       "Can't get bot",
			DetailedError: GetBotError.Error(),
		}
	}

	botId := ""

	for _, bot := range GetBotsResp {
		if bot.Username == BotUsername {
			botId = bot.UserId
		}
	}

	if botId == "" {
		createdBot, createBotError := p.API.CreateBot(&model.Bot{
			Username:    BotUsername,
			DisplayName: BotDisplayName,
		})
		if createBotError != nil {
			p.API.LogError(createBotError.Error())
			return &model.AppError{
				Message:       "Can't create bot",
				DetailedError: createBotError.Error(),
			}
		}

		botId = createdBot.UserId

	}

	p.SetBotId(botId)
	p.router = p.InitAPI()

	go NewBackgroundJob(p).Start()
	return nil
}

func (p *Plugin) OnDeactivate() error {
	GetBackgroundJob().Done <- true

	return nil
}

// handles HTTP requests.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "pluginRequest", c)
	r = r.Clone(ctx)
	p.router.ServeHTTP(w, r)
}

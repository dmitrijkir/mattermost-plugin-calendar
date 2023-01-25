package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/v6/plugin"
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
}

func (p *Plugin) OnActivate() error {

	config := p.API.GetUnsanitizedConfig()
	initDb(*config.SqlSettings.DriverName, *config.SqlSettings.DataSource)

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

	go NewBackgroundJob(p, botId).Start()
	return nil
}

func (p *Plugin) OnDeactivate() error {
    GetBackgroundJob().Done <- true

    return nil
}

// handles HTTP requests.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		switch r.URL.Path {
		case "/events":
			p.CreateEvent(c, w, r)
		}
	case "GET":
		switch r.URL.Path {
		case "/events":
			p.GetEvents(c, w, r)
		case "/event":
			p.GetEvent(c, w, r)

		}
	case "DELETE":
		switch r.URL.Path {
		case "/event":
			p.RemoveEvent(c, w, r)
		}
	default:
		fmt.Fprint(w, "ping")
	}
}

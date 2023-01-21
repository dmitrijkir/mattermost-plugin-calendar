package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-server/v6/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
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
	return nil
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	db := GetDb()
	err := db.Ping()
	if err != nil {
		fmt.Fprint(w, "error")
		fmt.Fprint(w, err.Error())
	}

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

// See https://developers.mattermost.com/extend/plugins/server/reference/

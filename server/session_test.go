package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/stretchr/testify/assert"
)

// The webapp initializes plugins on the login page, before the user has a session,
// so session-protected endpoints receive requests with an empty SessionId.
// They must get 401 without GetSession calls and without error log entries.
func TestSessionEndpoints_Unauthenticated(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/settings"},
		{http.MethodPut, "/settings"},
		{http.MethodGet, "/events"},
		{http.MethodPost, "/events"},
		{http.MethodPut, "/events"},
		{http.MethodGet, "/events/some-event"},
		{http.MethodDelete, "/events/some-event"},
		{http.MethodGet, "/schedule"},
		{http.MethodGet, "/ical/token"},
		{http.MethodPost, "/ical/token"},
		{http.MethodDelete, "/ical/token"},
	}

	for _, e := range endpoints {
		t.Run(e.method+" "+e.path, func(t *testing.T) {
			api := &plugintest.API{}
			api.On("LogDebug", "Plugin HTTP request", "method", e.method, "path", e.path, "user-agent", "").Return()

			calPlugin := Plugin{MattermostPlugin: plugin.MattermostPlugin{API: api}}
			calPlugin.router = calPlugin.InitAPI()

			w := httptest.NewRecorder()
			calPlugin.ServeHTTP(&plugin.Context{}, w, httptest.NewRequest(e.method, e.path, nil))

			assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
			// plugintest.API panics on any unexpected call (GetSession, LogError, LogWarn)
			api.AssertExpectations(t)
		})
	}
}

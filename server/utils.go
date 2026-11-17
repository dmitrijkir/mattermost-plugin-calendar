package main

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
)

func contains[K comparable](arr []K, item K) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}

	return false
}

func errorResponse(w http.ResponseWriter, err *model.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	w.Write(model.ToJSON(err))
	return
}

func apiResponse(w http.ResponseWriter, res interface{}) {
	w.Header().Set("Content-Type", "application/json")

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"data": res,
	})

	if _, errWrite := w.Write(jsonBytes); errWrite != nil {
		errorResponse(w, SomethingWentWrong)
		return
	}
}

// sessionFromRequest returns the session of the user who made the request.
// On failure it writes 401 to w and returns ok == false.
//
// Requests without a session are expected and are not an error: the Mattermost
// webapp initializes plugins before the user logs in (on the login page), so the
// plugin's first requests arrive unauthenticated. They are answered with 401
// without an error log entry, otherwise every visit to the login page produces
// a "can't get session" error in the server log.
func (p *Plugin) sessionFromRequest(w http.ResponseWriter, r *http.Request) (*model.Session, bool) {
	pluginContext := p.FromContext(r.Context())
	if pluginContext == nil || pluginContext.SessionId == "" {
		errorResponse(w, NotAuthorizedError)
		return nil, false
	}

	session, err := p.API.GetSession(pluginContext.SessionId)
	if err != nil {
		p.API.LogWarn("can't get session", "path", r.URL.Path, "error", err.Error())
		errorResponse(w, NotAuthorizedError)
		return nil, false
	}

	return session, true
}

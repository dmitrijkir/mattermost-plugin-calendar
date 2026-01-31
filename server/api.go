package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (p *Plugin) InitAPI() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/events", p.GetEvents).Methods("GET")
	r.HandleFunc("/events/{eventId}", p.GetEvent).Methods("GET")
	r.HandleFunc("/events/{eventId}", p.RemoveEvent).Methods("DELETE")
	r.HandleFunc("/events", p.CreateEvent).Methods("POST")
	r.HandleFunc("/events", p.UpdateEvent).Methods("PUT")

	r.HandleFunc("/settings", p.GetSettings).Methods("GET")
	r.HandleFunc("/settings", p.UpdateSettings).Methods("PUT")

	r.HandleFunc("/schedule", p.GetSchedule).Methods("GET")

	// iCal token management
	r.HandleFunc("/ical/token", p.GetICalToken).Methods("GET")
	r.HandleFunc("/ical/token", p.GenerateICalToken).Methods("POST")
	r.HandleFunc("/ical/token", p.RevokeICalToken).Methods("DELETE")
	// iCal feed endpoint (token is 64-char hex string)
	r.HandleFunc("/ical/feed/{token}", p.ServeICalFeed).Methods("GET")

	// CalDAV endpoints (use PathPrefix for all CalDAV requests)
	r.PathPrefix("/caldav/{token}/").HandlerFunc(p.ServeCalDAV)

	// 404 handler
	r.Handle("{anything:.*}", http.NotFoundHandler())
	return r
}

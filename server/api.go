package main

import (
	"github.com/gorilla/mux"
	"net/http"
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

	// 404 handler
	r.Handle("{anything:.*}", http.NotFoundHandler())
	return r
}

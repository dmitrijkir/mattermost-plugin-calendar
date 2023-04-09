package main

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v6/model"
	"net/http"
	"time"
)

func contains[K comparable](arr []K, item K) bool {
	for _, v := range arr {
		if v == item {
			return true
		}
	}

	return false
}

func containsTime(arr []time.Time, item time.Time) bool {
	for _, v := range arr {
		if v.Equal(item) {
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

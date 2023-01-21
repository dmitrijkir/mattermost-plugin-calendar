package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
    "github.com/mattermost/mattermost-server/v6/model"
    "github.com/mattermost/mattermost-server/v6/plugin"
	"net/http"
	"time"
)

type Event struct {
	Id        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Start     time.Time `json:"start" db:"start"`
	End       time.Time `json:"end" db:"end"`
	Attendees []string  `json:"attendees"`
	Created   time.Time `json:"created" db:"created"`
	Owner     string    `json:"owner" db:"owner"`
}

func (p *Plugin) GetUserLocation (user *model.User) (*time.Location) {
    userTimeZone := ""

    if user.Timezone["useAutomaticTimezone"] == "true" {
        userTimeZone = user.Timezone["automaticTimezone"]
    } else {
        userTimeZone = user.Timezone["manualTimezone"]
    }

    userLoc, loadError := time.LoadLocation(userTimeZone)

    if loadError != nil {
        userLoc, _ = time.LoadLocation("")
    }

    return userLoc
}

func (p *Plugin) GetEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		fmt.Fprint(w, err.Error())
        return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
    }

	query := r.URL.Query()

	eventId := query.Get("eventId")

	rows, errSelect := GetDb().Queryx(`SELECT ce.id,
                                              ce.title,
                                              ce."start",
                                              ce."end",
                                              ce.created,
                                              ce."owner",
                                              cm."user"
                                       FROM   calendar_events ce
                                              LEFT JOIN calendar_members cm
                                                      ON ce.id = cm."event"
                                       WHERE  id = $1 `, eventId)
	if errSelect != nil {
		fmt.Fprint(w, errSelect.Error())
        return
	}

	type EventFromDb struct {
		Event
		User *string `json:"user" db:"user"`
	}

	members := []string{}
	var eventDb EventFromDb

	for rows.Next() {
		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			fmt.Fprint(w, errScan.Error())
            return
		}

        if eventDb.User != nil {
            members = append(members, *eventDb.User)
        }

	}

	event := Event{
		Id:        eventDb.Id,
		Title:     eventDb.Title,
		Start:     eventDb.Start,
		End:       eventDb.End,
		Attendees: members,
		Created:   eventDb.Created,
		Owner:     eventDb.Owner,
	}

    userLoc := p.GetUserLocation(user)
    event.Start = event.Start.In(userLoc)
    event.End =event.End.In(userLoc)

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"data": &event,
	})

	w.Header().Set("Content-Type", "application/json")

	if _, errWrite := w.Write(jsonBytes); err != nil {
		http.Error(w, fmt.Sprintf("Error getting dynamic args: %s", errWrite.Error()), http.StatusInternalServerError)
		return
	}

}

func (p *Plugin) GetEvents(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		fmt.Fprint(w, err.Error())
        return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}

	var events []Event

    rows, errSelect := GetDb().Queryx(`SELECT ce.id,
                                              ce.title,
                                              ce."start",
                                              ce."end",
                                              ce.created,
                                              ce."owner"
                                       FROM   calendar_events ce
                                              FULL JOIN calendar_members cm
                                                     ON ce.id = cm."event"
                                       WHERE  cm."user" = $1
                                       OR ce."owner" = $2`, user.Id, user.Id)

	if errSelect != nil {
		fmt.Fprint(w, errSelect.Error())
		return
	}

    userLoc := p.GetUserLocation(user)

    addedEvent := map[string]bool{}

	for rows.Next() {
		var eventDb Event

		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			fmt.Fprint(w, errScan.Error())
			return
		}

        eventDb.Start = eventDb.Start.In(userLoc)
        eventDb.End =eventDb.End.In(userLoc)

        if !addedEvent[eventDb.Id] {
            events = append(events, eventDb)
            addedEvent[eventDb.Id] = true
        }
	}

	jsonBytes, _ := json.Marshal(map[string]interface{}{
		"data": &events,
	})

	w.Header().Set("Content-Type", "application/json")

	if _, errWrite := w.Write(jsonBytes); err != nil {
		http.Error(w, fmt.Sprintf("Error getting dynamic args: %s", errWrite.Error()), http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) CreateEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		fmt.Fprint(w, err.Error())
        return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		fmt.Fprint(w, err.Error())
        return
	}

	var event Event

	errDecode := json.NewDecoder(r.Body).Decode(&event)

	if errDecode != nil {
		fmt.Fprint(w, errDecode.Error())
		return
	}
	event.Id = uuid.New().String()

	event.Created = time.Now().UTC()
	event.Owner = user.Id

    loc := p.GetUserLocation(user)
    utcLoc, _ := time.LoadLocation("UTC")
    startDateInLocalTimeZone := time.Date(
        event.Start.Year(),
        event.Start.Month(),
        event.Start.Day(),
        event.Start.Hour(),
        event.Start.Minute(),
        event.Start.Second(),
        event.Start.Nanosecond(),
        loc,
        )
    endDateInLocalTimeZone := time.Date(
        event.End.Year(),
        event.End.Month(),
        event.End.Day(),
        event.End.Hour(),
        event.End.Minute(),
        event.End.Second(),
        event.End.Nanosecond(),
        loc,
        )
    event.Start = startDateInLocalTimeZone.In(utcLoc)
    event.End = endDateInLocalTimeZone.In(utcLoc)

    _, errInser := GetDb().NamedExec(`INSERT INTO PUBLIC.calendar_events
                                                  (id,
                                                   title,
                                                   "start",
                                                   "end",
                                                   created,
                                                   owner)
                                      VALUES      (:id,
                                                   :title,
                                                   :start,
                                                   :end,
                                                   :created,
                                                   :owner) `, &event)

	if errInser != nil {
		fmt.Fprint(w, errInser.Error())
		return
	}
	if event.Attendees != nil {
		for _, userId := range event.Attendees {
			_, errInser = GetDb().NamedExec(`INSERT INTO public.calendar_members ("event", "user") VALUES (:event, :user)`, map[string]interface{}{
				"event": event.Id,
				"user":  userId,
			})
		}
	}

	if errInser != nil {
		fmt.Fprint(w, errInser.Error())
		return
	}

	jsonBytes, _ := json.Marshal(event)
	w.Header().Set("Content-Type", "application/json")

	if _, errWrite := w.Write(jsonBytes); err != nil {
		http.Error(w, fmt.Sprintf("Error getting dynamic args: %s", errWrite.Error()), http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) RemoveEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
    _, err := p.API.GetSession(c.SessionId)

    if err != nil {
        fmt.Fprint(w, err.Error())
        return
    }

    query := r.URL.Query()

    eventId := query.Get("eventId")

    _, dbErr := GetDb().Exec("DELETE FROM calendar_events WHERE id=$1", eventId)

    if dbErr != nil {
        fmt.Fprint(w, dbErr.Error())
        return
    }

    jsonBytes, _ := json.Marshal(map[string]interface{}{
        "data": map[string]interface{}{
            "success": true,
        },
    })
    w.Header().Set("Content-Type", "application/json")

    if _, errWrite := w.Write(jsonBytes); err != nil {
        http.Error(w, fmt.Sprintf("Error getting dynamic args: %s", errWrite.Error()), http.StatusInternalServerError)
        return
    }
}
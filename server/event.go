package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const (
	EventDateTimeLayout = "2006-01-02T15:04:05"
)

type RecurrenceItem []int

func (r *RecurrenceItem) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &r)
		return nil
	case string:
		json.Unmarshal([]byte(v), &r)
		return nil
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %T", v))
	}
}
func (r *RecurrenceItem) Value() (driver.Value, error) {
	return json.Marshal(r)
}

type Event struct {
	Id         string          `json:"id" db:"id"`
	Title      string          `json:"title" db:"title"`
	Start      time.Time       `json:"start" db:"start"`
	End        time.Time       `json:"end" db:"end"`
	Attendees  []string        `json:"attendees"`
	Created    time.Time       `json:"created" db:"created"`
	Owner      string          `json:"owner" db:"owner"`
	Channel    *string         `json:"channel" db:"channel"`
	Processed  *time.Time      `json:"-" db:"processed"`
	Recurrent  bool            `json:"-" db:"recurrent"`
	Recurrence *RecurrenceItem `json:"recurrence" db:"recurrence"`
}

func (p *Plugin) GetUserLocation(user *model.User) *time.Location {
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
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		p.API.LogError("can't get user")
		errorResponse(w, UserNotFound)
		return
	}

	query := r.URL.Query()

	eventId := query.Get("eventId")

	if eventId == "" {
		errorResponse(w, InvalidRequestParams)
		return
	}

	rows, errSelect := GetDb().Queryx(`
									   SELECT ce.id,
                                              ce.title,
                                              ce."start",
                                              ce."end",
                                              ce.created,
                                              ce."owner",
                                              ce."channel",
                                              ce.recurrence,
                                              cm."user"
                                       FROM   calendar_events ce
                                              LEFT JOIN calendar_members cm
                                                     ON ce.id = cm."event"
                                       WHERE  id = $1 `, eventId)
	if errSelect != nil {
		p.API.LogError("Selecting data error")
		errorResponse(w, EventNotFound)
		return
	}

	type EventFromDb struct {
		Event
		User *string `json:"user" db:"user"`
	}

	var members []string
	var eventDb EventFromDb

	for rows.Next() {
		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			p.API.LogError("Can't scan row to struct EventFromDb")
			return
		}

		if eventDb.User != nil {
			members = append(members, *eventDb.User)
		}

	}

	if eventDb.Id == "" {
		errorResponse(w, EventNotFound)
		return
	}

	event := Event{
		Id:         eventDb.Id,
		Title:      eventDb.Title,
		Start:      eventDb.Start,
		End:        eventDb.End,
		Attendees:  members,
		Created:    eventDb.Created,
		Owner:      eventDb.Owner,
		Channel:    eventDb.Channel,
		Recurrence: eventDb.Recurrence,
	}

	userLoc := p.GetUserLocation(user)

	event.Start = event.Start.In(userLoc)
	event.End = event.End.In(userLoc)

	apiResponse(w, &event)
	return

}

func (p *Plugin) GetEvents(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		p.API.LogError("can't get user")
		errorResponse(w, UserNotFound)
		return
	}

	query := r.URL.Query()

	start := query.Get("start")
	end := query.Get("end")

	if start == "" || end == "" {
		errorResponse(w, InvalidRequestParams)
		return
	}

	userLoc := p.GetUserLocation(user)
	utcLoc, _ := time.LoadLocation("UTC")

	startEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, start, userLoc)
	EndEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, end, userLoc)

	var events []Event

	rows, errSelect := GetDb().Queryx(`
									   SELECT ce.id,
											  ce.title,
											  ce."start",
											  ce."end",
											  ce.created,
											  ce."owner",
											  ce."channel",
											  ce.recurrent,
											  ce.recurrence
									   FROM calendar_events ce
										    FULL JOIN calendar_members cm 
										           ON ce.id = cm."event"
									   WHERE (cm."user" = $1 OR ce."owner" = $2)
											AND (
											     (ce."start" >= $3 AND ce."start" <= $4) 
											         or ce.recurrent = true
											    )
                                       `, user.Id, user.Id, startEventLocal.In(utcLoc), EndEventLocal.In(utcLoc))

	if errSelect != nil {
		p.API.LogError(errSelect.Error())
		apiResponse(w, &events)
		return
	}

	addedEvent := map[string]bool{}
	recurrenEvents := map[int][]Event{}

	for rows.Next() {

		var eventDb Event

		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			p.API.LogError("Can't scan row to struct")
			continue
		}

		eventDb.Start = eventDb.Start.In(userLoc)
		eventDb.End = eventDb.End.In(userLoc)

		if eventDb.Recurrent {
			for _, recurrentDay := range *eventDb.Recurrence {
				recurrenEvents[recurrentDay] = append(recurrenEvents[recurrentDay], eventDb)
			}
			continue
		}

		if !addedEvent[eventDb.Id] && !eventDb.Recurrent {
			events = append(events, eventDb)
			addedEvent[eventDb.Id] = true
		}
	}

	currientDate := startEventLocal
	for currientDate.Before(EndEventLocal) {
		for _, ev := range recurrenEvents[int(currientDate.Weekday())] {
			eventTime := ev.End.Sub(ev.Start)
			ev.Start = time.Date(
				currientDate.Year(),
				currientDate.Month(),
				currientDate.Day(),
				ev.Start.Hour(),
				ev.Start.Minute(),
				ev.Start.Second(),
				ev.Start.Nanosecond(),
				ev.Start.Location(),
			)

			ev.End = ev.Start.Add(eventTime)

			events = append(events, ev)
		}
		currientDate = currientDate.Add(time.Hour * 24)
	}

	apiResponse(w, &events)
	return
}

func (p *Plugin) CreateEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		p.API.LogError(err.Error())
		errorResponse(w, NotAuthorizedError)
		return
	}

	user, err := p.API.GetUser(session.UserId)

	if err != nil {
		p.API.LogError(err.Error())
		errorResponse(w, UserNotFound)
		return
	}

	var event Event

	errDecode := json.NewDecoder(r.Body).Decode(&event)

	if errDecode != nil {
		p.API.LogError(errDecode.Error())
		errorResponse(w, InvalidRequestParams)
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

	if event.Recurrence != nil && len(*event.Recurrence) > 0 {
		event.Recurrent = true
	} else {
		event.Recurrent = false
	}

	_, errInsert := GetDb().NamedExec(`INSERT INTO PUBLIC.calendar_events
                                                  (id,
                                                   title,
                                                   "start",
                                                   "end",
                                                   created,
                                                   owner,
                                                   channel,
                                                   recurrent,
                                                   recurrence)
                                      VALUES      (:id,
                                                   :title,
                                                   :start,
                                                   :end,
                                                   :created,
                                                   :owner,
                                                   :channel,
                                                   :recurrent,
                                                   :recurrence) `, &event)

	if errInsert != nil {
		p.API.LogError(errInsert.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	if event.Attendees != nil {
		var insertParams []map[string]interface{}
		for _, userId := range event.Attendees {
			insertParams = append(insertParams, map[string]interface{}{
				"event": event.Id,
				"user":  userId,
			})
		}

		_, errInsert = GetDb().NamedExec(`INSERT INTO public.calendar_members 
															   ("event", 
															    "user") 
												   VALUES (:event,
												           :user)`, insertParams)
	}

	if errInsert != nil {
		p.API.LogError(errInsert.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	apiResponse(w, &event)
	return
}

func (p *Plugin) RemoveEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {

	_, err := p.API.GetSession(c.SessionId)

	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	query := r.URL.Query()

	eventId := query.Get("eventId")

	if eventId == "" {
		errorResponse(w, InvalidRequestParams)
		return
	}

	_, dbErr := GetDb().Exec("DELETE FROM calendar_events WHERE id=$1", eventId)

	if dbErr != nil {
		p.API.LogError("can't remove event from db")
		errorResponse(w, CantRemoveEvent)
		return
	}

	apiResponse(w, map[string]interface{}{
		"success": true,
	})
	return

}

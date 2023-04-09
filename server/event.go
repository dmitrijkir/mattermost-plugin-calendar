package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/teambition/rrule-go"
	"net/http"
	time "time"
)

func (p *Plugin) GetUserEvents(user *model.User, start, end string) (*[]Event, *model.AppError) {
	userLoc := p.GetUserLocation(user)
	utcLoc, _ := time.LoadLocation("UTC")

	startEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, start, userLoc)
	EndEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, end, userLoc)

	var events []Event

	rows, errSelect := p.DB.Queryx(`
									   SELECT ce.id,
											  ce.title,
											  ce."start",
											  ce."end",
											  ce.created,
											  ce."owner",
											  ce."channel",
											  ce.recurrent,
											  ce.recurrence,
											  ce.color
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
		return nil, SomethingWentWrong
	}

	addedEvent := map[string]bool{}

	for rows.Next() {

		var eventDb Event

		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			p.API.LogError("Can't scan row to struct")
			continue
		}
		if addedEvent[eventDb.Id] {
			continue
		}

		eventDb.Start = eventDb.Start.In(userLoc)
		eventDb.End = eventDb.End.In(userLoc)

		if eventDb.Color == nil {
			color := DefaultColor
			eventDb.Color = &color
		}

		if eventDb.Recurrent {
			eventRule, errRrule := rrule.StrToRRule(eventDb.Recurrence)
			if errRrule != nil {
				p.API.LogError(errRrule.Error())
				continue
			}
			eventRule.DTStart(eventDb.Start)
			eventStartUtc := startEventLocal.In(utcLoc)
			eventDates := eventRule.Between(
				time.Date(
					eventStartUtc.Year(),
					eventStartUtc.Month(),
					eventStartUtc.Day(),
					0,
					0,
					0,
					0,
					eventStartUtc.Location(),
				),
				EndEventLocal.In(utcLoc), false)

			if errRrule != nil {
				p.API.LogError(errRrule.Error())
				continue
			}

			for _, eventDate := range eventDates {
				eventTime := eventDb.End.Sub(eventDb.Start)
				eventDb.Start = time.Date(
					eventDate.Year(),
					eventDate.Month(),
					eventDate.Day(),
					eventDb.Start.Hour(),
					eventDb.Start.Minute(),
					eventDb.Start.Second(),
					eventDb.Start.Nanosecond(),
					eventDb.Start.Location(),
				)
				eventDb.End = eventDb.Start.Add(eventTime)

				events = append(events, eventDb)
			}
		} else {
			events = append(events, eventDb)
		}
		addedEvent[eventDb.Id] = true

	}

	//for _, ev := range events {
	//	if !ev.Recurrent {
	//		continue
	//	}
	//
	//
	//
	//}

	return &events, nil
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

	rows, errSelect := p.DB.Queryx(`
									   SELECT ce.id,
                                              ce.title,
                                              ce."start",
                                              ce."end",
                                              ce.created,
                                              ce."owner",
                                              ce."channel",
                                              ce.recurrence,
                                              ce.color,
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
		Color:      eventDb.Color,
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

	events, eventsError := p.GetUserEvents(user, start, end)
	if eventsError != nil {
		errorResponse(w, eventsError)
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

	if event.Recurrence != "" && len(event.Recurrence) > 0 {
		event.Recurrent = true
	} else {
		event.Recurrent = false
	}

	_, errInsert := p.DB.NamedExec(`INSERT INTO PUBLIC.calendar_events
                                                  (id,
                                                   title,
                                                   "start",
                                                   "end",
                                                   created,
                                                   owner,
                                                   channel,
                                                   recurrent,
                                                   recurrence,
                                                   color)
                                      VALUES      (:id,
                                                   :title,
                                                   :start,
                                                   :end,
                                                   :created,
                                                   :owner,
                                                   :channel,
                                                   :recurrent,
                                                   :recurrence,
                                                   :color) `, &event)

	if errInsert != nil {
		p.API.LogError(errInsert.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	if len(event.Attendees) > 0 {
		var insertParams []map[string]interface{}
		for _, userId := range event.Attendees {
			insertParams = append(insertParams, map[string]interface{}{
				"event": event.Id,
				"user":  userId,
			})
		}

		_, errInsert = p.DB.NamedExec(`INSERT INTO public.calendar_members 
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

	_, dbErr := p.DB.Exec("DELETE FROM calendar_events WHERE id=$1", eventId)

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

func (p *Plugin) UpdateEvent(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	session, err := p.API.GetSession(c.SessionId)

	if err != nil {
		p.API.LogError(err.Error())
		errorResponse(w, NotAuthorizedError)
		return
	}

	user, userErr := p.API.GetUser(session.UserId)

	if userErr != nil {
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

	if event.Recurrence != "" && len(event.Recurrence) > 0 {
		event.Recurrent = true
	} else {
		event.Recurrent = false
	}

	tx, txError := p.DB.Beginx()

	if txError != nil {
		p.API.LogError(txError.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}
	_, errUpdate := tx.NamedExec(`UPDATE PUBLIC.calendar_events 
										SET 
										    title = :title,
										    "start" = :start,
										    "end" = :end,
										    channel = :channel,
										    recurrence = :recurrence,
										    recurrent = :recurrent,
										    color = :color
                              			WHERE id = :id`,
		&event)

	_, errUpdate = tx.Exec(`DELETE FROM calendar_members WHERE "event" = $1`, event.Id)

	if errUpdate != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			p.API.LogError(rollbackError.Error())
			return
		}
		p.API.LogError(errUpdate.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}

	if len(event.Attendees) > 0 {
		var insertParams []map[string]interface{}
		for _, userId := range event.Attendees {
			insertParams = append(insertParams, map[string]interface{}{
				"event": event.Id,
				"user":  userId,
			})
		}

		_, errUpdate = tx.NamedExec(`INSERT INTO public.calendar_members 
														  ("event", "user") 
												   VALUES (:event, :user)`, insertParams)
	}
	if errUpdate != nil {
		if rollbackError := tx.Rollback(); rollbackError != nil {
			p.API.LogError(rollbackError.Error())
			return
		}
		p.API.LogError(errUpdate.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}

	errUpdate = tx.Commit()

	apiResponse(w, &event)
	return
}

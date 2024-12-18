package main

import (
	"encoding/json"
	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/teambition/rrule-go"
	"net/http"
	time "time"
)

// GetUserEventsUTC returns user events in UTC timezone
// start and end are in format EventDateTimeLayout in UTC timezone
// if we don't have userLocation we can't correct gen dates for recurrent events, it means that we can't return recurrent events correctly
func (p *Plugin) GetUserEventsUTC(
	userId string,
	userLocation *time.Location,
	start, end time.Time,
) ([]Event, *model.AppError) {
	var events []Event

	conditions := sq.Or{
		sq.Eq{"cm.member": userId},
		sq.Eq{"ce.owner": userId},
		sq.And{
			sq.GtOrEq{"ce.dt_start": start},
			sq.LtOrEq{"ce.dt_start": end},
		},
		sq.Eq{"ce.recurrent": true},
	}

	// Create a new select builder
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.description",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.owner",
			"ce.channel",
			"ce.recurrent",
			"ce.recurrence",
			"ce.color",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(conditions).PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, args, err := queryBuilder.ToSql()
	if err != nil {
		p.API.LogError(err.Error())
		return nil, SomethingWentWrong
	}
	rows, errSelect := p.DB.Queryx(querySql, args...)

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
			p.API.LogError(errScan.Error())
			continue
		}
		if addedEvent[eventDb.Id] {
			continue
		}

		if eventDb.Color == nil {
			color := DefaultColor
			eventDb.Color = &color
		}

		if userLocation != nil {
			eventDb.Start = eventDb.Start.In(userLocation)
			eventDb.End = eventDb.End.In(userLocation)
		}

		if eventDb.Recurrent {
			eventRule, errRrule := rrule.StrToRRule(eventDb.Recurrence)
			if errRrule != nil {
				p.API.LogError(errRrule.Error())
				continue
			}
			eventRule.DTStart(eventDb.Start)
			eventDates := eventRule.Between(
				time.Date(
					start.Year(),
					start.Month(),
					start.Day(),
					0,
					0,
					0,
					0,
					start.Location(),
				),
				end, false)

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

	return events, nil
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

func (p *Plugin) GetEvent(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)
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

	query := mux.Vars(r)

	eventId := query["eventId"]

	if eventId == "" {
		errorResponse(w, InvalidRequestParams)
		return
	}
	queryBuilder := sq.Select().
		Columns(
			"ce.id",
			"ce.title",
			"ce.dt_start",
			"ce.dt_end",
			"ce.created",
			"ce.owner",
			"ce.channel",
			"ce.recurrence",
			"ce.color",
			"ce.description",
			"cm.member",
		).
		From("calendar_events ce").
		LeftJoin("calendar_members cm ON ce.id = cm.event").
		Where(sq.Eq{"id": eventId}).PlaceholderFormat(p.GetDBPlaceholderFormat())

	querySql, sqlArgs, toSqlErr := queryBuilder.ToSql()
	if toSqlErr != nil {
		errorResponse(w, InvalidRequestParams)
		return
	}
	rows, errSelect := p.DB.Queryx(querySql, sqlArgs...)
	if errSelect != nil {
		p.API.LogError("Selecting data error")
		errorResponse(w, EventNotFound)
		return
	}

	type EventFromDb struct {
		Event
		User *string `json:"user" db:"member"`
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
		Id:          eventDb.Id,
		Title:       eventDb.Title,
		Description: eventDb.Description,
		Start:       eventDb.Start,
		End:         eventDb.End,
		Attendees:   members,
		Created:     eventDb.Created,
		Owner:       eventDb.Owner,
		Channel:     eventDb.Channel,
		Recurrence:  eventDb.Recurrence,
		Color:       eventDb.Color,
	}

	userLoc := p.GetUserLocation(user)

	event.Start = event.Start.In(userLoc)
	event.End = event.End.In(userLoc)

	apiResponse(w, &event)
	return

}

func (p *Plugin) GetEvents(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)

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

	startEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, start, userLoc)
	EndEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, end, userLoc)

	events, eventsError := p.GetUserEventsUTC(
		user.Id, userLoc, startEventLocal.In(time.UTC), EndEventLocal.In(time.UTC),
	)
	if eventsError != nil {
		errorResponse(w, eventsError)
	}
	apiResponse(w, &events)
	return
}

func (p *Plugin) CreateEvent(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())

	session, err := p.API.GetSession(pluginContext.SessionId)

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

	event.Start = startDateInLocalTimeZone.In(time.UTC)
	event.End = endDateInLocalTimeZone.In(time.UTC)

	if event.Recurrence != "" && len(event.Recurrence) > 0 {
		event.Recurrent = true
	} else {
		event.Recurrent = false
	}

	queryBuilder := sq.Insert("calendar_events").
		Columns(
			"id",
			"title",
			"description",
			"dt_start",
			"dt_end",
			"created",
			"owner",
			"channel",
			"recurrent",
			"recurrence",
			"color",
		).
		Values(
			event.Id,
			event.Title,
			event.Description,
			event.Start,
			event.End,
			event.Created,
			event.Owner,
			event.Channel,
			event.Recurrent,
			event.Recurrence,
			event.Color,
		).PlaceholderFormat(p.GetDBPlaceholderFormat())

	// Prepare the SQL query
	querySql, sqlArgs, errBuilder := queryBuilder.ToSql()
	if errBuilder != nil {
		p.API.LogError(err.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	_, errInsert := p.DB.Queryx(querySql, sqlArgs...)

	if errInsert != nil {
		p.API.LogError(errInsert.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	if len(event.Attendees) > 0 {
		builderAtt := sq.Insert("calendar_members").
			Columns("event", "member")
		for _, userId := range event.Attendees {
			builderAtt = builderAtt.Values(event.Id, userId)
		}

		queryAttendees, queryAttArgs, errAttendees := builderAtt.PlaceholderFormat(p.GetDBPlaceholderFormat()).ToSql()
		if errAttendees != nil {
			p.API.LogError(err.Error())
			errorResponse(w, CantCreateEvent)
			return
		}
		_, errInsert = p.DB.Queryx(queryAttendees, queryAttArgs...)
	}

	if errInsert != nil {
		p.API.LogError(errInsert.Error())
		errorResponse(w, CantCreateEvent)
		return
	}

	apiResponse(w, &event)
	return
}

func (p *Plugin) RemoveEvent(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	_, err := p.API.GetSession(pluginContext.SessionId)

	if err != nil {
		p.API.LogError("can't get session")
		errorResponse(w, NotAuthorizedError)
		return
	}

	query := mux.Vars(r)

	eventId := query["eventId"]

	if eventId == "" {
		errorResponse(w, InvalidRequestParams)
		return
	}

	deleteBuilder := sq.Delete("calendar_events").
		Where(sq.Eq{"id": eventId}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())
	deleteSql, deleteArgs, deleteErr := deleteBuilder.ToSql()

	if deleteErr != nil {
		p.API.LogError("can't remove event from db")
		p.API.LogError(deleteErr.Error())
		errorResponse(w, CantRemoveEvent)
		return
	}
	_, dbErr := p.DB.Queryx(deleteSql, deleteArgs...)

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

func (p *Plugin) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	pluginContext := p.FromContext(r.Context())
	session, err := p.API.GetSession(pluginContext.SessionId)

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

	event.Start = startDateInLocalTimeZone.In(time.UTC)
	event.End = endDateInLocalTimeZone.In(time.UTC)

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

	updateFields := map[string]interface{}{
		"title":       event.Title,
		"description": event.Description,
		"dt_start":    event.Start,
		"dt_end":      event.End,
		"channel":     event.Channel,
		"recurrence":  event.Recurrence,
		"recurrent":   event.Recurrent,
		"color":       event.Color,
	}
	updateQueryBuilder := sq.Update("calendar_events").
		SetMap(updateFields).
		Where(sq.Eq{"id": event.Id}).
		PlaceholderFormat(p.GetDBPlaceholderFormat())

	updateSql, updateArgs, _ := updateQueryBuilder.ToSql()

	rows, errUpdate := tx.Queryx(updateSql, updateArgs...)
	if errUpdate != nil {
		p.API.LogInfo("cant update calendar event: " + errUpdate.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}
	if rows != nil {
		rows.Close()
	}

	deleteBuilder := sq.Delete("calendar_members").Where(sq.Eq{"event": event.Id})
	deleteSql, deleteArgs, deleteErr := deleteBuilder.PlaceholderFormat(p.GetDBPlaceholderFormat()).ToSql()
	if deleteErr != nil {
		p.API.LogError(deleteErr.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}
	rows, errDelete := tx.Queryx(deleteSql, deleteArgs...)
	if errDelete != nil {
		p.API.LogError(errDelete.Error())
		if rollbackError := tx.Rollback(); rollbackError != nil {
			p.API.LogError(rollbackError.Error())
		}
		errorResponse(w, CantUpdateEvent)
		return
	}
	if rows != nil {
		rows.Close()
	}

	if len(event.Attendees) > 0 {
		attQueryBuilder := sq.Insert("calendar_members").Columns("event", "member")
		for _, userId := range event.Attendees {
			attQueryBuilder = attQueryBuilder.Values(event.Id, userId)
		}
		attUpdateSql, attArgs, _ := attQueryBuilder.PlaceholderFormat(p.GetDBPlaceholderFormat()).ToSql()
		rows, errUpdateAtt := tx.Queryx(attUpdateSql, attArgs...)

		if errUpdateAtt != nil {
			p.API.LogError(errUpdateAtt.Error())
			if rollbackError := tx.Rollback(); rollbackError != nil {
				p.API.LogError(rollbackError.Error())
			}

			errorResponse(w, CantUpdateEvent)
			return
		}
		if rows != nil {
			rows.Close()
		}
	}

	if commitError := tx.Commit(); commitError != nil {
		p.API.LogError("commit error" + commitError.Error())
		errorResponse(w, CantUpdateEvent)
		return
	}

	apiResponse(w, &event)
	return
}

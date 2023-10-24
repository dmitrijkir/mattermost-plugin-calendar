package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/teambition/rrule-go"
	"net/http"
	"strings"
	"sync"
	"time"
)

// GetUserEventsUTC returns user events in UTC timezone
// start and end are in format EventDateTimeLayout in UTC timezone
func (p *Plugin) GetUserEventsUTC(userId string, start, end time.Time) ([]Event, *model.AppError) {
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
                                       `, userId, userId, start, end)

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

type UserScheduleEvent struct {
	Start    time.Time `json:"start" db:"start"`
	End      time.Time `json:"end" db:"end"`
	Duration int32     `json:"duration"`
}

type GetScheduleResponse struct {
	Users          map[string][]UserScheduleEvent `json:"users"`
	AvailableTimes []string                       `json:"available_times"`
}

func (p *Plugin) GetSchedule(w http.ResponseWriter, r *http.Request) {
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

	if query.Get("users") == "" {
		http.Error(w, "users is required", http.StatusBadRequest)
		return
	}
	if query.Get("start") == "" || query.Get("end") == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}

	users := strings.Split(query.Get("users"), ",")

	userLoc := p.GetUserLocation(user)
	utcLoc, _ := time.LoadLocation("UTC")

	// start, end event in user request location
	startEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, query.Get("start"), userLoc)
	EndEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, query.Get("end"), userLoc)

	usersEvents := make(map[string][]UserScheduleEvent)
	wg := &sync.WaitGroup{}
	wg.Add(len(users))

	for _, userId := range users {
		go func(userId string) {
			defer wg.Done()

			userEvents, err := p.GetUserEventsUTC(userId, startEventLocal.In(utcLoc), EndEventLocal.In(utcLoc))
			if err != nil {
				p.API.LogError("can't get schedule for user")
				return
			}

			userSchEvents := []UserScheduleEvent{}
			// convert event utc time to user location
			for _, event := range userEvents {
				userSchEvents = append(userSchEvents, UserScheduleEvent{
					Start:    event.Start.In(userLoc),
					End:      event.End.In(userLoc),
					Duration: int32(event.End.Sub(event.Start).Minutes()),
				})
			}
			usersEvents[userId] = userSchEvents
		}(userId)
	}

	wg.Wait()

	response := &GetScheduleResponse{
		Users:          usersEvents,
		AvailableTimes: []string{},
	}
	apiResponse(w, response)
	return
}

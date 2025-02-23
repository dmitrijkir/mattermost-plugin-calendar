package main

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

	// time slot part it event duration in minutes / DefaultSlotTime
	slotParts := 1
	if query.Get("slot_time") != "" {
		slotTime, errConvert := strconv.Atoi(query.Get("slot_time"))
		slotParts = int(math.Ceil(float64(slotTime / DefaultSlotTime)))
		if slotParts <= 0 {
			slotParts = 1
		}
		if errConvert != nil {
			http.Error(w, "bad slot time request", http.StatusBadRequest)
			return
		}

	}

	users := strings.Split(query.Get("users"), ",")

	userLoc := p.GetUserLocation(user)

	// start, end event in user request location
	startEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, query.Get("start"), userLoc)
	EndEventLocal, _ := time.ParseInLocation(EventDateTimeLayout, query.Get("end"), userLoc)

	usersEvents := make(map[string][]UserScheduleEvent)
	usersBusyTimes := make([][]bool, 0)
	usersBusyTimesSync := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	wg.Add(len(users))

	for _, userId := range users {
		go func(userId string) {
			defer wg.Done()

			userEvents, err := p.GetUserEventsUTC(userId, userLoc, startEventLocal.In(time.UTC), EndEventLocal.In(time.UTC))
			if err != nil {
				p.API.LogError("can't get schedule for user")
				return
			}

			userBusyTime := make([]bool, 96)
			var userSchEvents []UserScheduleEvent
			// convert event utc time to user location
			for _, event := range userEvents {
				userEvent := UserScheduleEvent{
					Start:    event.Start.In(userLoc),
					End:      event.End.In(userLoc),
					Duration: int32(event.End.Sub(event.Start).Minutes()),
				}

				// convert event start time to slot position,for recurent event start time is event start time in user location
				eventStart := time.Date(
					startEventLocal.Year(),
					startEventLocal.Month(),
					startEventLocal.Day(),
					userEvent.Start.Hour(),
					userEvent.Start.Minute(),
					userEvent.Start.Second(),
					userEvent.Start.Nanosecond(),
					userEvent.Start.Location(),
				)
				startPosition := int32(math.Floor(eventStart.Sub(startEventLocal).Minutes() / DefaultSlotTime))
				slotCount := int32(math.Ceil(event.End.Sub(event.Start).Minutes() / DefaultSlotTime))

				for i := startPosition; i <= startPosition+slotCount-1; i++ {
					userBusyTime[i] = true
				}

				usersBusyTimesSync.Lock()
				usersBusyTimes = append(usersBusyTimes, userBusyTime)
				usersBusyTimesSync.Unlock()
				userSchEvents = append(userSchEvents, userEvent)
			}
			usersEvents[userId] = userSchEvents
		}(userId)
	}

	wg.Wait()

	availableTimes := make([]string, 0)

	// find free time
	for j := 0; j <= 95; j++ {
		isTimelineValid := true
		for i := 0; i <= len(usersBusyTimes)-1; i++ {
			for k := 0; k <= slotParts-1; k++ {
				if j+k > 95 {
					break
				}
				if usersBusyTimes[i][j+k] == false {
					continue
				}
				isTimelineValid = false
				break
			}
		}
		if isTimelineValid {
			freeTime := startEventLocal.Add(time.Minute * time.Duration(j) * DefaultSlotTime)
			availableTimes = append(availableTimes, freeTime.Format("15:04"))
		}
	}

	response := &GetScheduleResponse{
		Users:          usersEvents,
		AvailableTimes: availableTimes,
	}
	apiResponse(w, response)
	return
}

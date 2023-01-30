package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"time"
)

type Background struct {
	Ticker *time.Ticker
	Done   chan bool
	plugin *Plugin
	botId  string
}

func (b *Background) Start() {
	for {
		select {
		case <-b.Done:
			return
		case t := <-b.Ticker.C:
			b.process(&t)
		}
	}
}

func (b *Background) Stop() {
	b.Done <- true
}

func (b *Background) process(t *time.Time) {
	utcLoc, _ := time.LoadLocation("UTC")

	tickWithZone := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		0,
		0,
		utcLoc,
	)
	rows, errSelect := GetDb().Queryx(`
			SELECT ce.id,
				   ce.title,
                   ce."start",
                   ce."end",
                   ce.created,
                   ce."owner",
                   ce."channel",
                   cm."user",
                   ce.recurrent,
                   ce.recurrence
			FROM   calendar_events ce
                FULL JOIN calendar_members cm
                       ON ce.id = cm."event"
			WHERE (ce."start" = $1 OR (ce.recurrent = true AND ce."start"::time = $2)) 
			  	   AND (ce."processed" isnull OR ce."processed" != $3)
`, tickWithZone, tickWithZone, tickWithZone)

	if errSelect != nil {
		b.plugin.API.LogError(errSelect.Error())
		return
	}

	type EventFromDb struct {
		Event
		User *string `json:"user" db:"user"`
	}
	events := map[string]*Event{}

	for rows.Next() {
		var eventDb EventFromDb

		errScan := rows.StructScan(&eventDb)

		if errScan != nil {
			b.plugin.API.LogError(errSelect.Error())
			continue
		}

		if eventDb.User == nil {
			continue
		}

		if events[eventDb.Id] != nil {
			events[eventDb.Id].Attendees = append(events[eventDb.Id].Attendees, *eventDb.User)
		} else {

			if eventDb.Recurrent && contains(*eventDb.Recurrence, int(t.Weekday())) {
				eventTime := eventDb.End.Sub(eventDb.Start)
				eventDb.Start = time.Date(
					t.Year(),
					t.Month(),
					t.Day(),
					eventDb.Start.Hour(),
					eventDb.Start.Minute(),
					eventDb.Start.Second(),
					eventDb.Start.Nanosecond(),
					eventDb.Start.Location(),
				)
				eventDb.End = eventDb.Start.Add(eventTime)
			}
			events[eventDb.Id] = &Event{
				Id:         eventDb.Id,
				Title:      eventDb.Title,
				Start:      eventDb.Start,
				End:        eventDb.End,
				Attendees:  []string{*eventDb.User},
				Created:    eventDb.Created,
				Owner:      eventDb.Owner,
				Channel:    eventDb.Channel,
				Recurrence: eventDb.Recurrence,
				Recurrent:  false,
			}

		}
	}

	for _, value := range events {
		if value.Channel != nil {
			_, postErr := b.plugin.API.CreatePost(&model.Post{
				ChannelId: *value.Channel,
				Message:   value.Title,
				UserId:    b.botId,
			})
			if postErr != nil {
				b.plugin.API.LogError(postErr.Error())
				continue
			}

		} else {

			foundChannel, foundChannelError := b.plugin.API.GetGroupChannel(value.Attendees)
			if foundChannelError != nil {
				b.plugin.API.LogError(foundChannelError.Error())
				continue
			}

			_, postCreateError := b.plugin.API.CreatePost(&model.Post{
				UserId:    b.botId,
				Message:   value.Title,
				ChannelId: foundChannel.Id,
			})
			if postCreateError != nil {
				b.plugin.API.LogError(postCreateError.Error())
				continue
			}
		}

		_, errUpdate := GetDb().NamedExec(`UPDATE PUBLIC.calendar_events
                                           SET processed = :processed
                                           WHERE id = :eventId`, map[string]interface{}{
			"processed": tickWithZone,
			"eventId":   value.Id,
		})

		if errUpdate != nil {
			b.plugin.API.LogError(errUpdate.Error())
			continue
		}

	}

}

func NewBackgroundJob(plugin *Plugin, userId string) *Background {
	return &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: plugin,
		botId:  userId,
	}
}

var bgJob *Background

func GetBackgroundJob() *Background {
	return bgJob
}

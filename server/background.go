package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mattermost/mattermost-server/v6/model"
	"time"
)

type Background struct {
	Ticker *time.Ticker
	Done   chan bool
	plugin *Plugin
	botId  string
	DB     *sqlx.DB
}

func (b *Background) SetDb(db *sqlx.DB) {
	b.DB = db
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

func (b *Background) getMessageFromEvent(event *Event) string {
	message := ""
	message += fmt.Sprintf(":dart: *%s* :dart:\n", event.Title)

	if len(event.Attendees) > 0 {
		members := ""
		for _, member := range event.Attendees {
			user, userErr := b.plugin.API.GetUser(member)

			if userErr != nil {
				continue
			}

			members += fmt.Sprintf("@%s, ", user.Username)
		}
		message += fmt.Sprintf("**members:** %s\n", members)
	}

	return message
}

func (b *Background) sendGroupOrPersonalEventNotification(event *Event) {
	var attendees []string

	attendees = append(attendees, event.Attendees...)

	if len(attendees) == 0 {
		dChannel, dChannelErr := b.plugin.API.GetDirectChannel(event.Owner, b.botId)
		if dChannelErr != nil {
			b.plugin.API.LogError(dChannelErr.Error())
			return
		}

		_, postCreateError := b.plugin.API.CreatePost(&model.Post{
			UserId:    b.botId,
			Message:   b.getMessageFromEvent(event),
			ChannelId: dChannel.Id,
		})
		if postCreateError != nil {
			b.plugin.API.LogError(postCreateError.Error())
			return
		}

		return
	}

	if !contains[string](attendees, event.Owner) {
		attendees = append(attendees, event.Owner)
	}

	attendees = append(attendees, b.botId)

	foundChannel, foundChannelError := b.plugin.API.GetGroupChannel(attendees)
	if foundChannelError != nil {
		b.plugin.API.LogError(foundChannelError.Error())
		return
	}

	_, postCreateError := b.plugin.API.CreatePost(&model.Post{
		UserId:    b.botId,
		Message:   b.getMessageFromEvent(event),
		ChannelId: foundChannel.Id,
	})
	if postCreateError != nil {
		b.plugin.API.LogError(postCreateError.Error())
		return
	}
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
	rows, errSelect := b.DB.Queryx(`
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

		if events[eventDb.Id] != nil && eventDb.User != nil {
			events[eventDb.Id].Attendees = append(events[eventDb.Id].Attendees, *eventDb.User)
		} else {

			if eventDb.Recurrent && !contains[int](*eventDb.Recurrence, int(t.Weekday())) {
				continue
			}
			if eventDb.Recurrent {
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

			var att []string
			if eventDb.User != nil {
				att = append(att, *eventDb.User)
			}
			events[eventDb.Id] = &Event{
				Id:         eventDb.Id,
				Title:      eventDb.Title,
				Start:      eventDb.Start,
				End:        eventDb.End,
				Attendees:  att,
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
				Message:   b.getMessageFromEvent(value),
				UserId:    b.botId,
			})
			if postErr != nil {
				b.plugin.API.LogError(postErr.Error())
				continue
			}

		} else {
			b.sendGroupOrPersonalEventNotification(value)
		}

		_, errUpdate := b.DB.NamedExec(`UPDATE PUBLIC.calendar_events
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

func NewBackgroundJob(plugin *Plugin, userId string, db *sqlx.DB) *Background {
	return &Background{
		Ticker: time.NewTicker(15 * time.Second),
		Done:   make(chan bool),
		plugin: plugin,
		botId:  userId,
		DB:     db,
	}
}

var bgJob *Background

func GetBackgroundJob() *Background {
	return bgJob
}

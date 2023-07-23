package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/teambition/rrule-go"
	"time"
)

const wsEventOccur = "event_occur"

type Background struct {
	Ticker *time.Ticker
	Done   chan bool
	plugin *Plugin
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
		dChannel, dChannelErr := b.plugin.API.GetDirectChannel(event.Owner, b.plugin.BotId)
		if dChannelErr != nil {
			b.plugin.API.LogError(dChannelErr.Error())
			return
		}
		postModel := &model.Post{
			UserId:    b.plugin.BotId,
			ChannelId: dChannel.Id,
		}
		postModel.SetProps(b.getMessageProps(event))
		_, postCreateError := b.plugin.API.CreatePost(postModel)
		if postCreateError != nil {
			b.plugin.API.LogError(postCreateError.Error())
			return
		}
		return
	}

	if !contains[string](attendees, event.Owner) {
		attendees = append(attendees, event.Owner)
	}

	attendees = append(attendees, b.plugin.BotId)

	foundChannel, foundChannelError := b.plugin.API.GetGroupChannel(attendees)
	if foundChannelError != nil {
		b.plugin.API.LogError(foundChannelError.Error())
		return
	}

	postModel := &model.Post{
		UserId:    b.plugin.BotId,
		ChannelId: foundChannel.Id,
	}

	postModel.SetProps(b.getMessageProps(event))

	_, postCreateError := b.plugin.API.CreatePost(postModel)
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
	rows, errSelect := b.plugin.DB.Queryx(`
			SELECT ce.id,
				   ce.title,
                   ce."start",
                   ce."end",
                   ce.created,
                   ce."owner",
                   ce."channel",
                   cm."user",
                   ce.recurrent,
                   ce.recurrence,
                   ce.color
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
			if eventDb.Recurrent {
				eventRule, errRrule := rrule.StrToRRule(eventDb.Recurrence)
				if errRrule != nil {
					b.plugin.API.LogError(errRrule.Error())
					continue
				}
				eventTime := eventDb.End.Sub(eventDb.Start)
				eventEnd := tickWithZone.Add(eventTime)
				eventRule.DTStart(time.Date(
					eventDb.Start.Year(),
					eventDb.Start.Month(),
					eventDb.Start.Day(),
					0,
					0,
					0,
					0,
					utcLoc,
				))
				eventDates := eventRule.Between(
					time.Date(
						tickWithZone.Year(),
						tickWithZone.Month(),
						tickWithZone.Day(),
						0,
						0,
						0,
						0,
						utcLoc,
					),
					eventEnd,
					true)
				// Skip this event if recurrent event doesn't exist between two dates
				if len(eventDates) < 1 {
					continue
				}
				recEventTime := eventDb.End.Sub(eventDb.Start)
				eventDb.Start = time.Date(
					tickWithZone.Year(),
					tickWithZone.Month(),
					tickWithZone.Day(),
					eventDb.Start.Hour(),
					eventDb.Start.Minute(),
					eventDb.Start.Second(),
					eventDb.Start.Nanosecond(),
					eventDb.Start.Location(),
				)
				eventDb.End = eventDb.Start.Add(recEventTime)
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
				Color:      eventDb.Color,
			}

		}
	}

	for _, value := range events {
		go b.sendWsNotification(value)
		if value.Channel != nil {
			postModel := &model.Post{
				ChannelId: *value.Channel,
				UserId:    b.plugin.BotId,
			}

			postModel.SetProps(b.getMessageProps(value))
			_, postErr := b.plugin.API.CreatePost(postModel)
			if postErr != nil {
				b.plugin.API.LogError(postErr.Error())
				continue
			}

		} else {
			b.sendGroupOrPersonalEventNotification(value)
		}

		_, errUpdate := b.plugin.DB.NamedExec(`UPDATE PUBLIC.calendar_events
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

func (b *Background) sendWsNotification(event *Event) {
	var attendees []string

	attendees = append(attendees, event.Attendees...)
	if !contains(attendees, event.Owner) {
		attendees = append(attendees, event.Owner)
	}

	for _, user := range attendees {
		b.plugin.API.PublishWebSocketEvent(wsEventOccur, map[string]interface{}{
			"id":      event.Id,
			"title":   event.Title,
			"channel": nil,
		}, &model.WebsocketBroadcast{
			UserId: user,
		})
	}

}
func (b *Background) getMessageProps(event *Event) model.StringInterface {
	color := DefaultColor
	if event.Color == nil {
		event.Color = &color
	}

	slackAttachment := model.SlackAttachment{
		Text:  b.getMessageFromEvent(event),
		Color: *event.Color,
	}

	return model.StringInterface{
		"attachments": []*model.SlackAttachment{&slackAttachment},
	}
}

var bgJob *Background

func NewBackgroundJob(plugin *Plugin) *Background {
	if bgJob == nil {
		bgJob = &Background{
			Ticker: time.NewTicker(15 * time.Second),
			Done:   make(chan bool),
			plugin: plugin,
		}
	}
	return bgJob
}

func GetBackgroundJob() *Background {
	return bgJob
}

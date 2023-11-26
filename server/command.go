package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"sort"
	"strings"
	"time"
)

const calCommand = "cal"

func (p *Plugin) createCalCommand() (*model.Command, error) {
	return &model.Command{
		Trigger:          calCommand,
		AutoComplete:     true,
		AutoCompleteDesc: "Get calendar events.",
		AutoCompleteHint: "[command]",
	}, nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	command := split[0]
	action := ""
	if len(split) > 1 {
		action = split[1]
	}

	if command != "/"+calCommand {
		return &model.CommandResponse{}, nil
	}

	switch action {
	case "help":
		p.API.LogError("======help=======")
	case "week":
		return p.executeWeekCommand(c, args)
	default:
		return p.executeTodayCommand(c, args)
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) executeTodayCommand(
	c *plugin.Context,
	args *model.CommandArgs,
) (*model.CommandResponse, *model.AppError) {
	user, appErr := p.API.GetUser(args.UserId)
	if appErr != nil {
		return nil, NotAuthorizedError
	}

	userLoc := p.GetUserLocation(user)

	now := time.Now().UTC().In(userLoc)

	start := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		userLoc,
	)

	end := start.Add(time.Hour * 24)

	events, eventsError := p.GetUserEventsUTC(user.Id, userLoc, start.In(time.UTC), end.In(time.UTC))

	if eventsError != nil {
		p.API.LogError(eventsError.Error())
		return nil, eventsError
	}

	for ind, event := range events {
		events[ind].Start = event.Start.In(userLoc)
		events[ind].End = event.End.In(userLoc)
	}

	message := "| time | title | channel |\n| -----| ------| ------- |\n"
	for _, event := range events {
		line := fmt.Sprintf("|%s|%s|", event.Start.Format(EventDateTimeLayout), event.Title)
		if event.Channel != nil {
			eventChannel, eventChError := p.API.GetChannel(*event.Channel)
			if eventChError != nil {
				continue
			}
			line += fmt.Sprintf("%s|", eventChannel.DisplayName)
		} else {
			line += fmt.Sprintf("%s|", "empty")
		}
		message += fmt.Sprintf("%s\n", line)
	}

	dChannel, dChannelErr := p.API.GetDirectChannel(user.Id, p.BotId)
	if dChannelErr != nil {
		p.API.LogError(dChannelErr.Error())
		return nil, SomethingWentWrong
	}

	_, postCreateError := p.API.CreatePost(&model.Post{
		UserId:    p.BotId,
		Message:   message,
		ChannelId: dChannel.Id,
	})
	if postCreateError != nil {
		return nil, postCreateError
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) executeWeekCommand(
	c *plugin.Context,
	args *model.CommandArgs,
) (*model.CommandResponse, *model.AppError) {
	user, appErr := p.API.GetUser(args.UserId)
	if appErr != nil {
		return nil, NotAuthorizedError
	}

	userLoc := p.GetUserLocation(user)

	now := time.Now().UTC().In(userLoc)

	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		userLoc,
	)

	start := today.Add(-time.Hour * 24 * time.Duration(today.Weekday()))

	end := start.Add(time.Hour * 24 * 7)

	events, eventsError := p.GetUserEventsUTC(user.Id, userLoc, start.In(time.UTC), end.In(time.UTC))

	if eventsError != nil {
		p.API.LogError(eventsError.Error())
		return nil, eventsError
	}

	for ind, event := range events {
		events[ind].Start = event.Start.In(userLoc)
		events[ind].End = event.End.In(userLoc)
	}

	message := "| time | title | channel |\n| -----| ------| ------- |\n"
	sort.Slice(events, func(i, j int) bool {
		return (events)[j].Start.After((events)[i].Start)
	})
	for _, event := range events {
		line := fmt.Sprintf("|%s|%s|", event.Start.Format(EventDateTimeLayout), event.Title)
		if event.Channel != nil {
			eventChannel, eventChError := p.API.GetChannel(*event.Channel)
			if eventChError != nil {
				p.API.LogError(eventChError.Error())
				continue
			}
			line += fmt.Sprintf("%s|", eventChannel.DisplayName)
		} else {
			line += fmt.Sprintf("%s|", "empty")
		}
		message += fmt.Sprintf("%s\n", line)
	}

	dChannel, dChannelErr := p.API.GetDirectChannel(user.Id, p.BotId)
	if dChannelErr != nil {
		p.API.LogError(dChannelErr.Error())
		return nil, SomethingWentWrong
	}

	_, postCreateError := p.API.CreatePost(&model.Post{
		UserId:    p.BotId,
		Message:   message,
		ChannelId: dChannel.Id,
	})
	if postCreateError != nil {
		return nil, postCreateError
	}
	return &model.CommandResponse{}, nil
}

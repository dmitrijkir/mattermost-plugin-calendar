package main

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"testing"
	"time"
)

// Test personal notification
func TestSendGroupOrPersonalEventNotification(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	testEvent := &Event{
		Id:        "efe-fe",
		Title:     "test event for channel",
		Start:     time.Now(),
		End:       time.Now(),
		Attendees: []string{},
		Created:   time.Now(),
		Owner:     "owner-id",
		Channel:   &channelId,
		Processed: nil,
		Recurrent: false,
	}

	foundChannel := &model.Channel{
		Id: channelId,
	}

	postForSend := &model.Post{
		UserId:    botId,
		Message:   testEvent.Title,
		ChannelId: channelId,
	}

	api := plugintest.API{}
	api.On("GetDirectChannel", testEvent.Owner, botId).Return(foundChannel, nil)

	api.On("CreatePost", postForSend).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	background := NewBackgroundJob(pluginT, botId)

	background.sendGroupOrPersonalEventNotification(testEvent)

}

// Test group notification
func TestSendGroupOrPersonalEventGroupNotification(t *testing.T) {
	botId := "bot-id"
	channelId := "channel-id"
	attendees := []string{"first-id", "second-id"}
	testEvent := &Event{
		Id:        "efe-fe",
		Title:     "test event for channel",
		Start:     time.Now(),
		End:       time.Now(),
		Attendees: attendees,
		Created:   time.Now(),
		Owner:     "owner-id",
		Channel:   &channelId,
		Processed: nil,
		Recurrent: false,
	}

	foundChannel := &model.Channel{
		Id: channelId,
	}

	postForSend := &model.Post{
		UserId:    botId,
		Message:   testEvent.Title,
		ChannelId: channelId,
	}

	api := plugintest.API{}

	attendees = append(attendees, testEvent.Owner)
	attendees = append(attendees, botId)
	api.On("GetGroupChannel", attendees).Return(foundChannel, nil)

	api.On("CreatePost", postForSend).Return(nil, nil)

	pluginT := &Plugin{
		MattermostPlugin: plugin.MattermostPlugin{
			API:    &api,
			Driver: nil,
		},
	}

	background := NewBackgroundJob(pluginT, botId)

	background.sendGroupOrPersonalEventNotification(testEvent)
}

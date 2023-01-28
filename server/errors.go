package main

import "github.com/mattermost/mattermost-server/v6/model"

var (
	NotAuthorizedError = &model.AppError{
		Id:         "not_authorized",
		Message:    "Not authorized",
		StatusCode: 401,
		Where:      PluginId,
	}

	UserNotFound = &model.AppError{
		Id:         "user_not_found",
		Message:    "User not found",
		StatusCode: 404,
		Where:      PluginId,
	}

	EventNotFound = &model.AppError{
		Id:         "event_not_found",
		Message:    "Event not found",
		StatusCode: 404,
		Where:      PluginId,
	}

	InvalidRequestParams = &model.AppError{
		Id:         "invalid_or_missing_request_params",
		Message:    "Invalid or missing parameters in URL or request body",
		StatusCode: 400,
		Where:      PluginId,
	}

	CantCreateEvent = &model.AppError{
		Id:         "cant_create_event",
		Message:    "Can't create new event",
		StatusCode: 500,
		Where:      PluginId,
	}

	CantRemoveEvent = &model.AppError{
		Id:         "cant_remove_event",
		Message:    "Can't remove event",
		StatusCode: 500,
		Where:      PluginId,
	}

	SomethingWentWrong = &model.AppError{
		Id:         "something_went_wrong",
		Message:    "Something went wrong",
		StatusCode: 500,
		Where:      PluginId,
	}
)

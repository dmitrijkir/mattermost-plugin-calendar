package main

import (
	"database/sql"
	"encoding/json"
	sq "github.com/Masterminds/squirrel"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (p *Plugin) GetSettings(w http.ResponseWriter, r *http.Request) {
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

	userLoc := p.GetUserLocation(user)

	now := time.Now()
	BusinessStartTimeUtc, _ := time.ParseInLocation(
		BusinessTimeLayout, p.configuration.BusinessStartTime, time.UTC,
	)
	BusinessEndTimeUtc, _ := time.ParseInLocation(
		BusinessTimeLayout, p.configuration.BusinessEndTime, time.UTC,
	)

	// Add new year for new time formatting. Old format is LMT. Now use GMT. Problem with location
	BusinessStartTimeUtc = time.Date(
		now.Year(),
		BusinessStartTimeUtc.Month(),
		BusinessStartTimeUtc.Day(),
		BusinessStartTimeUtc.Hour(),
		BusinessStartTimeUtc.Minute(),
		BusinessStartTimeUtc.Second(),
		BusinessStartTimeUtc.Nanosecond(),
		BusinessStartTimeUtc.Location(),
	)

	BusinessEndTimeUtc = time.Date(
		now.Year(),
		BusinessEndTimeUtc.Month(),
		BusinessEndTimeUtc.Day(),
		BusinessEndTimeUtc.Hour(),
		BusinessEndTimeUtc.Minute(),
		BusinessEndTimeUtc.Second(),
		BusinessEndTimeUtc.Nanosecond(),
		BusinessEndTimeUtc.Location(),
	)

	var businessDays []int

	for _, i := range strings.Split(p.configuration.BusinessDays, ",") {
		day, err := strconv.Atoi(i)
		if err != nil {
			p.API.LogError(err.Error())
		}
		businessDays = append(businessDays, day)
	}

	userSettings := UserSettings{
		BusinessStartTime: BusinessStartTimeUtc.In(userLoc).Format(BusinessTimeLayout),
		BusinessEndTime:   BusinessEndTimeUtc.In(userLoc).Format(BusinessTimeLayout),
		BusinessDays:      businessDays,
	}

	queryBuilder := sq.Select().
		Columns("is_open_calendar_left_bar", "first_day_of_week", "hide_non_working_days").
		From("calendar_settings").
		Where(sq.Eq{"user": user.Id})

	querySql, argsSql, _ := queryBuilder.ToSql()
	errSelect := p.DB.Get(
		&userSettings,
		querySql,
		argsSql...,
	)

	// return default value
	if errSelect != nil {
		userSettings.IsOpenCalendarLeftBar = true
		userSettings.FirstDayOfWeek = 1
		userSettings.HideNonWorkingDays = false
		apiResponse(w, &userSettings)
		return
	}

	apiResponse(w, &userSettings)
	return
}

func (p *Plugin) UpdateSettings(w http.ResponseWriter, r *http.Request) {
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

	type UserSettingsRequest struct {
		IsOpenCalendarLeftBar bool `json:"isOpenCalendarLeftBar" db:"is_open_calendar_left_bar"`
		FirstDayOfWeek        int  `json:"firstDayOfWeek" db:"first_day_of_week"`
		HideNonWorkingDays    bool `json:"hideNonWorkingDays" db:"hide_non_working_days"`
	}

	var userSettings UserSettingsRequest
	var requestUserSettings UserSettingsRequest

	errDecode := json.NewDecoder(r.Body).Decode(&requestUserSettings)

	if errDecode != nil {
		p.API.LogError(errDecode.Error())
		errorResponse(w, InvalidRequestParams)
		return
	}

	if requestUserSettings.FirstDayOfWeek < 0 || requestUserSettings.FirstDayOfWeek > 6 {
		p.API.LogError(errDecode.Error())
		errorResponse(w, InvalidRequestParams)
		return
	}

	getQueryBuilder := sq.Select().
		Columns("is_open_calendar_left_bar", "first_day_of_week", "hide_non_working_days").
		From("calendar_settings").
		Where(sq.Eq{"user": user.Id})

	getQuerySql, getArgsSql, _ := getQueryBuilder.ToSql()
	errSelect := p.DB.Get(&userSettings, getQuerySql, getArgsSql)

	if errSelect == sql.ErrNoRows {
		insertQueryBuilder := sq.Insert("calendar_settings").
			Columns(
				"is_open_calendar_left_bar",
				"first_day_of_week",
				"hide_non_working_days",
				"user",
			).
			Values(
				requestUserSettings.IsOpenCalendarLeftBar,
				requestUserSettings.FirstDayOfWeek,
				requestUserSettings.HideNonWorkingDays,
				user.Id,
			)

		insertQuery, insertArgs, _ := insertQueryBuilder.ToSql()
		_, errInsert := p.DB.Queryx(insertQuery, insertArgs...)

		if errInsert != nil {
			p.API.LogError(err.Error())
			errorResponse(w, SomethingWentWrong)
			return
		}

		apiResponse(w, &userSettings)
		return
	}

	updateQueryBuilder := sq.Update("calendar_settings").
		Set("is_open_calendar_left_bar", requestUserSettings.IsOpenCalendarLeftBar).
		Set("first_day_of_week", requestUserSettings.FirstDayOfWeek).
		Set("hide_non_working_days", requestUserSettings.HideNonWorkingDays).
		Where(sq.Eq{"user": user.Id})

	updateQuery, updateArgs, _ := updateQueryBuilder.ToSql()
	_, errUpdate := p.DB.Queryx(updateQuery, updateArgs...)

	if errUpdate != nil {
		p.API.LogError(err.Error())
		errorResponse(w, SomethingWentWrong)
		return
	}

	apiResponse(w, &requestUserSettings)
	return
}

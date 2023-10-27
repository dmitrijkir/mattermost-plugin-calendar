import {DispatchFunc} from 'mattermost-redux/types/actions';

import {EventClickArg} from '@fullcalendar/common';

import {CalendarSettings} from './types/settings';
import {ApiClient} from './client';
import {CalendarEventNotification, SelectedEventTime} from './types/event';
import {UserProfile} from "mattermost-redux/types/users";
import {set} from "date-fns";

export const eventSelected = (event: EventClickArg) => {
    return {
        type: 'eventSelected',
        payload: event,
    };
};

export const openEventModal = () => {
    return {
        type: 'openEventModal',
        payload: true,
    };
};

export const closeEventModal = () => {
    return {
        type: 'closeEventModal',
        payload: false,
    };
};

export const updateCalendarSettings = (settings: CalendarSettings) => {
    return {
        type: 'updateCalendarSettings',
        payload: settings,
    };
};

export function updateCalendarSettingsOnServer(settings: CalendarSettings) {
    return async (dispatch: DispatchFunc) => {
        dispatch(updateCalendarSettings(settings));
        await ApiClient.updateCalendarSettings(settings);
    };
}

export const eventNotification = (event: CalendarEventNotification) => {
    return {
        type: 'eventNotification',
        payload: event,
    };
};

export const updateMembersAddedInEvent = (members: UserProfile[]) => {
    return {
        type: 'updateMembersAddedInEvent',
        payload: members,
    };
};

export const updateSelectedEventTime = (event: SelectedEventTime) => {
    const startT = event.startTime.split(':');
    const endT = event.endTime.split(':');
    event.start = set(event.start, {hours: parseInt(startT[0]), minutes: parseInt(startT[1])});
    event.end = set(event.end, {hours: parseInt(endT[0]), minutes: parseInt(endT[1])});
    return {
        type: 'updateSelectedEventTime',
        payload: event,
    };
};

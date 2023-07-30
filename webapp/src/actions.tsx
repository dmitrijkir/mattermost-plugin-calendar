import {DispatchFunc} from 'mattermost-redux/types/actions';

import {EventClickArg} from '@fullcalendar/common';

import {CalendarSettings} from './types/settings';
import {ApiClient} from './client';
import {CalendarEventNotification} from './types/event';

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

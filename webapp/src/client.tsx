import {Client4} from 'mattermost-redux/client';

import {UserProfile} from 'mattermost-redux/types/users';

import getSiteURL from 'components/utils';

import {id as PluginId} from './manifest';
import {CalendarSettings} from './types/settings';

export declare type ICalTokenResponse = {
    token?: string;
    url?: string;
    caldavUrl?: string;
    enabled: boolean;
}

export declare type GetEventResponse = {
    id: string;
    title: string;
    start: string;
    end: string;
    attendees: UserProfile[];
    created: string;
    owner: string;
    channel?: string;
    recurrence: string;
    color?: string
    description: string;
    team: string
    visibility: string
    alert: string
    type: string
    meetingLink?: string
    allDay?: boolean
}

export declare type GetEventsResponse = {
    id: string;
    title: string;
    start: string;
    end: string;
    created: string;
    owner: string;
    color?: string;
}
export declare type RemoveEventResponse = {
    success: boolean
}

export declare type UsersScheduleEvent = {
    start: string;
    end: string;
    duration: number;
}

export declare type UsersScheduleResponse = {
    users: Map<string, UsersScheduleEvent>
    available_times: string[]
}

export declare type ApiResponse<Type> = {
    data: Type
}

async function ensureOk(response: Response): Promise<void> {
    if (response.ok) {
        return;
    }
    let message = response.statusText || `Request failed with status ${response.status}`;
    try {
        const body = await response.clone().json();
        if (body?.message) {
            message = body.message;
        }
    } catch (e) {
        // response body isn't JSON, keep the default message
    }
    throw new Error(message);
}

export declare class ApiClientInterface {
    static getEventById(event: string): Promise<GetEventResponse>

    static getEvents(): Promise<GetEventsResponse>

    static createEvent(title: string, start: string, end: string, attendees: string[]): Promise<GetEventResponse>
}

export class ApiClient implements ApiClientInterface {
    static async getEventById(event: string): Promise<ApiResponse<GetEventResponse>> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/events/${event}`,
            Client4.getOptions({
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },

            }),
        );
        await ensureOk(response);
        const data = await response.json();
        // eslint-disable-next-line no-negated-condition
        if (data.data.attendees != null) {
            if (data.data.attendees.length > 0) {
                const users = await this.getUsersByIds(data.data.attendees);
                data.data.attendees = users;
            }
        } else {
            data.data.attendees = [];
        }

        return data;
    }

    static async getEvents(): Promise<GetEventsResponse> {
        throw new Error('Method not implemented.');
    }

    static async removeEvent(event: string): Promise<ApiResponse<RemoveEventResponse>> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/events/${event}`,
            Client4.getOptions({
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },

            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data;
    }

    static async getUsersByIds(users: string[]): Promise<UserProfile[]> {
        const response = await fetch(
            getSiteURL() + '/api/v4/users/ids',
            Client4.getOptions({
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(users),

            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data;
    }

    static async createEvent(
        title: string,
        start: string,
        end: string,
        attendees: string[],
        description: string,
        team: string,
        visibility: string,
        channel?: string,
        recurrence?: string,
        color?: string,
        alert?: string,
        type?: string,
        meetingLink?: string,
        allDay?: boolean,
    ): Promise<ApiResponse<GetEventResponse>> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/events`,
            Client4.getOptions({
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    title,
                    start,
                    end,
                    attendees,
                    description,
                    team,
                    visibility,
                    channel,
                    recurrence,
                    color,
                    alert,
                    type,
                    meetingLink,
                    allDay,
                }),
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data;
    }

    static async updateEvent(
        id: string,
        title: string,
        start: string,
        end: string,
        attendees: string[],
        description: string,
        team: string,
        visibility: string,
        channel?: string,
        recurrence?: string,
        color?: string,
        alert?: string,
        type?: string,
        meetingLink?: string,
        allDay?: boolean,
    ): Promise<ApiResponse<GetEventResponse>> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/events`,
            Client4.getOptions({
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id,
                    title,
                    start,
                    end,
                    attendees,
                    description,
                    team,
                    visibility,
                    channel,
                    recurrence,
                    color,
                    alert,
                    type,
                    meetingLink,
                    allDay,
                }),
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data;
    }

    static async getCalendarSettings(): Promise<CalendarSettings> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/settings`,
            Client4.getOptions({
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }

    static async getUsersSchedule(users: string[], start: string, end: string, slotTime: number): Promise<UsersScheduleResponse> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/schedule?` + new URLSearchParams({
                users: users.join(','),
                slot_time: slotTime.toString(),
                start,
                end,
            }),
            Client4.getOptions({
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }

    static async updateCalendarSettings(settings: CalendarSettings): Promise<CalendarSettings> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/settings`,
            Client4.getOptions({
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    isOpenCalendarLeftBar: settings.isOpenCalendarLeftBar,
                    firstDayOfWeek: settings.firstDayOfWeek,
                    hideNonWorkingDays: settings.hideNonWorkingDays,
                    callColor: settings.callColor,
                    eventColor: settings.eventColor,
                }),
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }

    static async getICalToken(): Promise<ICalTokenResponse> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/ical/token`,
            Client4.getOptions({
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }

    static async generateICalToken(): Promise<ICalTokenResponse> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/ical/token`,
            Client4.getOptions({
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }

    static async revokeICalToken(): Promise<ICalTokenResponse> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/ical/token`,
            Client4.getOptions({
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                },
            }),
        );
        await ensureOk(response);
        const data = await response.json();
        return data.data;
    }
}
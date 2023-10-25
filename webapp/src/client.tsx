import React from 'react';
import {Client4} from 'mattermost-redux/client';

import {UserProfile} from 'mattermost-redux/types/users';

import getSiteURL from 'components/utils';

import {id as PluginId} from './manifest';
import {CalendarSettings} from './types/settings';

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
        const data = await response.json();
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
        const data = await response.json();
        return data;
    }

    static async createEvent(
        title: string,
        start: string,
        end: string,
        attendees: string[],
        description: string,
        channel?: string,
        recurrence?: string,
        color?: string,
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
                    channel,
                    recurrence,
                    color,
                }),
            }),
        );
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
        channel?: string,
        recurrence?: string,
        color?: string,
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
                    channel,
                    recurrence,
                    color,
                }),
            }),
        );
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
        const data = await response.json();
        return data.data;
    }

    static async getUsersSchedule(users: string[], start: string, end: string): Promise<ApiResponse<UsersScheduleResponse>> {
        const response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/schedule?` + new URLSearchParams({
                users: users.join(','),
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
                }),
            }),
        );
        const data = await response.json();
        return data.data;
    }
}
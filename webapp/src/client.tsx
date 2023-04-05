import React, { useState } from 'react';
import { Client4 } from 'mattermost-redux/client';
import getSiteURL from 'components/utils';
import { UserProfile } from 'mattermost-redux/types/users';
import {id as PluginId} from './manifest';


export declare type GetEventResponse = {
    id: string;
    title: string;
    start: string;
    end: string;
    attendees: UserProfile[];
    created: string;
    owner: string;
    channel?: string;
    recurrence: number[];
    color?: string
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
        let response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/event?` + new URLSearchParams({
                eventId: event,
            }),
            Client4.getOptions({
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json'
                },

            })
        )
        let data = await response.json();
        if (data.data.attendees != null) {
            if (data.data.attendees.length > 0) {
                let users = await this.getUsersByIds(data.data.attendees);
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
        let response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/event?` + new URLSearchParams({
                eventId: event,
            }),
            Client4.getOptions({
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json'
                },

            })
        )
        let data = await response.json();
        return data;
    }
    static async getUsersByIds(users: string[]): Promise<UserProfile[]> {
        let response = await fetch(
            getSiteURL() + '/api/v4/users/ids',
            Client4.getOptions({
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(users)

            })
        )
        let data = await response.json();
        return data;
    }
    static async createEvent(title: string, start: string, end: string, attendees: string[], channel?: string, recurrence?: string, color?: string): Promise<ApiResponse<GetEventResponse>> {
        let response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/events`,
            Client4.getOptions({
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    "title": title,
                    "start": start,
                    "end": end,
                    "attendees": attendees,
                    "channel": channel,
                    "recurrence": recurrence,
                    "color": color,
                })
            })
        )
        let data = await response.json();
        return data;
    }


    static async updateEvent(id: string, title: string, start: string, end: string, attendees: string[], channel?: string, recurrence?: string, color?: string): Promise<ApiResponse<GetEventResponse>> {
        let response = await fetch(
            getSiteURL() + `/plugins/${PluginId}/event`,
            Client4.getOptions({
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    "id": id,
                    "title": title,
                    "start": start,
                    "end": end,
                    "attendees": attendees,
                    "channel": channel,
                    "recurrence": recurrence,
                    "color": color
                })
            })
        )
        let data = await response.json();
        return data;
    }

}
import {Action, Store} from 'redux';
import {GlobalState} from 'mattermost-redux/types/store';
import {getCurrentUserId} from 'mattermost-redux/selectors/entities/common';

import {render} from 'react-dom';

import React from 'react';

import {Provider} from 'react-redux';

import {FluentProvider, webLightTheme} from '@fluentui/react-components';

import reducer from 'reducers';

import MainApp from 'app';

import manifest from './manifest';
import {id as PluginId} from './manifest';

import './style.css';

import {PluginRegistry} from './types/mattermost-webapp';

import {eventNotification, updateCalendarSettings} from './actions';
import {ApiClient} from './client';
import NotificationWidget from './components/notification-widget';

const EmptyComponent = () => <></>;

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        registry.registerReducer(reducer);
        registry.registerProduct(
            '/calendar',
            'calendar-outline',
            'Calendar',
            '/calendar',
            MainApp,
            EmptyComponent,
            EmptyComponent,
            true,
        );

        // Load calendar settings like playbooks.
        // Mattermost initializes plugins before the user logs in (on the login page),
        // so wait for the current user: a request without a session only gets 401
        // and leaves "can't get session" in the server log.
        const getCalendarSettings = async () => {
            store.dispatch(updateCalendarSettings(await ApiClient.getCalendarSettings()));
        };
        if (getCurrentUserId(store.getState())) {
            getCalendarSettings();
        } else {
            // The flag guards against a second call: during a nested dispatch Redux may
            // still invoke the listener once more after unsubscribe().
            let requested = false;
            const unsubscribe = store.subscribe(() => {
                if (!requested && getCurrentUserId(store.getState())) {
                    requested = true;
                    unsubscribe();
                    getCalendarSettings();
                }
            });
        }

        // Register root DOM element for notification. This is where the widget will render.
        if (!document.getElementById('calendar-notifications')) {
            const notificationsRoot = document.createElement('div');
            notificationsRoot.setAttribute('id', 'calendar-notifications');
            document.body.appendChild(notificationsRoot);
        }

        render(
            <Provider store={store}>
                <FluentProvider
                    theme={webLightTheme}
                >
                    <NotificationWidget/>
                </FluentProvider>
            </Provider>,
            document.getElementById('calendar-notifications'),
        );

        registry.registerWebSocketEventHandler(`custom_${PluginId}_event_occur`, (ev) => {
            store.dispatch(eventNotification({id: ev.data.id, title: ev.data.title, channel: ev.data.channel}));
        });
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());

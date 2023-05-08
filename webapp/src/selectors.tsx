import {GlobalState} from 'mattermost-redux/types/store';

import {id as PluginId} from './manifest';
import {CalendarSettings} from './types/settings';

export const selectSelectedEvent = (state: GlobalState) => state[`plugins-${PluginId}`].selectEventModal;
export const selectIsOpenEventModal = (state: GlobalState) => state[`plugins-${PluginId}`].toggleEventModal.isOpen;
export const getCalendarSettings = (state: GlobalState) : CalendarSettings => state[`plugins-${PluginId}`].calendarSettings;

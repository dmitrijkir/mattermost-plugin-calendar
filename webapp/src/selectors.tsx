import {GlobalState} from 'mattermost-redux/types/store';

import {id as PluginId} from './manifest';
import {CalendarSettings} from './types/settings';

export const selectSelectedEvent = (state: GlobalState) => state[`plugins-${PluginId}`].selectEventModal;
export const selectIsOpenEventModal = (state: GlobalState) => state[`plugins-${PluginId}`].toggleEventModal.isOpen;
export const getCalendarSettings = (state: GlobalState) : CalendarSettings => state[`plugins-${PluginId}`].calendarSettings;
export const selectIsSettingsPanelOpen = (state: GlobalState): boolean => state[`plugins-${PluginId}`].settingsPanel.isOpen;
export const getEventNotification = (state: GlobalState) => state[`plugins-${PluginId}`].eventNotification;
export const getMembersAddedInEvent = (state: GlobalState) => state[`plugins-${PluginId}`].membersAddedInEvent;
export const getSelectedEventTime = (state: GlobalState) => state[`plugins-${PluginId}`].selectedEventTime;
export const getSelectedCalendarType = (state: GlobalState): string => state[`plugins-${PluginId}`].selectedCalendarType;
export const getSelectedCalendarView = (state: GlobalState): string => state[`plugins-${PluginId}`].selectedCalendarView;

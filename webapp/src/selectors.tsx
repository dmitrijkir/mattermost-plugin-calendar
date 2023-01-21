import { GlobalState } from 'mattermost-redux/types/store';
import {id as PluginId} from './manifest';


export const selectSelectedEvent = (state: GlobalState) => state[`plugins-${PluginId}`].selectEventModal;
export const selectIsOpenEventModal = (state: GlobalState) => state[`plugins-${PluginId}`].toggleEventModal.isOpen;

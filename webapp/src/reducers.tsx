import {combineReducers} from 'redux';

import {getDefaultCalendarView} from './components/utils';

const selectEventModal = (state = {}, action) => {
    switch (action.type) {
    case 'eventSelected':
        return action.payload;
    default:
        return state;
    }
};

const toggleEventModal = (state = {isOpen: false}, action) => {
    switch (action.type) {
    case 'openEventModal':
        return {
            ...state,
            isOpen: true,
        };
    case 'closeEventModal':
        return {
            ...state,
            isOpen: false,
        };
    default:
        return state;
    }
};

const calendarSettings = (state = {
    isOpenCalendarLeftBar: true,
    firstDayOfWeek: 1,
    businessStartTime: '08:00',
    businessEndTime: '18:00',
    businessDays: [1, 2, 3, 4, 5],
    hideNonWorkingDays: false,
    callColor: '#B3E1F7',
    eventColor: '#B6D9C7',
    jitsiBaseUrl: 'https://meet.jit.si',
}, action) => {
    switch (action.type) {

    // merged rather than replaced so a server that doesn't know about a newer
    // setting can't drop it and leave the UI reading undefined
    case 'updateCalendarSettings':
        return {...state, ...action.payload};
    default:
        return state;
    }
};

const selectedCalendarType = (state = 'call', action) => {
    switch (action.type) {
    case 'updateSelectedCalendarType':
        return action.payload;
    default:
        return state;
    }
};

// the header buttons and the grid live in sibling components, so the active
// view is kept here instead of in local state
const selectedCalendarView = (state = getDefaultCalendarView(), action) => {
    switch (action.type) {
    case 'updateSelectedCalendarView':
        return action.payload;
    default:
        return state;
    }
};

const eventNotification = (state = {}, action) => {
    switch (action.type) {
    case 'eventNotification':
        return action.payload;
    default:
        return state;
    }
};

const membersAddedInEvent = (state = [], action) => {
    switch (action.type) {
    case 'updateMembersAddedInEvent':
        return action.payload;
    default:
        return state;
    }
};

const selectedEventTime = (state = {start: new Date(), end: new Date(), startTime: '00:00', endTime: '00:00'}, action) => {
    switch (action.type) {
    case 'updateSelectedEventTime':
        return {...state, ...action.payload};
    default:
        return state;
    }
};

const reducer = combineReducers({
    selectEventModal,
    toggleEventModal,
    calendarSettings,
    eventNotification,
    membersAddedInEvent,
    selectedEventTime,
    selectedCalendarType,
    selectedCalendarView,
});

export default reducer;

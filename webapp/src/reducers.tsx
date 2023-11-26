import {combineReducers} from 'redux';

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
}, action) => {
    switch (action.type) {
    case 'updateCalendarSettings':
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
});

export default reducer;
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

const eventNotification = (state = {
    id: "243252fwefwef",
    title: 'test event',
    start: new Date(),
    end: new Date(),
}, action) => {
    switch (action.type) {
        case 'showEventNotification':
            return action.payload;
        default:
            return state;
    }
}

const reducer = combineReducers({
    selectEventModal,
    toggleEventModal,
    calendarSettings,
    eventNotification,
});

export default reducer;
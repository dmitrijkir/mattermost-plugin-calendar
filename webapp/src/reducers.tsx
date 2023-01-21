import {combineReducers} from 'redux';


const selectEventModal = (state = {}, action) => {
    switch (action.type) {
        case "eventSelected":
            return action.payload;
        default:
            return state;
    }
}

const toggleEventModal = (state = {isOpen: false}, action) => {
    switch (action.type) {
        case "openEventModal":
            return {
                ...state,
                isOpen: true
            }
        case "closeEventModal":
            return {
                ...state,
                isOpen: false
            }
        default:
            return state;
    }
}
const reducer = combineReducers({
    selectEventModal,
    toggleEventModal,
});

export default reducer;
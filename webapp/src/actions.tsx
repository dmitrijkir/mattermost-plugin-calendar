export const eventSelected = (event) => {
    return {
        type: 'eventSelected',
        payload: event,
    };
};

export const openEventModal = () => {
    return {
        type: 'openEventModal',
        payload: true,
    };
};

export const closeEventModal = () => {
    return {
        type: 'closeEventModal',
        payload: false,
    };
};
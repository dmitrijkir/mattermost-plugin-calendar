import React from 'react';
import FullCalendar from '@fullcalendar/react';

const CalendarRef = React.createRef<FullCalendar>();

// Refetching used to be written as getEventSources()[0].refetch(), which throws
// whenever the grid isn't mounted yet or the source has just been swapped out
// (switching calendars removes the old source and adds a new one).
export const refetchCalendarEvents = () => {
    const api = CalendarRef.current?.getApi();
    if (!api) {
        return;
    }
    api.getEventSources().forEach((source) => source.refetch());
};

export default CalendarRef;

import FullCalendar from '@fullcalendar/react';
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import React, {useEffect} from 'react';

import interactionPlugin from '@fullcalendar/interaction';
import {useDispatch, useSelector} from 'react-redux';

import {DayHeaderContentArg} from '@fullcalendar/core';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {id as PluginId} from '../manifest';
import {eventSelected, openEventModal} from 'actions';

import getSiteURL from './utils';
import CalendarRef from './calendar';

const eventDataTransformation = (content, response) => {
    return content.data;
};

const onDateTimeSelected = (dateTimeSelectInfo, dispatch) => {
    dispatch(eventSelected({
        event: {
            start: dateTimeSelectInfo.start.toISOString(),
            end: dateTimeSelectInfo.end.toISOString(),
        },
    }));
    dispatch(openEventModal());
};

const CalendarContent = () => {
    const dispatch = useDispatch();
    const user = useSelector(getCurrentUser);

    const getUserTimeZoneString = () => {
        if (user.timezone?.useAutomaticTimezone) {
            return user.timezone.automaticTimezone;
        }
        return user.timezone?.manualTimezone;
    };

    useEffect(() => {
    }, [user]);

    const onEventClicked = (eventInfo) => {
        dispatch(eventSelected(eventInfo));
        dispatch(openEventModal());
    };

    return (
        <div>
            <div className='calendar-main-greed'>
                <FullCalendar
                    plugins={[timeGridPlugin, interactionPlugin]}
                    initialView='timeGridWeek'
                    allDaySlot={false}
                    slotDuration='00:30:00'
                    selectable={true}

                    timeZone={getUserTimeZoneString()}
                    handleWindowResize={true}
                    headerToolbar={{
                        start: 'today,prev,next',
                        center: 'title',
                        end: '',
                    }}
                    nowIndicatorClassNames='now-indicator'
                    select={(info) => onDateTimeSelected(info, dispatch)}
                    dayHeaderFormat={{day: 'numeric', weekday: 'short', omitCommas: true}}
                    nowIndicator={true}
                    locales={[enLocale]}
                    contentHeight={window.innerHeight - 200}
                    eventClick={onEventClicked}
                    dayHeaderContent={(dayHeaderProps: DayHeaderContentArg) => {
                        function dayOfWeekAsString(dayIndex: number) {
                            return ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][dayIndex] || '';
                        }

                        return (<>
                            <div className={`custom-day-header  ${dayHeaderProps.isToday ? 'custom-day-today' : ''}`}>
                                <div className='custom-day-header-day'>{dayHeaderProps.date.getDate()}</div>
                                <div
                                    className='custom-day-header-weekday'
                                >{dayOfWeekAsString(dayHeaderProps.date.getDay())}</div>
                            </div>
                        </>);
                    }}
                    dayCellClassNames='custom-day-cell'
                    ref={CalendarRef}
                    eventSourceSuccess={eventDataTransformation}
                    eventSources={[
                        {
                            url: getSiteURL() + `/plugins/${PluginId}/events`,
                        },
                    ]}
                />
            </div>
        </div>
    );
};

export default CalendarContent;

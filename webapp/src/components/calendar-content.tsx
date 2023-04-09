import FullCalendar from '@fullcalendar/react';
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import React, {useEffect, useState} from 'react';

import interactionPlugin from '@fullcalendar/interaction';
import {useDispatch, useSelector} from 'react-redux';

import {DayHeaderContentArg} from '@fullcalendar/core';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {id as PluginId} from '../manifest';
import {eventSelected, openEventModal} from 'actions';
import {DateSelectArg, EventClickArg} from '@fullcalendar/common';
import {Button} from "@fluentui/react-components";
import getSiteURL from './utils';
import CalendarRef from './calendar';
import {DayPicker} from "react-day-picker";
import {en} from "date-fns/locale";
import {addMonths, isSameMonth} from 'date-fns';


const eventDataTransformation = (content, response) => {
    return content.data;
};

const LeftBarCalendar = () => {
    const today = new Date();
    const nextMonth = addMonths(new Date(), 1);
    const [month, setMonth] = useState<Date>(nextMonth);

    const footer = (
        <Button
            disabled={isSameMonth(today, month)}
            onClick={() => setMonth(today)}
        >
            Go to Today
        </Button>
    );
    return (<DayPicker
        mode='single'

        onDayClick={(day: Date) => {
            CalendarRef.current.getApi().gotoDate(day);
        }}
        locale={en}
        weekStartsOn={1}
        month={month}
        onMonthChange={setMonth}
        footer={footer}
    />);
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
        console.log(user);
    }, [user]);

    const onEventClicked = (eventInfo: EventClickArg) => {
        dispatch(eventSelected(eventInfo));
        dispatch(openEventModal());
    };

    const onDateTimeSelected = (dateTimeSelectInfo: DateSelectArg) => {
        dispatch(eventSelected({
            event: {
                start: dateTimeSelectInfo.start.toISOString(),
                end: dateTimeSelectInfo.end.toISOString(),
            },
        }));
        dispatch(openEventModal());
    };

    return (
        <div className='calendar-content'>
            <div className='left-bar-calendar-content'>
                <LeftBarCalendar/>
            </div>
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
                    select={(info: DateSelectArg) => onDateTimeSelected(info)}
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

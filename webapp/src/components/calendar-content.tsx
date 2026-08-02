import FullCalendar from '@fullcalendar/react';
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import dayGridPlugin from '@fullcalendar/daygrid';

import React, {useEffect, useMemo, useState} from 'react';

import interactionPlugin from '@fullcalendar/interaction';
import {useDispatch, useSelector} from 'react-redux';

import {DayHeaderContentArg} from '@fullcalendar/core';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {DateSelectArg, EventClickArg} from '@fullcalendar/common';
import {Calendar, DateRangeType, DayOfWeek, initializeIcons} from '@fluentui/react';

import {addMonths, format} from 'date-fns';

import {eventSelected, openEventModal} from 'actions';
import {id as PluginId} from '../manifest';
import {CalendarSettings} from '../types/settings';
import {getCalendarSettings} from '../selectors';

import CalendarRef from './calendar';
import getSiteURL from './utils';

initializeIcons();

const eventDataTransformation = (content, response) => {
    return content.data;
};

const DAY_HEADER_FORMAT = {day: 'numeric', weekday: 'short', omitCommas: true} as const;
const LOCALES = [enLocale];
const PLUGINS = [timeGridPlugin, interactionPlugin, dayGridPlugin];

const LeftBarCalendar = () => {
    const [selectedDate, setSelectedDate] = useState<Date>();
    const dateRangeType = DateRangeType.Week;

    const settings: CalendarSettings = useSelector(getCalendarSettings);

    const onSelectDate = React.useCallback((date: Date, dateRangeArray: Date[]): void => {
        setSelectedDate(date);
        // Format as YYYY-MM-DD to avoid timezone issues with gotoDate
        const dateString = format(date, 'yyyy-MM-dd');
        CalendarRef.current?.getApi().gotoDate(dateString);
    }, []);

    if (settings.isOpenCalendarLeftBar) {
        return (
            <Calendar
                showMonthPickerAsOverlay={true}
                dateRangeType={dateRangeType}
                highlightSelectedMonth={true}
                showGoToToday={true}
                onSelectDate={onSelectDate}
                value={selectedDate}
                firstDayOfWeek={settings.firstDayOfWeek}
            />
        );
    }

    return <div className='hided-left-bar-calendar'/>;
};

const CalendarContent = () => {
    const dispatch = useDispatch();
    const user = useSelector(getCurrentUser);
    const settings = useSelector(getCalendarSettings);

    const [contentHeight, setContentHeight] = useState<number>(window.innerHeight - 200);

    const getUserTimeZoneString = () => {
        if (!user) {
            return 'local';
        }
        if (user.timezone?.useAutomaticTimezone) {
            return user.timezone.automaticTimezone;
        }
        return user.timezone?.manualTimezone || 'local';
    };

    useEffect(() => {
        if (!user) {
            return;
        }
        const now: Date = new Date();
        const scrollTo: Date = new Date();
        scrollTo.setMinutes(scrollTo.getMinutes() - 30);
        if (now.getDate() === scrollTo.getDate()) {
            CalendarRef.current?.getApi().scrollToTime(format(scrollTo, 'HH:mm'));
        }
    }, [user]);

    useEffect(() => {
        const onResize = () => setContentHeight(window.innerHeight - 200);
        window.addEventListener('resize', onResize);
        return () => window.removeEventListener('resize', onResize);
    }, []);

    const onEventClicked = (eventInfo: EventClickArg) => {
        dispatch(eventSelected(eventInfo));
        dispatch(openEventModal());
    };

    const calcHiddenDays = (): number[] => {
        if (!settings.hideNonWorkingDays) {
            return [];
        }
        const noneWorkingDays: number[] = [];
        const allDays = [0, 1, 2, 3, 4, 5, 6];
        allDays.forEach((item) => {
            if (!settings.businessDays.includes(item)) {
                noneWorkingDays.push(item);
            }
        });
        return noneWorkingDays;
    };

    const onDateTimeSelected = (dateTimeSelectInfo: DateSelectArg) => {
        dispatch(eventSelected({
            event: {
                start: dateTimeSelectInfo.start.setMinutes(dateTimeSelectInfo.start.getMinutes() + dateTimeSelectInfo.start.getTimezoneOffset()),
                end: dateTimeSelectInfo.end.setMinutes(dateTimeSelectInfo.end.getMinutes() + dateTimeSelectInfo.end.getTimezoneOffset()),
            },
        }));
        dispatch(openEventModal());
    };

    const businessHours = useMemo(() => ({
        startTime: settings.businessStartTime,
        endTime: settings.businessEndTime,
        daysOfWeek: settings.businessDays,
    }), [settings.businessStartTime, settings.businessEndTime, settings.businessDays]);

    const hiddenDays = useMemo(calcHiddenDays, [settings.hideNonWorkingDays, settings.businessDays]);

    const eventSources = useMemo(() => [
        {
            url: getSiteURL() + `/plugins/${PluginId}/events`,
        },
    ], []);

    if (!user) {
        return (
            <div className='calendar-content'>
                <div className='calendar-main-greed'/>
            </div>
        );
    }

    return (
        <div className='calendar-content'>
            <div className='left-bar-calendar-content'>
                <LeftBarCalendar/>
            </div>
            <div className='calendar-main-greed'>
                <FullCalendar
                    plugins={PLUGINS}
                    initialView='timeGridWeek'
                    allDaySlot={false}
                    slotDuration='00:30:00'
                    selectable={true}
                    firstDay={settings.firstDayOfWeek}
                    businessHours={businessHours}
                    timeZone={getUserTimeZoneString()}
                    handleWindowResize={true}
                    headerToolbar={{
                        start: 'today,prev,next',
                        center: 'title',
                        end: '',
                    }}
                    hiddenDays={hiddenDays}
                    nowIndicatorClassNames='now-indicator'
                    select={(info: DateSelectArg) => onDateTimeSelected(info)}
                    dayHeaderFormat={DAY_HEADER_FORMAT}
                    nowIndicator={true}
                    locales={LOCALES}
                    contentHeight={contentHeight}
                    eventClick={onEventClicked}
                    dayHeaderContent={(dayHeaderProps: DayHeaderContentArg) => {
                        function dayOfWeekAsString(dayIndex: number) {
                            return ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'][dayIndex] || '';
                        }
                        const showDay = CalendarRef.current?.getApi().view.type !== 'dayGridMonth';
                        return (<>
                            <div className={`custom-day-header  ${dayHeaderProps.isToday ? 'custom-day-today' : ''}`}>
                                {showDay ? <div className='custom-day-header-day'>{dayHeaderProps.date.getDate()}</div> : ''}
                                <div
                                    className='custom-day-header-weekday'
                                >{dayOfWeekAsString(dayHeaderProps.date.getDay())}</div>
                            </div>
                        </>);
                    }}
                    dayCellClassNames='custom-day-cell'
                    ref={CalendarRef}
                    eventSourceSuccess={eventDataTransformation}
                    eventSources={eventSources}
                />
            </div>
        </div>
    );
};

export default CalendarContent;

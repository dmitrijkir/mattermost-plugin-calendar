import FullCalendar from '@fullcalendar/react';
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import dayGridPlugin from '@fullcalendar/daygrid';

import React, {useEffect, useMemo, useState} from 'react';

import interactionPlugin from '@fullcalendar/interaction';
import {useDispatch, useSelector} from 'react-redux';

import {DayHeaderContentArg} from '@fullcalendar/core';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/users';

import {DateSelectArg, DatesSetArg, EventClickArg} from '@fullcalendar/common';
import {initializeIcons} from '@fluentui/react';

import {format} from 'date-fns';

import {eventSelected, openEventModal, setSettingsPanelOpen, updateSelectedCalendarView} from 'actions';
import {id as PluginId} from '../manifest';
import {getCalendarSettings, getSelectedCalendarType, getSelectedCalendarView} from '../selectors';

import CalendarRef from './calendar';
import getSiteURL, {isMobileViewport} from './utils';

initializeIcons();

// FullCalendar iterates whatever comes out of here. Anything that isn't an
// array (an empty calendar serialises as `data: null`, an error body has no
// `data` at all) throws inside its task queue, and that exception leaves the
// queue permanently stuck: every later action — refetch, prev/next, changeView —
// is then silently dropped and the whole calendar looks dead.
const eventDataTransformation = (content) => {
    return Array.isArray(content?.data) ? content.data : [];
};

const onEventSourceFailure = (error) => {
    // eslint-disable-next-line no-console
    console.error('calendar plugin: could not load events', error);
};

const DAY_HEADER_FORMAT = {day: 'numeric', weekday: 'short', omitCommas: true} as const;
const LOCALES = [enLocale];
const PLUGINS = [timeGridPlugin, interactionPlugin, dayGridPlugin];

// the settings button sits in the calendar's own toolbar next to the month
// title. FullCalendar renders custom buttons itself and only takes a text
// label, so the gear is painted on in CSS (.fc-calendarSettings-button).
const HEADER_TOOLBAR = {
    start: 'today,prev,next',
    center: 'title',
    end: 'calendarSettings',
} as const;

// what's left of the viewport once the button bar, the navigation row and the
// gaps around them are accounted for. Kept in step with the paddings in
// style.css — tightening those there means adding the difference back here,
// otherwise the grid just leaves the space empty.
const contentHeightForViewport = (): number => {
    return Math.max(320, window.innerHeight - (isMobileViewport() ? 150 : 178));
};

const CalendarContent = () => {
    const dispatch = useDispatch();
    const user = useSelector(getCurrentUser);
    const settings = useSelector(getCalendarSettings);
    const selectedCalendarType: string = useSelector(getSelectedCalendarType);
    const selectedCalendarView: string = useSelector(getSelectedCalendarView);

    // FullCalendar only reads initialView once, so it has to stay stable even
    // when the header switches views afterwards
    const [initialView] = useState<string>(selectedCalendarView);
    const [contentHeight, setContentHeight] = useState<number>(contentHeightForViewport);

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
        const onResize = () => setContentHeight(contentHeightForViewport());
        window.addEventListener('resize', onResize);
        window.addEventListener('orientationchange', onResize);
        return () => {
            window.removeEventListener('resize', onResize);
            window.removeEventListener('orientationchange', onResize);
        };
    }, []);

    // keeps the header buttons in step with views changed from the grid itself
    const onDatesSet = (arg: DatesSetArg) => {
        if (arg.view.type !== selectedCalendarView) {
            dispatch(updateSelectedCalendarView(arg.view.type));
        }
    };

    const onEventClicked = (eventInfo: EventClickArg) => {
        dispatch(eventSelected(eventInfo));
        dispatch(openEventModal());
    };

    const calcHiddenDays = (): number[] => {
        if (!settings.hideNonWorkingDays || !Array.isArray(settings.businessDays)) {
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

    // `icon`, not `text`: FullCalendar renders a text label as a bare text node
    // inside the button and only falls back to an icon span when `icon` is set.
    // Hiding that label in CSS was a coin flip on which stylesheet was injected
    // last, so the word "Settings" showed up at random. `hint` only sets the
    // title attribute, not anything in the layout.
    const customButtons = useMemo(() => ({
        calendarSettings: {
            icon: 'settings-gear',
            hint: 'Settings',
            click: () => dispatch(setSettingsPanelOpen(true)),
        },
    }), []);

    const eventSources = useMemo(() => [
        {
            url: getSiteURL() + `/plugins/${PluginId}/events`,
            extraParams: {
                // "all" shows both calendars together, so no type filter is sent
                type: selectedCalendarType === 'all' ? '' : selectedCalendarType,
            },
        },
    ], [selectedCalendarType]);

    if (!user) {
        return (
            <div className='calendar-content'>
                <div className='calendar-main-greed'/>
            </div>
        );
    }

    return (
        <div className='calendar-content'>
            <div className='calendar-main-greed'>
                <FullCalendar
                    plugins={PLUGINS}
                    initialView={initialView}
                    allDaySlot={true}
                    slotDuration='00:30:00'
                    selectable={true}
                    firstDay={settings.firstDayOfWeek}
                    businessHours={businessHours}
                    timeZone={getUserTimeZoneString()}
                    handleWindowResize={true}
                    headerToolbar={HEADER_TOOLBAR}
                    customButtons={customButtons}
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
                        const showDay = dayHeaderProps.view.type !== 'dayGridMonth';
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
                    datesSet={onDatesSet}
                    eventSourceSuccess={eventDataTransformation}
                    eventSourceFailure={onEventSourceFailure}
                    eventSources={eventSources}
                />
            </div>
        </div>
    );
};

export default CalendarContent;

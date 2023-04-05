import FullCalendar from '@fullcalendar/react'
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import React, {useEffect, useState} from 'react';
import CalendarRef from './calendar';
import getSiteURL from './utils';
import interactionPlugin from '@fullcalendar/interaction';
import {useDispatch, useSelector} from 'react-redux';
import {eventSelected, openEventModal} from 'actions';
import {Client4} from 'mattermost-redux/client';
import {id as PluginId} from '../manifest';
import {getTheme} from "mattermost-redux/selectors/entities/preferences";


const eventDataTransformation = (content, response) => {
    return content.data
}

const onDateTimeSelected = (dateTimeSelectInfo, dispatch) => {
    console.log(dateTimeSelectInfo);
    dispatch(eventSelected({
        event: {
            start: dateTimeSelectInfo.start.toISOString(),
            end: dateTimeSelectInfo.end.toISOString()
        }
    }));
    dispatch(openEventModal())

}

const CalendarContent = () => {
    const theme = useSelector(getTheme);

    const dispatch = useDispatch()
    const [userTimezone, setUserTimeZone] = useState("");

    const onEventClicked = (eventInfo) => {
        dispatch(eventSelected(eventInfo))
        dispatch(openEventModal())
    }


    const CalendarComponent = () => {

        useEffect(() => {
            let mounted = true;
            if (mounted) {
                Client4.getMe().then((user) => {
                    if (user.timezone != null) {
                        if (user.timezone.useAutomaticTimezone === 'true') {
                            setUserTimeZone(user.timezone.automaticTimezone)
                        } else {
                            setUserTimeZone(user.timezone.manualTimezone);
                        }
                    }
                })
            }

            mounted = false;
            return;
        }, [userTimezone])
        return <FullCalendar
            plugins={[timeGridPlugin, interactionPlugin]}
            initialView='timeGridWeek'
            allDaySlot={false}
            slotDuration="00:30:00"
            selectable={true}
            timeZone={userTimezone}
            handleWindowResize={true}
            headerToolbar={{
                start: 'today,prev,next',
                center: 'title',
                end: ''
            }}
            nowIndicatorClassNames="now-indicator"
            // now={() => {
            //     return new Date()
            // }}

            select={(info) => onDateTimeSelected(info, dispatch)}
            // duration={{ days: 7 }}
            // views={{
            //     timeGridWeek: {

            //     }
            // }}
            dayHeaderFormat={{day: 'numeric', weekday: 'short', omitCommas: true}}
            // weekends={true}
            // weekNumberCalculation="ISO"
            // firstDay={1}
            nowIndicator={true}
            locales={[enLocale]}
            contentHeight={window.innerHeight - 200}
            eventClick={onEventClicked}

            eventBackgroundColor={theme.sidebarHeaderBg}
            eventBorderColor={theme.sidebarHeaderBg}
            eventTextColor={theme.sidebarHeaderTextColor}

            dayHeaderContent={(props) => {
                function dayOfWeekAsString(dayIndex: number) {
                    return ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"][dayIndex] || '';
                }

                return <>
                    <div className={`custom-day-header  ${props.isToday ? "custom-day-today" : ""}`}>
                        <div className='custom-day-header-day'>{props.date.getDate()}</div>
                        <div className='custom-day-header-weekday'>{dayOfWeekAsString(props.date.getDay())}</div>
                    </div>
                </>
            }}
            dayCellClassNames="custom-day-cell"
            ref={CalendarRef}
            // headerToolbar={{
            //     left: 'prev,next today',
            //     center: 'title',
            //     right: 'timeGridWeek'
            // }}
            eventSourceSuccess={eventDataTransformation}
            eventSources={[
                {
                    url: getSiteURL() + `/plugins/${PluginId}/events`,
                }
            ]}
        />
    }

    return (
        <div>
            <div className='calendar-main-greed'>
                <CalendarComponent/>
            </div>
        </div>
    );
};

export default CalendarContent;

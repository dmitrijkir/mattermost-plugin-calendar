import FullCalendar from '@fullcalendar/react'
import enLocale from '@fullcalendar/core/locales/en-gb';
import timeGridPlugin from '@fullcalendar/timegrid';
import React, { useEffect, useState } from 'react';
import CalendarRef from './calendar';
import getSiteURL from './utils';
import interactionPlugin from '@fullcalendar/interaction';
import { useDispatch } from 'react-redux';
import { eventSelected, openEventModal } from 'actions';
import { Client4 } from 'mattermost-redux/client';
import {id as PluginId} from '../manifest';


function eventDataTransformation(content, response) {
    return content.data
}

const CalendarContent = () => {

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
            // now={() => {
            //     return new Date()
            // }}
            select={(dateClickInfo) => console.log(dateClickInfo)}
            // duration={{ days: 7 }}
            // views={{
            //     timeGridWeek: {
            //         duration: { days: 7 },
            //         firstDay: 1,
            //     }
            // }}
            // weekends={true}
            // weekNumberCalculation="ISO"
            // firstDay={1}
            nowIndicator={true}
            locales={[enLocale]}
            contentHeight={window.innerHeight - 200}
            eventClick={onEventClicked}
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
                <CalendarComponent />
            </div>
        </div>
    );
};

export default CalendarContent;
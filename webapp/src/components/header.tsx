import React, {useState} from 'react';
import {useDispatch} from 'react-redux';
import {openEventModal} from 'actions';
import {Button} from "@fluentui/react-components";
import {
    Calendar3Day20Regular,
    CalendarDay20Regular,
    CalendarEmpty16Filled,
    CalendarLtr20Regular,
    Settings20Regular,
} from "@fluentui/react-icons";
import CalendarRef from "./calendar";


const HeaderComponent = () => {
    const dispatch = useDispatch();

    const [selectedView, setSelectedView] = useState<string>('timeGridWeek');

    return (
        <div className='calendar-header-container'>
            <div className='calendar-header-toolbar'>
                <div className='left-allign-header-toolbar-item'>
                    <Button
                        appearance='primary'
                        size='medium'
                        onClick={() => dispatch(openEventModal())}
                        icon={<CalendarEmpty16Filled/>}
                    >
                        <div className='create-event-button-text'>New event</div>
                    </Button>
                    <Button
                        appearance='subtle'
                        icon={<CalendarDay20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('dayGridDay');
                            setSelectedView('dayGridDay');
                        }}
                        disabled={selectedView === 'dayGridDay'}
                    >Day</Button>
                    <Button
                        appearance='subtle'
                        icon={<Calendar3Day20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('timeGridWeek');
                            setSelectedView('timeGridWeek');
                        }}
                        disabled={selectedView === 'timeGridWeek'}
                    >week</Button>
                    <Button
                        appearance='subtle'
                        icon={<CalendarLtr20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('dayGridMonth');
                            setSelectedView('dayGridMonth');
                        }}
                        disabled={selectedView === 'dayGridMonth'}
                    >month</Button>
                </div>
                <div className='left-allign-header-toolbar-item'>
                    <Button
                        appearance='subtle'
                        icon={<Settings20Regular/>}
                    />
                </div>
            </div>

        </div>
    );
};

export default HeaderComponent;
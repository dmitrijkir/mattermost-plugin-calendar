import React from 'react';
import CalendarContent from 'components/calendar-content';
import HeaderComponent from 'components/header';
import EventModalComponent from 'components/event';

const MainApp = () => {
    return (
        <div>
            <EventModalComponent />
            <HeaderComponent />
            <CalendarContent />
        </div>
    );
};

export default MainApp;
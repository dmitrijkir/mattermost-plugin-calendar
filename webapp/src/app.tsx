import React from 'react';

import {FluentProvider, teamsLightTheme} from '@fluentui/react-components';

import EventModalComponent from 'components/event';
import HeaderComponent from 'components/header';
import CalendarContent from 'components/calendar-content';

const MainApp = () => {
    return (
        <div className='calendar-full-content-provider'>
            <FluentProvider
                theme={teamsLightTheme}
            >
                <span className='calendar-full-content'>
                    <EventModalComponent/>
                    <HeaderComponent/>
                    <CalendarContent/>
                </span>
            </FluentProvider>
        </div>

    );
};

export default MainApp;
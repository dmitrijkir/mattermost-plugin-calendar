import React from 'react';

import {FluentProvider, webLightTheme} from '@fluentui/react-components';

import EventModalComponent from 'components/event';
import HeaderComponent from 'components/header';
import CalendarContent from 'components/calendar-content';

const MainApp = () => {
    return (
        <FluentProvider
            theme={webLightTheme}
            className='calendar-full-content-provider'
        >
            <span className='calendar-full-content'>
                <EventModalComponent/>
                <HeaderComponent/>
                <CalendarContent/>
            </span>
        </FluentProvider>
    );
};

export default MainApp;
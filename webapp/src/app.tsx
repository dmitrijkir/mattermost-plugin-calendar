import React from 'react';

import {FluentProvider, webLightTheme} from '@fluentui/react-components';

import EventModalComponent from 'components/event';
import HeaderComponent from 'components/header';
import CalendarContent from 'components/calendar-content';

const MainApp = () => {
    return (
        <div className='calendar-full-content-provider'>
            <FluentProvider theme={webLightTheme}>
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
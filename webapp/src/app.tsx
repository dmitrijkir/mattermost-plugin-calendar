import React from 'react';

import {FluentProvider, webDarkTheme, webLightTheme} from '@fluentui/react-components';
import {useSelector} from 'react-redux';
import {getTheme} from 'mattermost-redux/selectors/entities/preferences';
import {Theme} from 'mattermost-redux/types/preferences';

import EventModalComponent from 'components/event';
import HeaderComponent from 'components/header';
import CalendarContent from 'components/calendar-content';

const MainApp = () => {
    const theme: Theme = useSelector(getTheme);
    return (
        <FluentProvider
            theme={['indigo', 'onyx'].includes(theme.type!) ? webDarkTheme : webLightTheme}
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
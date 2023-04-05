import React, {useEffect, useState} from 'react';
import CalendarContent from 'components/calendar-content';
import HeaderComponent from 'components/header';
import EventModalComponent from 'components/event';
import {FluentProvider, webLightTheme, webDarkTheme} from '@fluentui/react-components';
import {useSelector} from "react-redux";
import {getTheme} from "mattermost-redux/selectors/entities/preferences";

const MainApp = () => {
    const theme = useSelector(getTheme);
    const [selectedTheme, setSelectedTheme] = useState(webLightTheme);

    useEffect(() => {
        if (['quartz', 'denim', 'sapphire'].includes(theme.type)) {
            setSelectedTheme(webLightTheme);
        } else {
            setSelectedTheme(webDarkTheme);
        }
    }, [theme]);

    return (
        <FluentProvider
            theme={selectedTheme}
            className='calendar-full-content'
        >
            <EventModalComponent/>
            <HeaderComponent/>
            <CalendarContent/>
        </FluentProvider>
    );
};

export default MainApp;
import React from 'react';
// import Button from 'react-bootstrap/Button';
import { useDispatch, useSelector } from 'react-redux';
import { openEventModal } from 'actions';
import { getTheme } from "mattermost-redux/selectors/entities/preferences";
import {
    Dialog,
    DialogTrigger,
    DialogSurface,
    DialogTitle,
    DialogBody,
    DialogActions,
    DialogContent,
    Button,
    Select,

} from "@fluentui/react-components";

import {
    Add16Filled,
    CalendarEmpty16Filled
} from "@fluentui/react-icons";

const HeaderComponent = () => {

    const theme = useSelector(getTheme);

    const dispatch = useDispatch();

    return (
        <div className='calendar-header-container'>
            {/* <Button appearance="primary" onClick={() => dispatch(openEventModal())}>
                <div className='create-event-button-inner-container' style={{color: theme.buttonColor}}>
                    <i className='icon fa fa-plus' />
                    <div className='create-event-button-text'>Create event</div>
                </div>
            </Button> */}
            <Button appearance="primary" onClick={() => dispatch(openEventModal())} icon={<CalendarEmpty16Filled />}>
                <div className='create-event-button-text'>New event</div>
            </Button>
        </div>
    );
};

export default HeaderComponent;
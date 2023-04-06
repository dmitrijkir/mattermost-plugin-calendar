import React from 'react';
import {useDispatch} from 'react-redux';
import {openEventModal} from 'actions';
import {Button} from "@fluentui/react-components";
import {CalendarEmpty16Filled} from "@fluentui/react-icons";

const HeaderComponent = () => {
    const dispatch = useDispatch();

    return (
        <div className='calendar-header-container'>
            <Button appearance='primary' onClick={() => dispatch(openEventModal())} icon={<CalendarEmpty16Filled/>}>
                <div className='create-event-button-text'>New event</div>
            </Button>
        </div>
    );
};

export default HeaderComponent;
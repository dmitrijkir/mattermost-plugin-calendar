import React from 'react';
import Button from 'react-bootstrap/Button';
import { useDispatch, useSelector } from 'react-redux';
import { openEventModal } from 'actions';
import {getTheme} from  "mattermost-redux/selectors/entities/preferences";



const HeaderComponent = () => {

    const theme = useSelector(getTheme);

    const dispatch = useDispatch();

    return (
        <div className='calendar-header-container'>
            <Button variant="primary" onClick={() => dispatch(openEventModal())} style={{backgroundColor: theme.buttonBg}}>
                <div className='create-event-button-inner-container' style={{color: theme.buttonColor}}>
                    <i className='icon fa fa-plus' />
                    <div className='create-event-button-text'>Create event</div>
                </div>
            </Button>
        </div>
    );
};

export default HeaderComponent;
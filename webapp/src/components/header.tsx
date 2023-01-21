import React from 'react';
import Button from 'react-bootstrap/Button';
import { useDispatch } from 'react-redux';
import { openEventModal } from 'actions';

const HeaderComponent = () => {

    const dispatch = useDispatch();

    return (
        <div className='calendar-header-container'>
            <Button variant="primary" onClick={() => dispatch(openEventModal())}>
                <div className='create-event-button-inner-container'>
                    <i className='icon fa fa-plus' />
                    <div className='create-event-button-text'>Create event</div>
                </div>
            </Button>
        </div>
    );
};

export default HeaderComponent;
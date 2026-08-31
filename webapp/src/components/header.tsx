import React, {useEffect, useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import {Button, Dropdown, Option, useId} from '@fluentui/react-components';

import {Toggle} from '@fluentui/react/lib/Toggle';
import {DrawerHeader, DrawerHeaderTitle, DrawerOverlay, DrawerBody} from '@fluentui/react-components/unstable';

import {
    Apps20Regular,
    Calendar3Day20Regular,
    CalendarDay20Regular,
    CalendarEmpty16Filled,
    CalendarLtr20Regular,
    Call20Regular,
    PeopleCommunity20Regular,
    Dismiss24Regular,
} from '@fluentui/react-icons';

import {
    openEventModal,
    setSettingsPanelOpen,
    updateCalendarSettingsOnServer,
    updateSelectedCalendarType,
    updateSelectedCalendarView,
} from 'actions';

import {
    getCalendarSettings,
    getSelectedCalendarType,
    getSelectedCalendarView,
    selectIsSettingsPanelOpen,
} from '../selectors';
import {CalendarSettings} from '../types/settings';

import CalendarRef, {refetchCalendarEvents} from './calendar';
import ICalSettings from './ical-settings';

const HeaderComponent = () => {
    const dispatch = useDispatch();
    const settings: CalendarSettings = useSelector(getCalendarSettings);
    const selectedCalendarType: string = useSelector(getSelectedCalendarType);
    const selectedView: string = useSelector(getSelectedCalendarView);

    // opened from the settings button that lives in the calendar's own toolbar
    const settingsPanelOpen: boolean = useSelector(selectIsSettingsPanelOpen);

    const changeView = (view: string) => {
        dispatch(updateSelectedCalendarView(view));
        CalendarRef.current?.getApi().changeView(view);
    };

    // a color input fires onChange for every step of a drag, so the picked value
    // is kept locally and only saved once the user is done with it
    const [callColorDraft, setCallColorDraft] = useState<string>(settings.callColor);
    const [eventColorDraft, setEventColorDraft] = useState<string>(settings.eventColor);

    useEffect(() => {
        setCallColorDraft(settings.callColor);
    }, [settings.callColor]);

    useEffect(() => {
        setEventColorDraft(settings.eventColor);
    }, [settings.eventColor]);

    const saveCalendarColor = async (key: 'callColor' | 'eventColor', value: string) => {
        if (value === settings[key]) {
            return;
        }
        await dispatch(updateCalendarSettingsOnServer({
            ...settings,
            [key]: value,
        }));

        // colors are resolved server-side on read, so the grid has to be refetched
        // for the new one to show up on events that already exist
        refetchCalendarEvents();
    };

    const dayDropdown = useId('dropdown-dayDropdown');
    const dropdownDaysOfWeek = [
        {key: 0, text: 'Sunday'},
        {key: 1, text: 'Monday'},
        {key: 2, text: 'Tuesday'},
        {key: 3, text: 'Wednesday'},
        {key: 4, text: 'Thursday'},
        {key: 5, text: 'Friday'},
        {key: 6, text: 'Saturday'},
    ];

    const dayOfWeekByNumber = {
        0: 'Sunday',
        1: 'Monday',
        2: 'Tuesday',
        3: 'Wednesday',
        4: 'Thursday',
        5: 'Friday',
        6: 'Saturday',
    };

    return (
        <div className='calendar-header-container'>
            <DrawerOverlay
                open={settingsPanelOpen}
                position='end'
                modalType='non-modal'
                onOpenChange={(_, {open}) => dispatch(setSettingsPanelOpen(open))}
            >
                <DrawerHeader>
                    <DrawerHeaderTitle
                        action={
                            <Button
                                appearance='subtle'
                                aria-label='Close'
                                icon={<Dismiss24Regular/>}
                                onClick={() => dispatch(setSettingsPanelOpen(false))}
                            />
                        }
                    >
                        <span className='settings-pannel-header'>Settings</span>
                    </DrawerHeaderTitle>
                </DrawerHeader>

                <DrawerBody>
                    {/* a div, not a p: the rows below are block elements and the
                        browser would close the paragraph early, breaking the layout */}
                    <div className='settings-right-bar-content'>
                        <label id={dayDropdown}>First day of week</label>
                        <Dropdown
                            onOptionSelect={(event, item) => {
                                dispatch(updateCalendarSettingsOnServer({
                                    ...settings,
                                    firstDayOfWeek: Number(item.optionValue),
                                }));
                            }}
                            placeholder='Select day'
                            options={dropdownDaysOfWeek}
                            selectedOptions={[settings.firstDayOfWeek.toString()]}
                            value={dayOfWeekByNumber[settings.firstDayOfWeek]}
                        >
                            {dropdownDaysOfWeek.map((option) => (
                                <Option
                                    key={option.key}
                                    value={option.key.toString()}
                                >
                                    {option.text}
                                </Option>
                            ))}
                        </Dropdown>
                        <div className='settings-right-bar-hide-non-working-days'>
                            <Toggle
                                label='Hide non working days'
                                checked={settings.hideNonWorkingDays}
                                onChange={(ev, data) => {
                                    if (data === undefined) {
                                        return;
                                    }
                                    dispatch(updateCalendarSettingsOnServer({
                                        ...settings,
                                        hideNonWorkingDays: data,
                                    }));
                                }}
                            />
                        </div>
                        <div className='settings-right-bar-calendar-color'>
                            <label>Calls calendar color</label>
                            <input
                                type='color'
                                value={callColorDraft}
                                onChange={(e) => setCallColorDraft(e.target.value)}
                                onBlur={(e) => saveCalendarColor('callColor', e.target.value)}
                            />
                        </div>
                        <div className='settings-right-bar-calendar-color'>
                            <label>Events calendar color</label>
                            <input
                                type='color'
                                value={eventColorDraft}
                                onChange={(e) => setEventColorDraft(e.target.value)}
                                onBlur={(e) => saveCalendarColor('eventColor', e.target.value)}
                            />
                        </div>
                    </div>
                    <ICalSettings/>
                </DrawerBody>
            </DrawerOverlay>

            <div className='calendar-header-toolbar'>
                <Button
                    appearance='primary'
                    size='medium'
                    className='create-event-button'
                    onClick={() => dispatch(openEventModal())}
                    icon={<CalendarEmpty16Filled/>}
                >
                    <div className='create-event-button-text'>{'Create'}</div>
                </Button>

                {/* both groups share a wrapper so they wrap as whole rows on a
                    phone instead of splitting mid-group */}
                <div className='header-toolbar-groups'>
                    <div className='left-allign-header-toolbar-item'>
                        <Button
                            appearance='subtle'
                            icon={<CalendarDay20Regular/>}
                            onClick={() => changeView('timeGridDay')}
                            disabled={selectedView === 'timeGridDay'}
                        >Day</Button>
                        <Button
                            appearance='subtle'
                            icon={<Calendar3Day20Regular/>}
                            onClick={() => changeView('timeGridWeek')}
                            disabled={selectedView === 'timeGridWeek'}
                        >Week</Button>
                        <Button
                            appearance='subtle'
                            icon={<CalendarLtr20Regular/>}
                            onClick={() => changeView('dayGridMonth')}
                            disabled={selectedView === 'dayGridMonth'}
                        >Month</Button>
                    </div>
                    <div className='left-allign-header-toolbar-item'>
                        <Button
                            appearance='subtle'
                            icon={<Call20Regular/>}
                            onClick={() => dispatch(updateSelectedCalendarType('call'))}
                            disabled={selectedCalendarType === 'call'}
                        >Calls</Button>
                        <Button
                            appearance='subtle'
                            icon={<PeopleCommunity20Regular/>}
                            onClick={() => dispatch(updateSelectedCalendarType('event'))}
                            disabled={selectedCalendarType === 'event'}
                        >Events</Button>
                        <Button
                            appearance='subtle'
                            icon={<Apps20Regular/>}
                            onClick={() => dispatch(updateSelectedCalendarType('all'))}
                            disabled={selectedCalendarType === 'all'}
                        >All</Button>
                    </div>
                </div>
            </div>

        </div>
    );
};

export default HeaderComponent;
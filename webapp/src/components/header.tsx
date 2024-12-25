import React, {useState} from 'react';
import {useDispatch, useSelector} from 'react-redux';

import {Button, Dropdown, Option, useId} from '@fluentui/react-components';

import {Toggle} from '@fluentui/react/lib/Toggle';
import {DrawerHeader, DrawerHeaderTitle, DrawerOverlay, DrawerBody} from '@fluentui/react-components/unstable';

import {
    Calendar3Day20Regular,
    CalendarDay20Regular,
    CalendarEmpty16Filled,
    CalendarLtr20Regular,
    LineHorizontal3Regular,
    Settings20Regular,
    Dismiss24Regular,
} from '@fluentui/react-icons';

import {openEventModal, updateCalendarSettingsOnServer} from 'actions';

import {getCalendarSettings} from '../selectors';
import {CalendarSettings} from '../types/settings';

import CalendarRef from './calendar';

const HeaderComponent = () => {
    const dispatch = useDispatch();
    const settings: CalendarSettings = useSelector(getCalendarSettings);
    const [settingsPanelOpen, setSettingsPanelOpen] = useState<boolean>(false);
    const [selectedView, setSelectedView] = useState<string>('timeGridWeek');

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
                onOpenChange={(_, {open}) => setSettingsPanelOpen(open)}
            >
                <DrawerHeader>
                    <DrawerHeaderTitle
                        action={
                            <Button
                                appearance='subtle'
                                aria-label='Close'
                                icon={<Dismiss24Regular/>}
                                onClick={() => setSettingsPanelOpen(false)}
                            />
                        }
                    >
                        <span className='settings-pannel-header'>Settings</span>
                    </DrawerHeaderTitle>
                </DrawerHeader>

                <DrawerBody>
                    <p className='settings-right-bar-content'>
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
                    </p>
                </DrawerBody>
            </DrawerOverlay>

            <div className='calendar-header-toolbar'>
                <div className='left-allign-header-toolbar-item'>
                    <Button
                        appearance='subtle'
                        icon={<LineHorizontal3Regular/>}
                        onClick={
                            () => {
                                dispatch(updateCalendarSettingsOnServer({
                                    ...settings,
                                    isOpenCalendarLeftBar: !settings.isOpenCalendarLeftBar,
                                }));
                                CalendarRef.current.getApi().changeView(selectedView);
                            }
                        }

                    />
                    <Button
                        appearance='primary'
                        size='medium'
                        onClick={() => dispatch(openEventModal())}
                        icon={<CalendarEmpty16Filled/>}
                    >
                        <div className='create-event-button-text'>New event</div>
                    </Button>
                    <Button
                        appearance='subtle'
                        icon={<CalendarDay20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('dayGridDay');
                            setSelectedView('dayGridDay');
                        }}
                        disabled={selectedView === 'dayGridDay'}
                    >Day</Button>
                    <Button
                        appearance='subtle'
                        icon={<Calendar3Day20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('timeGridWeek');
                            setSelectedView('timeGridWeek');
                        }}
                        disabled={selectedView === 'timeGridWeek'}
                    >week</Button>
                    <Button
                        appearance='subtle'
                        icon={<CalendarLtr20Regular/>}
                        onClick={() => {
                            CalendarRef.current.getApi().changeView('dayGridMonth');
                            setSelectedView('dayGridMonth');
                        }}
                        disabled={selectedView === 'dayGridMonth'}
                    >month</Button>
                </div>
                <div className='left-allign-header-toolbar-item'>
                    <Button
                        appearance='subtle'
                        icon={<Settings20Regular/>}
                        onClick={() => {
                            setSettingsPanelOpen(true);
                        }}
                    />
                </div>
            </div>

        </div>
    );
};

export default HeaderComponent;
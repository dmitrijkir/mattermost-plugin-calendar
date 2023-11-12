// import logo from './logo.svg';
import './style.css';
import {addDays, addHours, addMinutes, differenceInMinutes, format, parse, set} from 'date-fns';
import React, {useEffect, useState} from 'react';
import {
    Dialog,
    DialogBody,
    DialogContent,
    DialogSurface,
    DialogTitle,
    Label,
    makeStyles,
    Persona,
    SpinButton,
    Spinner,
    Text,
    Toolbar,
    ToolbarButton,
    Tooltip,
    useId,
} from '@fluentui/react-components';
import {ChevronLeft24Regular, ChevronRight24Regular} from '@fluentui/react-icons';
import {DialogOpenChangeEventHandler} from '@fluentui/react-dialog';
import {UserProfile} from 'mattermost-redux/types/users';

import {useDispatch, useSelector} from 'react-redux';
import {getTeammateNameDisplaySetting} from 'mattermost-redux/selectors/entities/preferences';

import {ApiClient} from '../../client';
import {getCalendarSettings, getMembersAddedInEvent, getSelectedEventTime} from '../../selectors';
import {updateSelectedEventTime} from '../../actions';

const TimeInterval = 15;

const useOverrides = makeStyles({
    dialog: {maxWidth: '800px'},
});

interface BuildTimeLineProps {
    freeTimes: string[]
    onOpenChange?: DialogOpenChangeEventHandler
    duration: number
}

interface FindFreeTimeProps {
    open: boolean
    onOpenChange?: DialogOpenChangeEventHandler
}

interface UsersListProps {
    members: UserProfile[]
}

function PlanningAssistant(props: FindFreeTimeProps) {
    const displayNameSettings = useSelector(getTeammateNameDisplaySetting);
    const membersAddedInEvent = useSelector(getMembersAddedInEvent);
    const selectedEventTime = useSelector(getSelectedEventTime);
    const settings = useSelector(getCalendarSettings);

    const [isLoading, setIsLoading] = useState(true);

    const StarHour = settings.businessStartTime ? parseInt(settings.businessStartTime.split(':')[0], 10) : 0;
    const EndHour = settings.businessEndTime ? parseInt(settings.businessEndTime.split(':')[0], 10) : 0;

    const today = new Date();
    const [currentDate, setCurrentDate] = useState(selectedEventTime.start > today ? selectedEventTime.start : today)
    const [usersAvailability, setUsersAvailability] = useState(null);
    const [duration, setDuration] = useState(differenceInMinutes(selectedEventTime.end, selectedEventTime.start));

    const overrides = useOverrides();
    const slotTimeId = useId('slotTimeId');
    const dispatch = useDispatch();

    useEffect(() => {
        if (!props.open) {
            return;
        }
        const startTimeLine = set(currentDate, {hours: 0, seconds: 0, minutes: 0});
        const endEvent = addHours(startTimeLine, 24);
        const members = membersAddedInEvent.map((member: UserProfile) => member.id);

        setIsLoading(true);
        ApiClient.getUsersSchedule(
            members,
            format(startTimeLine, 'yyyy-MM-dd\'T\'HH:mm:ss'),
            format(endEvent, 'yyyy-MM-dd\'T\'HH:mm:ss'),
            duration).then((response) => {
            setUsersAvailability(response);
            setIsLoading(false);
        });
    }, [props.open, currentDate, duration]);

    const getDisplayUserName = (user: UserProfile | undefined) => {
        if (user === undefined) {
            return '';
        }

        if (displayNameSettings === 'full_name') {
            return user.first_name + ' ' + user.last_name;
        }
        if (displayNameSettings === 'username') {
            return user.username;
        }

        if (displayNameSettings === 'nickname_full_name') {
            if (user.nickname !== '') {
                return user.nickname;
            }
            return user.first_name + ' ' + user.last_name;
        }
    };

    function BuildHeader(props: BuildTimeLineProps) {
        const current = new Date();

        let start = set(current, {hours: StarHour, minutes: 0, seconds: 0, milliseconds: 0});
        const end = set(current, {hours: EndHour, minutes: 0, seconds: 0, milliseconds: 0});
        const minutes = differenceInMinutes(end, start) / TimeInterval;
        const columns = [];

        for (let i = 0; i < minutes; i++) {
            columns.push(start);
            start = addMinutes(start, TimeInterval);
        }
        return (
            <div className='time-header'>
                <div className='header-timeline'>
                    {columns.map((value) => {
                        if (props.freeTimes.includes(format(value, 'HH:mm'))) {
                            return (
                                <div
                                    className='time-column time-column-free-time'
                                    onClick={(event) => {
                                        dispatch(updateSelectedEventTime({
                                            start: value,
                                            startTime: format(value, 'HH:mm'),
                                            end: addMinutes(value, duration),
                                            endTime: format(addMinutes(value, props.duration), 'HH:mm'),
                                        }));
                                        if (props.onOpenChange) {
                                            props.onOpenChange(event, {open: false});
                                        }
                                    }}
                                >
                                    <Text weight='semibold'>
                                        {format(value, 'HH:mm')}
                                    </Text>
                                </div>
                            );
                        }
                        return (
                            <div
                                className='time-column'
                                onClick={(event) => {
                                    dispatch(updateSelectedEventTime({
                                        start: value,
                                        startTime: format(value, 'HH:mm'),
                                        end: addMinutes(value, props.duration),
                                        endTime: format(addMinutes(value, props.duration), 'HH:mm'),
                                    }));
                                    if (props.onOpenChange) {
                                        props.onOpenChange(event, {open: false});
                                    }
                                }}
                            >
                                <Text weight='semibold'>
                                    {format(value, 'HH:mm')}
                                </Text>
                            </div>
                        );
                    })}
                </div>
            </div>);
    }

    const BuildTimeLine = (props: BuildTimeLineProps) => {
        // const current = new Date();
        let start = set(currentDate, {hours: StarHour, minutes: 0, seconds: 0, milliseconds: 0});
        const end = set(currentDate, {hours: EndHour, minutes: 0, seconds: 0, milliseconds: 0});
        const minutes = differenceInMinutes(end, start) / TimeInterval;
        const columns = [];

        for (let i = 0; i < minutes; i++) {
            columns.push(start);
            start = addMinutes(start, TimeInterval);
        }
        return (
            <div className='time-line'>{columns.map((value, index) => {
                if (props.freeTimes.includes(format(value, 'HH:mm'))) {
                    return (
                        <div
                            className='time-column time-column-free-time'
                            onClick={(event) => {
                                dispatch(updateSelectedEventTime({
                                    start: value,
                                    startTime: format(value, 'HH:mm'),
                                    end: addMinutes(value, props.duration),
                                    endTime: format(addMinutes(value, props.duration), 'HH:mm'),
                                }));
                                if (props.onOpenChange) {
                                    props.onOpenChange(event, {open: false});
                                }
                            }}
                        />);
                }
                return (
                    <div
                        className='time-column'
                        onClick={(event) => {
                            dispatch(updateSelectedEventTime({
                                start: value,
                                startTime: format(value, 'HH:mm'),
                                end: addMinutes(value, props.duration),
                                endTime: format(addMinutes(value, props.duration), 'HH:mm'),
                            }));
                            if (props.onOpenChange) {
                                props.onOpenChange(event, {open: false});
                            }
                        }}
                    />);
            })}</div>);
    };

    const UsersList = () => {
        const membersById: Map<string, UserProfile> = new Map<string, UserProfile>();
        membersAddedInEvent.forEach((member: UserProfile) => {
            membersById.set(member.id, member);
        });
        if (usersAvailability == null) {
            return <div/>;
        }
        return (
            <>
                {Object.keys(usersAvailability.users).map((key) => {
                    return (
                        <div className='find-free-time-table-users-row'>
                            <Persona
                                name={getDisplayUserName(membersById.get(key))}
                            />
                        </div>
                    );
                })}
            </>
        );
    };

    const UsersTimeLine = () => {
        const start = set(currentDate, {hours: StarHour, minutes: 0, seconds: 0, milliseconds: 0});
        const end = set(currentDate, {hours: EndHour, minutes: 0, seconds: 0, milliseconds: 0});
        const pixels = differenceInMinutes(end, start) / TimeInterval * 50;
        if (usersAvailability == null) {
            return <div/>;
        }
        return (
            <>
                {
                    Object.keys(usersAvailability.users).map((userId) => {
                        return (
                            <div className='find-free-time-table-users-time-column'>
                                <BuildTimeLine
                                    freeTimes={usersAvailability == null ? [] : usersAvailability?.available_times}
                                    onOpenChange={props.onOpenChange}
                                    duration={duration}
                                />
                                {
                                    usersAvailability.users[userId].map((event) => {
                                        const current = set(currentDate, {hours: StarHour, seconds: 0, minutes: 0});
                                        const startTime = parse(event.start, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());

                                        const leftPad = differenceInMinutes(startTime, current) / TimeInterval * 50;

                                        if (leftPad >= pixels) {
                                            return <div/>;
                                        }
                                        if ((leftPad + event.duration / TimeInterval * 50) > pixels) {
                                            const diff = (leftPad + event.duration / TimeInterval * 50) - pixels;
                                            return (
                                                <div
                                                    className='event-container'
                                                    style={{
                                                        left: leftPad,
                                                        width: (event.duration / TimeInterval * 50) - diff,
                                                    }}
                                                />
                                            );
                                        }

                                        return (
                                            <div
                                                className='event-container'
                                                style={{
                                                    left: leftPad,
                                                    width: event.duration / TimeInterval * 50,
                                                }}
                                            />
                                        );
                                    })
                                }
                            </div>);
                    })
                }
            </>
        );
    };

    const onSpinButtonChange = React.useCallback(
        (_ev, data) => {
            if (data.value !== undefined) {
                setDuration(data.value);
            } else if (data.displayValue !== undefined) {
                const newValue = parseFloat(data.displayValue);
                if (!Number.isNaN(newValue)) {
                    setDuration(newValue);
                } else {
                    console.error(`Cannot parse "${data.displayValue}" as a number.`);
                }
            }
        },
        [setDuration],
    );

    return (
        <Dialog
            open={props.open}
            onOpenChange={(event, data) => (props.onOpenChange ? props.onOpenChange(event, data) : null)}
            modalType='non-modal'
            inertTrapFocus={true}
        >
            <DialogSurface className={overrides.dialog}>
                <DialogBody>
                    <DialogTitle>
                        <div className='find-free-time-current-date'>
                            <Toolbar aria-label='Default'>
                                <ToolbarButton
                                    aria-label='today'
                                    onClick={() => setCurrentDate(today)}
                                >{'today'}</ToolbarButton>
                                <ToolbarButton
                                    aria-label='prev day'
                                    icon={<ChevronLeft24Regular/>}
                                    onClick={() => setCurrentDate(addDays(currentDate, -1))}
                                />
                                <ToolbarButton
                                    aria-label='next day'
                                    icon={<ChevronRight24Regular/>}
                                    onClick={() => setCurrentDate(addDays(currentDate, 1))}
                                />
                                <ToolbarButton
                                    aria-label='next day'
                                >{format(currentDate, 'EEE, dd MMMM yyyy')}</ToolbarButton>
                                <div className='slot-duration-select-container'>
                                    <Tooltip
                                        withArrow={true}
                                        content={'event duration in minutes'}
                                    >
                                        <Label
                                            htmlFor={slotTimeId}
                                            className='event-duration-label'
                                        >
                                            {'duration'}
                                        </Label>
                                    </Tooltip>

                                    <SpinButton
                                        className='slot-duration-select-input'
                                        id={slotTimeId}
                                        defaultValue={15}
                                        appearance='underline'
                                        value={duration}
                                        onChange={onSpinButtonChange}
                                    />
                                </div>
                                {isLoading ? <Spinner size='tiny'/> : <div/>}
                            </Toolbar>
                        </div>
                    </DialogTitle>
                    <DialogContent>
                        {usersAvailability == null ? <Spinner size='huge'/> :
                            <div className='find-free-time-table-container'>
                                <div className='find-free-time-table-left-nav'>
                                    <div className='find-free-time-table-users-column'>
                                        <UsersList/>

                                    </div>
                                </div>
                                <div className='find-free-time-table-body-container'>
                                    <div className='find-free-time-table-body'>
                                        <div className='find-free-time-table-header'>
                                            <BuildHeader
                                                freeTimes={usersAvailability == null ? [] : usersAvailability.available_times}
                                                {...props}
                                            />
                                        </div>
                                        <UsersTimeLine/>
                                    </div>
                                </div>
                            </div>
                        }
                    </DialogContent>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default PlanningAssistant;
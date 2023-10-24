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
    Text,
    Toolbar,
    ToolbarButton,
    useId,
} from '@fluentui/react-components';
import {ChevronLeft24Regular, ChevronRight24Regular} from '@fluentui/react-icons';
import {DialogOpenChangeEventHandler} from '@fluentui/react-dialog';
import {UserProfile} from "mattermost-redux/types/users";
import {ApiClient} from "../../client";
import {useSelector} from "react-redux";
import {getTeammateNameDisplaySetting} from "mattermost-redux/selectors/entities/preferences";
import {getMembersAddedInEvent} from "../../selectors";

const StarHour = 8;
const EndHour = 20;
const TimeInterval = 15;

const useOverrides = makeStyles({
    dialog: {maxWidth: '800px'},
});

interface BuildTimeLineProps {
    freeTimes: string[]
}

interface FindFreeTimeProps {
    open: boolean
    onOpenChange?: DialogOpenChangeEventHandler
    onSlotSelected?: (data: Date) => void
}

interface UsersListProps {
    members: UserProfile[]
}

function BuildTimeLine(props: BuildTimeLineProps) {
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
        <div className='time-line'>{columns.map((value, index) => {
            if (props.freeTimes.includes(format(value, 'HH:mm'))) {
                return (
                    <div
                        className='time-column time-column-free-time'
                        onClick={(event) => console.log(value)}
                    />);
            }
            return (<div
                className='time-column'
                onClick={(event) => console.log(value)}
            ></div>);
        })}</div>);
}

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
                {columns.map((value, index) => {
                    if (props.freeTimes.includes(format(value, 'HH:mm'))) {
                        return (<div className='time-column time-column-free-time'><Text
                            weight='semibold'
                        >{format(value, 'HH:mm')}</Text></div>);
                    }
                    return (<div className='time-column'><Text weight='semibold'>{format(value, 'HH:mm')}</Text></div>);
                })}
            </div>
        </div>);
}

function FindTimeFree(props: FindFreeTimeProps) {
    const displayNameSettings = useSelector(getTeammateNameDisplaySetting);
    const membersAddedInEvent = useSelector(getMembersAddedInEvent);

    const today = new Date();
    const [currentDate, setCurrentDate] = useState(today);
    const [usersAvailability, setUsersAvailability] = useState(null);

    const overrides = useOverrides();
    const slotTimeId = useId('slotTimeId');

    useEffect(() => {
        if (!props.open) {
            return;
        }
        const startTimeLine = set(currentDate, {hours: 0, seconds: 0, minutes: 0});
        const endEvent = addHours(startTimeLine, 24);
        const members = membersAddedInEvent.map((member) => member.id);
        ApiClient.getUsersSchedule(
            members,
            format(startTimeLine, 'yyyy-MM-dd\'T\'HH:mm:ss'),
            format(endEvent, 'yyyy-MM-dd\'T\'HH:mm:ss')).then((response) => {
            setUsersAvailability(response);
        });
    }, [props.open]);

    const getDisplayUserName = (user: UserProfile) => {
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

    const UsersList = () => {
        const membersById: Map<string, UserProfile> = new Map<string, UserProfile>();
        membersAddedInEvent.forEach((member) => {
            membersById.set(member.id, member);
        });
        if (usersAvailability == null) {
            return <div/>;
        }
        return (
            Object.keys(usersAvailability.users).map((key) => {
                return <div className='find-free-time-table-users-row'><Persona
                    name={getDisplayUserName(membersById.get(key))}/></div>;
            })
        );
    };

    const UsersTimeLine = () => {
        if (usersAvailability == null) {
            return <div></div>;
        }
        return (
            Object.keys(usersAvailability.users).map((userId) => {
                return <div className='find-free-time-table-users-time-column'>
                    <BuildTimeLine freeTimes={usersAvailability.available_times}/>
                    {
                        usersAvailability.users[userId].map((event, index) => {

                            const current = set(currentDate, {hours: StarHour, seconds: 0, minutes: 0});
                            const startTime = parse(event.start, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());

                            let leftPad = differenceInMinutes(startTime, current) / TimeInterval * 50;

                            return <div
                                className='event-container'
                                style={{
                                    left: leftPad,
                                    width: event.duration / TimeInterval * 50
                                }}
                                onClick={() => {
                                    console.log(current);
                                }}
                            >
                            </div>
                        })
                    }
                </div>
            })
        )
    };

    return (
        <Dialog
            open={props.open}
            onOpenChange={(event, data) => props.onOpenChange(event, data)}
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
                                >today</ToolbarButton>
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
                                >{format(currentDate, 'dd MMMM yyyy')}</ToolbarButton>
                                <div className='slot-duration-select-container'>
                                    <Label htmlFor={slotTimeId}>duration</Label>
                                    <SpinButton
                                        className='slot-duration-select-input'
                                        id={slotTimeId}
                                        defaultValue={15}
                                        appearance='underline'
                                    />
                                </div>
                            </Toolbar>
                        </div>
                    </DialogTitle>
                    <DialogContent>
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
                                        />
                                    </div>
                                    <UsersTimeLine/>
                                </div>
                            </div>
                        </div>
                    </DialogContent>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
}

export default FindTimeFree;
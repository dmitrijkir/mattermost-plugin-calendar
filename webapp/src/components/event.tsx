import React, { useEffect, useState } from 'react';
import Button from 'react-bootstrap/Button';
import Modal from 'react-bootstrap/Modal';
import Form from 'react-bootstrap/Form';
import FormControl from 'react-bootstrap/FormControl';
import { Client4 } from 'mattermost-redux/client';
import { UserProfile } from 'mattermost-redux/types/users';
import { Channel } from 'mattermost-redux/types/channels';
import CalendarRef from './calendar';
import { ApiClient } from 'client';
import { useSelector, useDispatch } from 'react-redux';
import { selectSelectedEvent, selectIsOpenEventModal } from 'selectors';
import { closeEventModal, eventSelected } from 'actions';
import { getCurrentTeamId } from 'mattermost-redux/selectors/entities/teams';
import {getTheme} from  "mattermost-redux/selectors/entities/preferences";


interface AddedUserComponentProps {
    user: UserProfile
}


interface TimeSelectItemsProps {
    start?: string;
    end?: string;
}

function formatTimeWithZero(i: number) {
    if (i < 10) {
        return "0" + i.toString();
    }
    return i.toString();
}

const TimeSelectItems = (props: TimeSelectItemsProps) => {
    let now = new Date();
    let times = [
        "00:00",
        "00:30",
        "01:00",
        "01:30",
        "02:00",
        "02:30",
        "03:00",
        "03:30",
        "04:00",
        "04:30",
        "05:00",
        "05:30",
        "06:00",
        "06:30",
        "07:00",
        "07:30",
        "08:00",
        "08:30",
        "09:00",
        "09:30",
        "10:00",
        "10:30",
        "11:00",
        "11:30",
        "12:00",
        "12:30",
        "13:00",
        "13:30",
        "14:00",
        "14:30",
        "15:00",
        "15:30",
        "16:00",
        "16:30",
        "17:00",
        "17:30",
        "18:00",
        "18:30",
        "19:00",
        "19:30",
        "20:00",
        "20:30",
        "21:00",
        "21:30",
        "22:00",
        "22:30",
        "23:00",
        "23:30"

    ]
    return (

        <>
            {times.map((time, index) => {
                if (props.start != null && props.start == time) {
                    return <option value={time} selected>{time}</option>
                } 
                if (props.end != null && props.end == time) {
                    return <option value={time} selected>{time}</option>
                }
                return <option value={time}>{time}</option>
                // if (props.start != null) {

                //     if (props.start == time) {
                //         return <option value={time} selected>{time}</option>
                //     }
                // }
                // let minutes = 0;
                // let hours = 0;
                // if (now.getMinutes() < 30) {
                //     minutes = 30
                //     hours = now.getHours();
                // } else {
                //     hours = (now.getHours() + 1);
                // }

                // if (time == formatTimeWithZero(hours) + ':' + formatTimeWithZero(minutes)) {
                //     return <option value={time} selected>{time}</option>
                // }
                // return <option value={time}>{time}</option>
            })}
        </>
    );
}

const getNextHour = (hour: number) => {
    if (hour == 23) {
        return 0
    }
    return hour + 1;
}
const initialStartTime = () => {
    let now = new Date();
    let minutes = 0;
    let hours = 0;
    if (now.getMinutes() < 30) {
        minutes = 30
        hours = now.getHours();
    } else {
        hours = getNextHour(now.getHours());
    }
    return formatTimeWithZero(hours) + ':' + formatTimeWithZero(minutes);
}

const initialEndTime = () => {
    let now = new Date();
    let minutes = 0;
    let hours = getNextHour(now.getHours());
    if (now.getMinutes() < 30) {
        minutes = 0
    } else {
        minutes = 30
    }
    return formatTimeWithZero(hours) + ':' + formatTimeWithZero(minutes);
}
const EventModalComponent = () => {
    const selectedEvent = useSelector(selectSelectedEvent)
    const isOpenEventModal = useSelector(selectIsOpenEventModal);

    const theme = useSelector(getTheme);

    const CurrentTeamId = useSelector(getCurrentTeamId);

    const dispatch = useDispatch();

    let now = new Date();
    let initialDate = now.toISOString().split('T')[0];

    const [showUserList, setshowUserList] = useState(false);

    const [usersAutocomplete, setUsersAutocomplete] = useState<UserProfile[]>([]);
    const [usersAdded, setUsersAdded] = useState<UserProfile[]>([]);

    const [searchUsersInput, setSearchUsersInput] = useState("");

    const [channelsAutocomplete, setChannelsAutocomplete] = useState<Channel[]>([]);
    const [searchChannelInput, setSearchChannelInput] = useState("");
    const [selectedChannel, setSelectedChannel] = useState({});
    const [showChannelsList, setShowChannelsList] = useState(false);

    const [selectedDays, setSelectedDays] = useState([]);

    const [titleEvent, setTitleEvent] = useState("");
    const [startEventData, setStartEventDate] = useState(initialDate);
    const [endEventData, setEndEventDate] = useState(initialDate);

    const [startEventTime, setStartEventTime] = useState(initialStartTime);
    const [endEventTime, setEndEventTime] = useState(initialEndTime);

    // methods
    const ViewEventModalHandleClose = () => {
        cleanState();
        dispatch(closeEventModal());
        dispatch(eventSelected({}));
    };

    const cleanState = () => {
        setTitleEvent("");
        setStartEventTime(initialStartTime);
        setEndEventTime(initialEndTime);
        setStartEventDate(initialDate);
        setEndEventDate(initialDate);
        setUsersAutocomplete([]);
        setChannelsAutocomplete([]);
        setSearchUsersInput("");
        setSearchChannelInput("")

        setSelectedDays([]);

        setSelectedChannel({});
        setUsersAdded([]);

    }
    const onTitleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTitleEvent(event.target.value)
    }

    const onStartDateChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setStartEventDate(event.target.value);
    }

    const onEndDateChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setEndEventDate(event.target.value);

    }

    const onStartTimeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setStartEventTime(event.target.value);

    }

    const onEndTimeChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setEndEventTime(event.target.value);
    }

    const onInputUserAction = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setSearchUsersInput(event.target.value);
        if (event.target.value != "") {
            let resp = await Client4.searchUsers(event.target.value, "");
            setUsersAutocomplete(resp);
            setshowUserList(true);
            return
        }

        setshowUserList(false);

    }

    const onInputChannelAction = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setSearchChannelInput(event.target.value);
        if (event.target.value != "") {
            let resp = await Client4.autocompleteChannels(CurrentTeamId, event.target.value);
            resp.push({
                id: 'empty',
                create_at: 0,
                update_at: 0,
                delete_at: 0,
                team_id: '',
                type: 'O',
                display_name: 'Empty',
                name: '',
                header: '',
                purpose: '',
                last_post_at: 0,
                total_msg_count: 0,
                extra_update_at: 0,
                creator_id: '',
                scheme_id: '',
                group_constrained: false
            });
            setChannelsAutocomplete(resp);
            setShowChannelsList(true);
            return
        }

        setshowUserList(false);

    }

    const onDaySelected = (day: number) => {
        let foundIndex = selectedDays.findIndex(elem => elem === day);
        if (foundIndex != -1) {
            let slice = selectedDays.filter((value, ind, arr) => value != day);
            setSelectedDays(slice);
            return
        }
        setSelectedDays(selectedDays.concat([day]));
    };

    const onSaveEvent = async () => {
        let members: string[] = usersAdded.map((user) => user.id);
        console.log(selectedEvent);
        if (selectedEvent?.event?.id == null) {
            let response = await ApiClient.createEvent(
                titleEvent,
                startEventData + 'T' + startEventTime + ':00Z',
                endEventData + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                selectedDays,
            );
            CalendarRef.current.getApi().getEventSources()[0].refetch();
            cleanState();
            ViewEventModalHandleClose();
            return
        } else {
            let response = await ApiClient.updateEvent(
                selectedEvent.event.id,
                titleEvent,
                startEventData + 'T' + startEventTime + ':00Z',
                endEventData + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                selectedDays,
            );
            CalendarRef.current.getApi().getEventSources()[0].refetch();
            cleanState();
            ViewEventModalHandleClose();
            return
        }
        
    }

    const onSelectedDay = (day: number): boolean => {
        let found = selectedDays.find(elem => elem === day);
        if (found == -1 || found == undefined) {
            return false;
        }

        return true;
    }

    const onRemoveEvent = async () => {
        await ApiClient.removeEvent(selectedEvent.event.id);
        CalendarRef.current.getApi().getEventSources()[0].refetch();
        cleanState();
        ViewEventModalHandleClose();
    }

    useEffect(() => {
        let mounted = true;
        if (mounted && selectedEvent?.event?.id != null) {
            console.log("theme");
            console.log(theme);
            ApiClient.getEventById(selectedEvent.event.id).then((data) => {
                setTitleEvent(data.data.title);
                setStartEventDate(data.data.start.split('T')[0]);
                setEndEventDate(data.data.end.split('T')[0]);
                setUsersAdded(data.data.attendees);
                setStartEventTime(data.data.start.split('T')[1].split(':')[0] + ':' + data.data.start.split('T')[1].split(':')[1]);
                setEndEventTime(data.data.end.split('T')[1].split(':')[0] + ':' + data.data.end.split('T')[1].split(':')[1]);

                if (data.data.recurrence != null) {
                    setSelectedDays(data.data.recurrence);
                }

                if (data.data.channel != null) {
                    Client4.getChannel(data.data.channel).then((channel) => {
                        setSelectedChannel(channel);
                        setSearchChannelInput(channel.display_name);
                    });
                }
            });
        } else if (mounted && selectedEvent?.event?.id == null && selectedEvent?.event?.start != null) {
            setStartEventDate(selectedEvent?.event.start.split('T')[0]);
            setEndEventDate(selectedEvent?.event.end.split('T')[0]);
                
            setStartEventTime(selectedEvent?.event.start.split('T')[1].split(':')[0] + ':' + selectedEvent?.event.start.split('T')[1].split(':')[1]);
            setEndEventTime(selectedEvent?.event.end.split('T')[1].split(':')[0] + ':' + selectedEvent?.event.end.split('T')[1].split(':')[1]);
        }
        mounted = false;
        return;

    }, [selectedEvent]);


    // Components

    const AddedUserComponent = (props: AddedUserComponentProps) => {

        return <span className='added-user-badge-container'>
            {props.user.username}
            <i className='icon fa fa-times' onClick={() => {
                setUsersAdded(usersAdded.filter(item => item.id != props.user.id))
            }} />

        </span>
    };


    const UserListAutocomplite = () => {
        if (showUserList) {
            return <div className='add-users-list'>
                {
                    usersAutocomplete.map((user) => {
                        return <div onClick={() => {
                            let foundIndex = usersAdded.findIndex(e => e.id === user.id);
                            if (foundIndex > -1) {
                                setSearchUsersInput("");
                                setshowUserList(false);
                                return
                            }
                            setUsersAdded(usersAdded.concat([user]))
                            setSearchUsersInput("");
                            setshowUserList(false);
                        }} className='user-list-item'>{user.username}</div>
                    }
                    )}
            </div>
        }
        return <></>
    }

    const ChannelsListAutocomplite = () => {
        if (showChannelsList) {
            return <div className='add-channels-list'>
                {
                    channelsAutocomplete.map((channel) => {
                        return <div onClick={() => {
                            setSelectedChannel(channel);
                            setSearchChannelInput(channel.display_name);
                            setShowChannelsList(false);
                        }} className='channels-list-item'>{channel.display_name}</div>
                    }
                    )}
            </div>
        }
        return <></>
    }

    const UsersAddedComponent = () => {
        if (usersAdded.length > 0) {
            return <div className='added-users-list'>
                {
                    usersAdded.map((user) => {
                        return <AddedUserComponent user={user} />
                    }
                    )}
            </div>
        }
        return <></>
    }

    
    const RemoveEventButton = () => {
        if (Object.keys(selectedEvent).length !== 0) {
            return <Button variant="danger" onClick={onRemoveEvent}>Remove event</Button>
        }
        return <></>
    }


    return (
        <Modal className='modal-view-event' show={isOpenEventModal} onHide={ViewEventModalHandleClose} centered animation={false}>
            <Modal.Header closeButton className='create-event-modal-header' style={{backgroundColor: theme.sidebarTeamBarBg, color: theme.sidebarHeaderTextColor}}>
                <Modal.Title className='create-event-modal-title'>Create new event</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <Form className='create-event-form'>
                    <div className='event-title-container'>
                        <i className='icon fa fa-pencil' />
                        <div className='event-input-container'>
                            <FormControl type="text" placeholder='Enter event title' value={titleEvent} onChange={onTitleChange}></FormControl>
                        </div>
                    </div>
                    <div className='datetime-container'>
                        <i className='icon fa fa-clock-o' />
                        <div className='event-input-container'>
                            <FormControl type="date" className='start-date-input' value={startEventData} onChange={onStartDateChange}></FormControl>
                            <Form.Control as="select" onChange={onStartTimeChange}>
                                <TimeSelectItems start={startEventTime} />
                            </Form.Control>
                            <span className='date-arrow'><i className='icon fa fa-arrow-right' /></span>
                            <FormControl type="date" className='end-date-input' value={endEventData} onChange={onEndDateChange}></FormControl>
                            <Form.Control as="select" onChange={onEndTimeChange}>
                                <TimeSelectItems end={endEventTime} />
                            </Form.Control>
                        </div>
                    </div>

                    <div className='event-add-users-container'>
                        <i className='icon fa fa-user-plus' />
                        <div className='event-input-container'>
                            <div className='event-input-users-wrapper'>
                                <FormControl type="text" placeholder='Add users' value={searchUsersInput} onChange={onInputUserAction}></FormControl>
                                <UserListAutocomplite />
                            </div>


                        </div>


                    </div>
                    <div className='users-added-container'>
                        <UsersAddedComponent />
                    </div>
                    <div className='event-channel-container'>
                        <i className='icon fa fa-comments-o' />
                        <div className='event-channel-input-container'>
                            <div className='event-input-channel-wrapper'>
                                <FormControl type="text" placeholder='Select channel' value={searchChannelInput} onChange={onInputChannelAction}></FormControl>
                                <ChannelsListAutocomplite />
                            </div>


                        </div>


                    </div>

                    <div className='recurrence-container'>
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='mon'
                            label='Mon'
                            checked={onSelectedDay(1)}
                            onChange={() => onDaySelected(1)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='tues'
                            label='Tues'
                            checked={onSelectedDay(2)}
                            onChange={() => onDaySelected(2)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='wed'
                            label='Wed'
                            checked={onSelectedDay(3)}
                            onChange={() => onDaySelected(3)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='thurs'
                            label='Thurs'
                            checked={onSelectedDay(4)}
                            onChange={() => onDaySelected(4)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='fri'
                            label='Fri'
                            checked={onSelectedDay(5)}
                            onChange={() => onDaySelected(5)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='sat'
                            label='Sat'
                            checked={onSelectedDay(6)}
                            onChange={() => onDaySelected(6)}
                        />
                        <Form.Check
                            className='form-check-days'
                            type='checkbox'
                            id='sun'
                            label='Sun'
                            checked={onSelectedDay(0)}
                            onChange={() => onDaySelected(0)}
                        />
                    </div>
                </Form>
            </Modal.Body>
            <Modal.Footer>
                <div className='event-modal-footer'>
                    <div>
                        <RemoveEventButton />
                    </div>
                    <div>
                        <Button variant="primary" onClick={onSaveEvent}>
                            Save
                        </Button>
                        <Button variant="secondary" onClick={ViewEventModalHandleClose}>
                            Close
                        </Button>
                    </div>
                </div>
            </Modal.Footer>
        </Modal>
    );
};

export default EventModalComponent;
import React, {useEffect, useRef, useState} from 'react';
// import Button from 'react-bootstrap/Button';
import {Client4} from 'mattermost-redux/client';
import {UserProfile} from 'mattermost-redux/types/users';
import {Channel} from 'mattermost-redux/types/channels';
import CalendarRef from './calendar';
import {ApiClient} from 'client';
import {useDispatch, useSelector} from 'react-redux';
import {selectIsOpenEventModal, selectSelectedEvent} from 'selectors';
import {closeEventModal, eventSelected} from 'actions';
import {getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {getUserStatuses} from 'mattermost-redux/selectors/entities/users';
import {getTeammateNameDisplaySetting, getTheme} from "mattermost-redux/selectors/entities/preferences";
import RepeatEventCustom from './repeat-event';
import {
    ChatMultiple24Regular,
    Clock24Regular,
    Delete16Regular,
    Dismiss12Regular,
    Pen24Regular,
    PersonAdd24Regular,
    Save16Regular,
} from "@fluentui/react-icons";
import {
    Button,
    Combobox,
    Dialog,
    DialogActions,
    DialogBody,
    DialogContent,
    DialogSurface,
    DialogTrigger,
    Input,
    Option,
    Persona,
    Select,
} from "@fluentui/react-components";

interface AddedUserComponentProps {
    user: UserProfile
}

interface TimeSelectItemsProps {
    start?: string;
    end?: string;
}

type SelectionEvents =
    React.ChangeEvent<HTMLElement>
    | React.KeyboardEvent<HTMLElement>
    | React.MouseEvent<HTMLElement, MouseEvent>
declare type OptionOnSelectData = {
    optionValue: string | undefined;
    optionText: string | undefined;
    selectedOptions: string[];
};

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
    const selectedEvent = useSelector(selectSelectedEvent);
    const isOpenEventModal = useSelector(selectIsOpenEventModal);

    const displayNameSettings = useSelector(getTeammateNameDisplaySetting);

    const theme = useSelector(getTheme);

    const CurrentTeamId = useSelector(getCurrentTeamId);
    const UserStatusSelector = useSelector(getUserStatuses);

    const dispatch = useDispatch();

    const now = new Date();
    const initialDate = now.toISOString().split('T')[0];

    const [usersAutocomplete, setUsersAutocomplete] = useState<UserProfile[]>([]);
    const [usersAdded, setUsersAdded] = useState<UserProfile[]>([]);
    // const UserInputRef = useRef();

    const [searchUsersInput, setSearchUsersInput] = useState('');
    const [selectedColor, setSelectedColor] = useState('#D0D0D0');

    const [channelsAutocomplete, setChannelsAutocomplete] = useState<Channel[]>([]);
    const [selectedChannel, setSelectedChannel] = useState({});
    const [selectedChannelText, setSelectedChannelText] = useState('');

    const [titleEvent, setTitleEvent] = useState('');
    const [startEventData, setStartEventDate] = useState(initialDate);
    const [endEventData, setEndEventDate] = useState(initialDate);

    const [startEventTime, setStartEventTime] = useState(initialStartTime);
    const [endEventTime, setEndEventTime] = useState(initialEndTime);

    // const repeatRule = useRef("RRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=TU,SA");
    const repeatRule = useRef('');
    const [showCustomRepeat, setShowCustomRepeat] = useState(false);
    const [repeatOption, setRepeatOption] = useState("Don't repeat");
    const [repeatOptionsSelected, setRepeatOptionsSelected] = useState(['empty'])

    // methods
    const viewEventModalHandleClose = () => {
        cleanState();
        dispatch(closeEventModal());
        dispatch(eventSelected({}));
    };

    const cleanState = () => {
        setTitleEvent('');
        setStartEventTime(initialStartTime);
        setEndEventTime(initialEndTime);
        setStartEventDate(initialDate);
        setEndEventDate(initialDate);
        setUsersAutocomplete([]);
        setChannelsAutocomplete([]);
        setSelectedChannelText('');
        setSelectedChannel({});
        setSearchUsersInput('');
        setShowCustomRepeat(false);

        repeatRule.current = '';

        setSelectedChannel({});
        setUsersAdded([]);
        setSelectedColor('#D0D0D0');
    };

    const onTitleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTitleEvent(event.target.value);
    };

    const onStartDateChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setStartEventDate(event.target.value);
    };

    const onEndDateChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setEndEventDate(event.target.value);
    };

    const onStartTimeChange = (event: React.ChangeEvent<HTMLSelectElement>, data: any) => {
        setStartEventTime(event.target.value);
    };

    const onEndTimeChange = (event: React.ChangeEvent<HTMLSelectElement>, data: any) => {
        setEndEventTime(event.target.value);
    };

    const onInputUserAction = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setSearchUsersInput(event.target.value);
        if (event.target.value !== '') {
            const resp = await Client4.searchUsers(event.target.value, '');
            setUsersAutocomplete(resp);
        }
    };

    const onSelectChannelOption = (event: SelectionEvents, data: OptionOnSelectData) => {
        channelsAutocomplete.map((option) => {
            if (option.id === data.optionValue) {
                setSelectedChannel(option);
                setSelectedChannelText(option.display_name);
                return;
            }
        });
    };

    const onInputChannelAction = async (event: React.ChangeEvent<HTMLInputElement>) => {
        setSelectedChannelText(event.target.value);
        if (event.target.value !== '') {
            const resp = await Client4.autocompleteChannels(CurrentTeamId, event.target.value);
            setChannelsAutocomplete(resp);
        } else {
            // if channel input empty, remove selected channel
            setSelectedChannel({});
        }
    };

    const onSaveEvent = async () => {
        const members: string[] = usersAdded.map((user) => user.id);
        console.log(selectedEvent);
        if (selectedEvent?.event?.id == null) {
            const response = await ApiClient.createEvent(
                titleEvent,
                startEventData + 'T' + startEventTime + ':00Z',
                endEventData + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                repeatRule.current,
                selectedColor,
            );
            CalendarRef.current.getApi().getEventSources()[0].refetch();
            cleanState();
            viewEventModalHandleClose();
            return
        } else {
            let response = await ApiClient.updateEvent(
                selectedEvent.event.id,
                titleEvent,
                startEventData + 'T' + startEventTime + ':00Z',
                endEventData + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                repeatRule.current,
                selectedColor,
            );
            CalendarRef.current.getApi().getEventSources()[0].refetch();
            cleanState();
            viewEventModalHandleClose();
            return;
        }
    };

    const onRemoveEvent = async () => {
        await ApiClient.removeEvent(selectedEvent.event.id);
        CalendarRef.current.getApi().getEventSources()[0].refetch();
        cleanState();
        viewEventModalHandleClose();
    };

    const onSelectColor = (color: string | null) => {
        if (color == null) {
            setSelectedColor('#D0D0D0');
            return;
        }
        setSelectedColor(color);
    };

    useEffect(() => {
        let mounted = true;
        if (mounted && selectedEvent?.event?.id != null) {
            console.log("display settings");
            console.log(displayNameSettings);
            ApiClient.getEventById(selectedEvent.event.id).then((data) => {
                setTitleEvent(data.data.title);
                setStartEventDate(data.data.start.split('T')[0]);
                setEndEventDate(data.data.end.split('T')[0]);
                setUsersAdded(data.data.attendees);
                setStartEventTime(data.data.start.split('T')[1].split(':')[0] + ':' + data.data.start.split('T')[1].split(':')[1]);
                setEndEventTime(data.data.end.split('T')[1].split(':')[0] + ':' + data.data.end.split('T')[1].split(':')[1]);
                setSelectedColor(data.data.color!);

                if (data.data.recurrence != null) {
                    // setSelectedDays(data.data.recurrence);
                }

                if (data.data.channel != null) {
                    Client4.getChannel(data.data.channel).then((channel) => {
                        setSelectedChannel(channel);
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

    const getDisplayUserName = (user: UserProfile) => {
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

    const repeatOnSelect = (event: SelectionEvents, data: OptionOnSelectData) => {
        if (data.optionValue === 'custom') {
            setRepeatOption('Custom');
            setShowCustomRepeat(true);
            setRepeatOptionsSelected(['custom']);
        } else {
            setRepeatOption("Don't repeat");
            setShowCustomRepeat(false);
            setRepeatOptionsSelected(['empty']);
        }
    };

    // Components
    // full_name, nickname_full_name, username
    const AddedUserComponent = (props: AddedUserComponentProps) => {
        let stat = 'unknown';
        if (UserStatusSelector[props.user.id] === 'online') {
            stat = 'available';
        }

        return <span className='added-user-badge-container'>
            <Persona name={getDisplayUserName(props.user)} avatar={{color: "colorful"}} presence={{status: stat}}/>
            <Dismiss12Regular className='added-user-badge-icon-container' onClick={() => {
                setUsersAdded(usersAdded.filter(item => item.id != props.user.id))
            }}/>

        </span>
    };


    const UsersAddedComponent = () => {
        if (usersAdded.length > 0) {
            return <div className='added-users-list'>
                {
                    usersAdded.map((user) => {
                        return <AddedUserComponent user={user}/>
                    })
                }
            </div>
        }
        return <></>
    }


    const RemoveEventButton = () => {
        if (selectedEvent?.event?.id != null) {
            return <DialogTrigger disableButtonEnhancement>
                <Button appearance="secondary" icon={<Delete16Regular/>}
                        onClick={viewEventModalHandleClose}>Remove</Button>
            </DialogTrigger>

        }
        return <></>
    }


    const RepeatComponent = () => {
        if (showCustomRepeat) {
            return <RepeatEventCustom
                rrule={repeatRule.current}
                onChange={(respRule) => {
                    console.log(respRule);
                    repeatRule.current = respRule;
                }}
            />;
        }
        return <></>;
    };

    return (
        <Dialog open={isOpenEventModal}>
            <DialogSurface>
                <DialogBody className='event-modal'>
                    <DialogContent className='modal-container'>
                        <div className='event-title-container'>
                            <Pen24Regular/>
                            <div className='event-input-container'>
                                <Input
                                    type='text'
                                    className='event-input-title'
                                    size='large'
                                    appearance='underline'
                                    placeholder='Add a title'
                                    value={titleEvent}
                                    onChange={onTitleChange}
                                />
                            </div>
                        </div>
                        <div className='datetime-container'>
                            <Clock24Regular/>
                            <div className='event-input-container-datetime event-input-container'>
                                <div className='datetime-group'>
                                    <Input
                                        type='date'
                                        className='start-date-input'
                                        value={startEventData}
                                        onChange={onStartDateChange}
                                    />
                                    <Select
                                        className='time-selector'
                                        onChange={onStartTimeChange}
                                    >
                                        <TimeSelectItems start={startEventTime}/>
                                    </Select>
                                </div>
                                <div className='datetime-group datetime-group-end'>
                                    <Input
                                        type='date'
                                        className='end-date-input'
                                        value={endEventData}
                                        onChange={onEndDateChange}
                                    />
                                    <Select
                                        className='time-selector'
                                        onChange={onEndTimeChange}
                                    >
                                        <TimeSelectItems end={endEventTime}/>
                                    </Select>
                                </div>

                            </div>
                        </div>
                        <div className='repeat-container'>
                            <Combobox
                                onOptionSelect={repeatOnSelect}
                                selectedOptions={repeatOptionsSelected}
                                value={repeatOption}
                            >
                                <Option
                                    key='empty'
                                    text='empty'
                                >
                                    Don't repeat
                                </Option>
                                <Option
                                    key='custom'
                                    text='custom'
                                >
                                    Custom
                                </Option>
                            </Combobox>
                            <RepeatComponent/>
                        </div>

                        <div className='event-add-users-container'>
                            <PersonAdd24Regular/>
                            <div className='event-input-container'>
                                <div className='event-input-users-wrapper'>
                                    <Combobox
                                        placeholder='Select a user'
                                        onChange={onInputUserAction}
                                        onOptionSelect={(event, data) => {
                                            usersAutocomplete.map((user) => {
                                                if (user.id === data.optionValue) {
                                                    setUsersAdded(usersAdded.concat([user]));
                                                    return;
                                                }
                                            });
                                            setSearchUsersInput('');
                                            setUsersAutocomplete([]);
                                        }}
                                        value={searchUsersInput}
                                    >
                                        {usersAutocomplete.map((user) => {
                                            let stat = 'unknown';
                                            if (UserStatusSelector[user.id] === 'online') {
                                                stat = 'available';
                                            }
                                            return <Option text={user.id}>
                                                <Persona
                                                    name={getDisplayUserName(user)}
                                                    className='user-list-item'
                                                    as='div'
                                                    presence={{status: stat}}
                                                />
                                            </Option>
                                        })}

                                        {usersAutocomplete.length === 0 ? (
                                            <Option key='no-results' text=''>
                                                No results found
                                            </Option>
                                        ) : null}
                                    </Combobox>

                                </div>
                            </div>
                        </div>
                        <div className='users-added-container'>
                            <UsersAddedComponent/>
                        </div>


                        <div className='event-channel-container'>
                            <ChatMultiple24Regular/>
                            <div className='event-channel-input-container'>
                                <div className='event-input-channel-wrapper'>
                                    <Combobox
                                        placeholder='Select a channel'
                                        onChange={onInputChannelAction}
                                        onOptionSelect={onSelectChannelOption}
                                        value={selectedChannelText}
                                    >
                                        {channelsAutocomplete.map((option) => (
                                            <Option
                                                key={option.id}
                                                text={option.id}
                                            >
                                                {option.display_name}
                                            </Option>
                                        ))}

                                        {channelsAutocomplete.length === 0 ? (
                                            <Option key='no-results' text=''>
                                                No results found
                                            </Option>
                                        ) : null}
                                    </Combobox>
                                </div>
                            </div>
                        </div>

                    </DialogContent>
                    <DialogActions>
                        <RemoveEventButton/>
                        <DialogTrigger disableButtonEnhancement>
                            <Button
                                appearance='secondary'
                                onClick={viewEventModalHandleClose}
                            >
                                Close
                            </Button>
                        </DialogTrigger>
                        <Button
                            appearance='primary'
                            onClick={onSaveEvent}
                            icon={<Save16Regular/>}
                        >
                            Save
                        </Button>
                    </DialogActions>
                </DialogBody>
            </DialogSurface>
        </Dialog>
    );
};

export default EventModalComponent;
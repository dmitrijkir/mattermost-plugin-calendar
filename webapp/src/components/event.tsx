import React, {useEffect, useState} from 'react';
import {Client4} from 'mattermost-redux/client';
import {UserProfile} from 'mattermost-redux/types/users';
import {Channel} from 'mattermost-redux/types/channels';
import CalendarRef from './calendar';
import {ApiClient} from 'client';
import {useDispatch, useSelector} from 'react-redux';
import {selectIsOpenEventModal, selectSelectedEvent} from 'selectors';
import {closeEventModal, eventSelected} from 'actions';
import {getCurrentTeam, getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {getUserStatuses} from 'mattermost-redux/selectors/entities/users';
import {getTeammateNameDisplaySetting} from "mattermost-redux/selectors/entities/preferences";
import RepeatEventCustom from './repeat-event';
import {
    ChatMultiple24Regular,
    Circle20Filled,
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
    DialogTitle,
    DialogTrigger,
    Input,
    Option,
    Persona,
    Select,
} from "@fluentui/react-components";
import {format, parse} from "date-fns";
import {InputOnChangeData} from "@fluentui/react-input";

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

    const CurrentTeamId = useSelector(getCurrentTeamId);
    const UserStatusSelector = useSelector(getUserStatuses);

    const dispatch = useDispatch();

    const initialDate = new Date();

    const [usersAutocomplete, setUsersAutocomplete] = useState<UserProfile[]>([]);
    const [usersAdded, setUsersAdded] = useState<UserProfile[]>([]);

    const [searchUsersInput, setSearchUsersInput] = useState('');

    const [selectedColor, setSelectedColor] = useState('#D0D0D0');
    const [selectedColorStyle, setSelectedColorStyle] = useState('event-color-default');

    const [channelsAutocomplete, setChannelsAutocomplete] = useState<Channel[]>([]);
    const [selectedChannel, setSelectedChannel] = useState({});
    const [selectedChannelText, setSelectedChannelText] = useState('');

    const [titleEvent, setTitleEvent] = useState('');
    const [startEventData, setStartEventDate] = useState(initialDate);
    const [endEventData, setEndEventDate] = useState(initialDate);

    const [startEventTime, setStartEventTime] = useState(initialStartTime);
    const [endEventTime, setEndEventTime] = useState(initialEndTime);

    const [repeatRule, setRepeatRule] = useState<string>('');
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

        // repeat state
        setShowCustomRepeat(false);
        setRepeatOptionsSelected(['empty']);
        setRepeatOption('Don\'t repeat');
        setRepeatRule('');

        setSelectedChannel({});
        setUsersAdded([]);
        setSelectedColor('#D0D0D0');
    };

    const onTitleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTitleEvent(event.target.value);
    };

    const onStartDateChange = (event: React.ChangeEvent<HTMLInputElement>, data: InputOnChangeData) => {
        setStartEventDate(parse(data.value, 'yyyy-MM-dd', new Date()));
    };

    const onEndDateChange = (event: React.ChangeEvent<HTMLInputElement>, data: InputOnChangeData) => {
        setEndEventDate(parse(data.value, 'yyyy-MM-dd', new Date()));
    };

    const onStartTimeChange = (event: React.ChangeEvent<HTMLSelectElement>, data: any) => {
        setStartEventTime(data.value);
    };

    const onEndTimeChange = (event: React.ChangeEvent<HTMLSelectElement>, data: any) => {
        setEndEventTime(data.value);
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
        let repeat = '';
        if (repeatOption === 'Custom') {
            repeat = repeatRule;
        }
        if (selectedEvent?.event?.id == null) {
            const response = await ApiClient.createEvent(
                titleEvent,
                format(startEventData, 'yyyy-MM-dd') + 'T' + startEventTime + ':00Z',
                format(endEventData, 'yyyy-MM-dd') + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                repeat,
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
                format(startEventData, 'yyyy-MM-dd') + 'T' + startEventTime + ':00Z',
                format(endEventData, 'yyyy-MM-dd') + 'T' + endEventTime + ':00Z',
                members,
                Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                repeat,
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

    const colorsMap: { [name: string]: string } = {
        '': 'event-color-default',
        'default': 'event-color-default',
        '#F2B3B3': 'event-color-red',
        '#FCECBE': 'event-color-yellow',
        '#B6D9C7': 'event-color-green',
        '#B3E1F7': 'event-color-blue',
    };
    const onSelectColor = (event: SelectionEvents, data: OptionOnSelectData) => {
        setSelectedColor(data.optionValue!);
        setSelectedColorStyle(colorsMap[data.optionValue!]);
    };

    useEffect(() => {
        let mounted = true;
        if (mounted && selectedEvent?.event?.id != null) {
            ApiClient.getEventById(selectedEvent.event.id).then((data) => {
                setTitleEvent(data.data.title);

                const startEventResp: Date = parse(data.data.start, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());
                const endEventResp: Date = parse(data.data.end, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());
                setStartEventDate(startEventResp);
                setEndEventDate(endEventResp);
                setUsersAdded(data.data.attendees);

                setStartEventTime(format(startEventResp, 'HH:mm'));
                setEndEventTime(format(endEventResp, 'HH:mm'));
                setSelectedColor(data.data.color!);
                setSelectedColorStyle(colorsMap[data.data.color!]);

                if (data.data.recurrence.length !== 0) {
                    setRepeatRule(data.data.recurrence);
                    setRepeatOption('Custom');
                    setShowCustomRepeat(true);
                }

                if (data.data.channel != null) {
                    Client4.getChannel(data.data.channel).then((channel) => {
                        setSelectedChannel(channel);
                        setSelectedChannelText(channel.display_name);
                    });
                }
            });
        } else if (mounted && selectedEvent?.event?.id == null && selectedEvent?.event?.start != null) {
            setStartEventDate(selectedEvent?.event.start);
            setEndEventDate(selectedEvent?.event.end);

            setStartEventTime(format(selectedEvent?.event.start, 'HH:mm'));
            setEndEventTime(format(selectedEvent?.event.end, 'HH:mm'));
        }
        mounted = false;

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
            return (<DialogActions position='star'>
                <Button
                    appearance='outline'
                    icon={<Delete16Regular/>}
                    onClick={viewEventModalHandleClose}
                >
                    Remove
                </Button>
            </DialogActions>);

        }
        return <></>
    }


    const RepeatComponent = () => {
        if (showCustomRepeat) {
            return <RepeatEventCustom
                selected={repeatRule}
                onSelect={setRepeatRule}
            />;
        }
        return <></>;
    };

    return (
        <Dialog open={isOpenEventModal}>
            <DialogSurface>
                <DialogBody className='event-modal'>
                    <DialogTitle className='event-modal-title'/>
                    <DialogContent className='modal-container'>
                        <div className='event-color-button'>
                            <Combobox
                                onOptionSelect={onSelectColor}
                                className={`dropdown-color-button ${selectedColorStyle}`}
                                style={{color: selectedColor, borderColor: 'unset'}}
                                defaultSelectedOptions={['default']}
                                expandIcon={<Circle20Filled className={selectedColorStyle}/>}
                                width='50px'
                                listbox={{
                                    className: 'dropdown-color-button-listbox',
                                }}
                            >
                                <Option
                                    key='default'
                                    text='default'
                                    className='event-color-items event-color-default'
                                >
                                    <i className='icon fa fa-circle'/>
                                </Option>
                                <Option
                                    key='default'
                                    text='#F2B3B3'
                                    className='event-color-items event-color-red'
                                >
                                    <i className='icon fa fa-circle'/>
                                </Option>
                                <Option
                                    key='default'
                                    text='#FCECBE'
                                    className='event-color-items event-color-yellow'
                                >
                                    <i className='icon fa fa-circle'/>
                                </Option>
                                <Option
                                    key='default'
                                    text='#B6D9C7'
                                    className='event-color-items event-color-green'
                                >
                                    <i className='icon fa fa-circle'/>
                                </Option>
                                <Option
                                    key='default'
                                    text='#B3E1F7'
                                    className='event-color-items event-color-blue'
                                >
                                    <i className='icon fa fa-circle'/>
                                </Option>
                            </Combobox>
                        </div>
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
                                        value={format(startEventData, 'yyyy-MM-dd')}
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
                                        value={format(endEventData, 'yyyy-MM-dd')}
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
                                                if (user.id === data.optionValue && !usersAdded.some((u) => u.id === data.optionValue)) {
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
                    <RemoveEventButton/>
                    <DialogActions position='end'>
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
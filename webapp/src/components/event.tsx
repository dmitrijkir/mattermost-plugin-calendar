import React, { useEffect, useState } from 'react';
import { Client4 } from 'mattermost-redux/client';
import { UserProfile } from 'mattermost-redux/types/users';
import { Channel } from 'mattermost-redux/types/channels';

import { useDispatch, useSelector } from 'react-redux';

import { getCurrentTeamId, getCurrentTeam } from 'mattermost-redux/selectors/entities/teams';
import { getCurrentUserId, getUserStatuses, makeGetProfilesInChannel } from 'mattermost-redux/selectors/entities/users';
import { getTeammateNameDisplaySetting } from 'mattermost-redux/selectors/entities/preferences';
import { getProfilesInChannel } from 'mattermost-redux/actions/users';

// importing the editor and the plugin from their full paths
import {
    Eye24Regular,
    ChatMultiple24Regular,
    Clock24Regular,
    Delete16Regular,
    Dismiss12Regular,
    Pen24Regular,
    PersonAdd24Regular,
    Save16Regular,
    TextDescription24Regular,
    PeopleTeam24Regular,
    Video24Regular
} from '@fluentui/react-icons';
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
    OptionGroup,
    Persona,
    Skeleton,
    SkeletonItem,
    Spinner,
    Textarea,
    Toolbar,
    ToolbarButton,
    Tag,
    useId,
    Toast,
    ToastIntent,
    ToastTitle,
    useToastController,
    Toaster
} from '@fluentui/react-components';
import { format, parse, set } from 'date-fns';
import { InputOnChangeData } from '@fluentui/react-input';

import roundToNearestMinutes from 'date-fns/roundToNearestMinutes';

import { GlobalState } from 'mattermost-redux/types/store';

import { closeEventModal, eventSelected, updateMembersAddedInEvent, updateSelectedEventTime } from 'actions';
import { getCalendarSettings, getMembersAddedInEvent, getSelectedCalendarType, getSelectedEventTime, selectIsOpenEventModal, selectSelectedEvent } from 'selectors';
import { ApiClient } from 'client';

import RepeatEventCustom from './repeat-event';

import CalendarRef from './calendar';
import TimeSelector from './time-selector';
import PlanningAssistant from './planning-assistant';
import EventAlertSelect from "./alert-input";
import VisibilitySelect from './visibility-input';
import EventTypeSelect from './event-type-input';

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

const initialStartTime = (): string => {
    return format(roundToNearestMinutes(new Date(), {
        nearestTo: 30,
        roundingMethod: 'ceil',
    }), 'HH:mm');
};

const initialEndTime = (): string => {
    const dt = new Date();
    dt.setMinutes(dt.getMinutes() + 30);
    return format(roundToNearestMinutes(dt, {
        nearestTo: 30,
        roundingMethod: 'ceil',
    }), 'HH:mm');
};

const EventModalComponent = () => {
    const selectedEvent = useSelector(selectSelectedEvent);
    const isOpenEventModal = useSelector(selectIsOpenEventModal);

    const displayNameSettings = useSelector(getTeammateNameDisplaySetting);

    const CurrentTeamId = useSelector(getCurrentTeamId);
    const CurrentTeam = useSelector(getCurrentTeam);

    const UserStatusSelector = useSelector(getUserStatuses);
    const currentUserId = useSelector(getCurrentUserId);
    const selectedEventTime = useSelector(getSelectedEventTime);
    const settings = useSelector(getCalendarSettings);
    const selectedCalendarType = useSelector(getSelectedCalendarType);

    const dispatch = useDispatch();

    const initialDate = new Date();

    const [isLoading, setIsLoading] = useState(false);
    const [isSaving, setIsSaving] = useState(false);

    const usersMentionTags: {
        [name: string]: string
    } = {
        '@channel': 'users from channel',
    };
    const [usersAutocomplete, setUsersAutocomplete] = useState<UserProfile[]>([]);

    const [searchUsersInput, setSearchUsersInput] = useState('');

    const [selectedAlert, setSelectedAlert] = useState('');
    const [selectedType, setSelectedType] = useState('call');
    const [meetingLink, setMeetingLink] = useState('');
    const [eventOwner, setEventOwner] = useState('');

    const [channelsAutocomplete, setChannelsAutocomplete] = useState<Channel[]>([]);
    const [selectedChannel, setSelectedChannel] = useState({});
    const [selectedChannelText, setSelectedChannelText] = useState('');

    const [selectedVisibility, setSelectedVisibility] = useState('private')

    const [isPlanningAssistantOpen, setIsPlanningAssistantOpen] = useState(false);
    const inputEventTitleRef = React.useRef<HTMLInputElement>(null);

    const getProfilesInChannelSelector = makeGetProfilesInChannel();
    const profilesInCurrentChannelSelector = (state: GlobalState) => getProfilesInChannelSelector(state, selectedChannel?.id);
    const profilesInChannel = useSelector(profilesInCurrentChannelSelector);

    const usersAddedInEvent = useSelector(getMembersAddedInEvent);

    const [titleEvent, setTitleEvent] = useState('');
    const [descriptionEvent, setDescriptionEvent] = useState('');

    const [repeatRule, setRepeatRule] = useState<string>('');
    const [showCustomRepeat, setShowCustomRepeat] = useState(false);
    const [repeatOption, setRepeatOption] = useState("Don't repeat");
    const [repeatOptionsSelected, setRepeatOptionsSelected] = useState(['empty']);

    const toasterId = useId("toasterEventForm");
    const { dispatchToast } = useToastController(toasterId);

    // methods
    const viewEventModalHandleClose = () => {
        cleanState();
        dispatch(closeEventModal());
        dispatch(eventSelected({}));
    };

    const cleanState = () => {
        setTitleEvent('');
        setDescriptionEvent('');

        setIsSaving(false);
        setIsLoading(false);

        dispatch(updateSelectedEventTime({
            start: initialDate,
            end: initialDate,
            startTime: initialStartTime(),
            endTime: initialEndTime(),
        }));

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
        dispatch(updateMembersAddedInEvent([]));

        // a new event belongs to the calendar the user is currently looking at,
        // otherwise it would be saved out of view
        setSelectedType(selectedCalendarType);
        setMeetingLink('');
        setEventOwner('');

        setSelectedVisibility('private');
        setSelectedAlert('');
    };

    const onTitleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setTitleEvent(event.target.value);
    };

    const onStartDateChange = (event: React.ChangeEvent<HTMLInputElement>, data: InputOnChangeData) => {
        dispatch(updateSelectedEventTime({ start: parse(data.value, 'yyyy-MM-dd', new Date()) }));
    };

    const onEndDateChange = (event: React.ChangeEvent<HTMLInputElement>, data: InputOnChangeData) => {
        dispatch(updateSelectedEventTime({ end: parse(data.value, 'yyyy-MM-dd', new Date()) }));
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
                dispatch(getProfilesInChannel(option.id, 0, 1000));
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

    const showErrorToast = (message: string) => {
        dispatchToast(
            <Toast>
                <ToastTitle></ToastTitle>
                {message}
            </Toast>,
            { intent: 'error' }
        );
    };

    // combines a date with an "HH:mm" time string into a real Date, used only to
    // compare start against end locally.
    const buildDateTime = (date: Date, time: string): Date => {
        const [hours, minutes] = time.split(':').map(Number);
        return set(date, { hours, minutes: minutes || 0, seconds: 0, milliseconds: 0 });
    };

    // The server re-reads the wall-clock fields of whatever timestamp it receives
    // and interprets them in the user's Mattermost timezone, so the picked wall
    // clock has to be sent as-is. Converting to a real UTC instant here would make
    // the server apply the offset a second time.
    const toServerDateTime = (date: Date, time: string): string => {
        return format(date, 'yyyy-MM-dd') + 'T' + time + ':00Z';
    };

    const onSaveEvent = async () => {

        if (selectedVisibility === "channel" && Object.keys(selectedChannel).length === 0) {
            showErrorToast('You selected channel visibility but you didn\'t select a channel');
            return;
        }

        const start = buildDateTime(selectedEventTime.start, selectedEventTime.startTime);
        const end = buildDateTime(selectedEventTime.end, selectedEventTime.endTime);

        if (end.getTime() <= start.getTime()) {
            showErrorToast('End time must be after start time');
            return;
        }

        const members: string[] = usersAddedInEvent.map((user: UserProfile) => user.id);
        let repeat = '';
        if (repeatOption === 'Custom') {
            repeat = repeatRule;
        }
        setIsSaving(true);
        try {
            if (selectedEvent?.event?.id == null) {
                await ApiClient.createEvent(
                    titleEvent,
                    toServerDateTime(selectedEventTime.start, selectedEventTime.startTime),
                    toServerDateTime(selectedEventTime.end, selectedEventTime.endTime),
                    members,
                    descriptionEvent,
                    CurrentTeamId,
                    selectedVisibility,
                    Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                    repeat,

                    // the color comes from the calendar the event belongs to and is
                    // resolved server-side on read, so nothing is stored per event
                    undefined,
                    selectedAlert,
                    selectedType,
                    meetingLink,
                );
            } else {
                await ApiClient.updateEvent(
                    selectedEvent.event.id,
                    titleEvent,
                    toServerDateTime(selectedEventTime.start, selectedEventTime.startTime),
                    toServerDateTime(selectedEventTime.end, selectedEventTime.endTime),
                    members,
                    descriptionEvent,
                    CurrentTeamId,
                    selectedVisibility,
                    Object.keys(selectedChannel).length !== 0 ? selectedChannel.id : null,
                    repeat,

                    // the color comes from the calendar the event belongs to and is
                    // resolved server-side on read, so nothing is stored per event
                    undefined,
                    selectedAlert,
                    selectedType,
                    meetingLink,
                );
            }
            CalendarRef.current?.getApi().getEventSources()[0].refetch();
            cleanState();
            viewEventModalHandleClose();
        } catch (e) {
            showErrorToast('Failed to save event, please try again');
        } finally {
            setIsSaving(false);
        }
    };

    const onRemoveEvent = async () => {
        if (repeatRule !== '' && !window.confirm('This is a recurring event. Removing it will remove the entire series. Continue?')) {
            return;
        }
        try {
            await ApiClient.removeEvent(selectedEvent.event.id);
            CalendarRef.current?.getApi().getEventSources()[0].refetch();
            cleanState();
            viewEventModalHandleClose();
        } catch (e) {
            showErrorToast('Failed to remove event, please try again');
        }
    };

    const onGenerateMeetingLink = () => {
        const roomSlug = `mm-calendar-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
        setMeetingLink(`${settings.jitsiBaseUrl}/${roomSlug}`);
    };

    useEffect(() => {
        if (isOpenEventModal && selectedEvent?.event?.id == null) {
            setSelectedType(selectedCalendarType);
        }
    }, [isOpenEventModal]);

    useEffect(() => {
        let cancelled = false;

        if (selectedEvent?.event?.id != null) {
            setIsLoading(true);
            ApiClient.getEventById(selectedEvent.event.id).then((data) => {
                if (cancelled) {
                    return;
                }
                setTitleEvent(data.data.title);
                setDescriptionEvent(data.data.description);

                const startEventResp: Date = parse(data.data.start, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());
                const endEventResp: Date = parse(data.data.end, "yyyy-MM-dd'T'HH:mm:ssxxx", new Date());
                dispatch(updateSelectedEventTime({
                    start: startEventResp,
                    end: endEventResp,
                    startTime: format(startEventResp, 'HH:mm'),
                    endTime: format(endEventResp, 'HH:mm'),
                }));
                dispatch(updateMembersAddedInEvent(data.data.attendees));

                setSelectedType(data.data.type || 'call');
                setMeetingLink(data.data.meetingLink || '');
                setEventOwner(data.data.owner);
                setSelectedVisibility(data.data.visibility);
                setSelectedAlert(data.data.alert);

                if (data.data.recurrence.length !== 0) {
                    setRepeatRule(data.data.recurrence);
                    setRepeatOption('Custom');
                    setShowCustomRepeat(true);
                }

                if (data.data.channel != null) {
                    Client4.getChannel(data.data.channel).then((channel: Channel) => {
                        if (!cancelled) {
                            setSelectedChannel(channel);
                            setSelectedChannelText(channel.display_name);
                        }
                    });
                }
                setIsLoading(false);
            }).catch(() => {
                if (!cancelled) {
                    setIsLoading(false);
                    showErrorToast('Failed to load event, please try again');
                }
            });
        } else if (selectedEvent?.event?.id == null && selectedEvent?.event?.start != null) {
            dispatch(updateSelectedEventTime({
                start: selectedEvent?.event.start,
                end: selectedEvent?.event.end,
                startTime: format(selectedEvent?.event.start, 'HH:mm'),
                endTime: format(selectedEvent?.event.end, 'HH:mm'),
            }));
        }

        return () => {
            cancelled = true;
        };
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

    const AddedUserComponent = (props: AddedUserComponentProps) => {
        let stat = 'unknown';
        if (UserStatusSelector[props.user.id] === 'online') {
            stat = 'available';
        }

        return (<span className='added-user-badge-container'>
            <Persona
                name={getDisplayUserName(props.user)}
                avatar={{ color: 'colorful' }}
                presence={{ status: stat }}
            />
            <Dismiss12Regular
                className='added-user-badge-icon-container'
                onClick={() => {
                    dispatch(updateMembersAddedInEvent(usersAddedInEvent.filter((item: UserProfile) => item.id !== props.user.id)));
                }}
            />

        </span>);
    };

    const UsersAddedComponent = () => {
        if (usersAddedInEvent.length > 0) {
            return (<div className='added-users-list'>
                {
                    usersAddedInEvent.map((user: UserProfile) => {
                        return <AddedUserComponent user={user} />;
                    })
                }
            </div>);
        }
        return <></>;
    };

    const RemoveEventButton = () => {
        // removing drops the event for every attendee, so the server only lets
        // the owner do it — don't offer the button to anyone else
        if (selectedEvent?.event?.id != null && eventOwner === currentUserId) {
            return (<DialogActions position='star'>
                <Button
                    appearance='outline'
                    icon={<Delete16Regular />}
                    onClick={onRemoveEvent}
                >
                    {'Remove'}
                </Button>
            </DialogActions>);
        }
        return <></>;
    };

    const RepeatComponent = () => {
        if (showCustomRepeat) {
            return (
                <RepeatEventCustom
                    selected={repeatRule}
                    onSelect={setRepeatRule}
                />
            );
        }
        return <></>;
    };

    return (
        <div>
            {
                usersAddedInEvent.length > 0 ? <PlanningAssistant
                    open={isPlanningAssistantOpen}
                    onOpenChange={(ev, data) => {
                        setIsPlanningAssistantOpen(data.open);
                        inputEventTitleRef.current?.focus();
                    }}
                /> : null
            }
            <Dialog open={isOpenEventModal}>
                <DialogSurface>
                    <DialogBody className='event-modal'>
                        <DialogTitle className='event-modal-title' />
                        <DialogContent className='modal-container'>
                            <div className='title-toolbar'>
                                <Toolbar aria-label='Default'>
                                    <ToolbarButton
                                        aria-label='planning assistant'
                                        onClick={() => setIsPlanningAssistantOpen(true)}
                                        disabled={usersAddedInEvent.length === 0}
                                    >
                                        planning assistant
                                    </ToolbarButton>
                                </Toolbar>
                            </div>
                            <div className='event-title-container'>
                                <Pen24Regular />
                                <div className='event-input-container'>
                                    {isLoading ? (<Skeleton className='event-input-title'>
                                        <SkeletonItem />
                                    </Skeleton>) : (<Input
                                        ref={inputEventTitleRef}
                                        type='text'
                                        className='event-input-title'
                                        size='large'
                                        appearance='underline'
                                        placeholder='Add a title'
                                        value={titleEvent}
                                        onChange={onTitleChange}
                                    />)}

                                </div>
                            </div>
                            <div className='datetime-container'>
                                <Clock24Regular />
                                <div className='event-input-container-datetime event-input-container'>
                                    <div className='datetime-group'>
                                        {isLoading ? (<Skeleton className='start-date-input'>
                                            <SkeletonItem />
                                        </Skeleton>) : (<Input
                                            type='date'
                                            className='start-date-input'
                                            value={format(selectedEventTime?.start, 'yyyy-MM-dd')}
                                            onChange={onStartDateChange}
                                        />)}

                                        {isLoading ? (<Skeleton className='start-date-input'>
                                            <SkeletonItem />
                                        </Skeleton>) : (<TimeSelector
                                            selected={selectedEventTime.startTime}
                                            onSelect={(value) => dispatch(updateSelectedEventTime({ startTime: value }))}
                                        />)}

                                    </div>
                                    <div className='datetime-group datetime-group-end'>
                                        {isLoading ? (
                                            <Skeleton className='end-date-input'>
                                                <SkeletonItem />
                                            </Skeleton>
                                        ) :
                                            (<Input
                                                type='date'
                                                className='end-date-input'
                                                value={format(selectedEventTime?.end, 'yyyy-MM-dd')}
                                                onChange={onEndDateChange}
                                            />)}
                                        {isLoading ? (<Skeleton className='end-date-input'>
                                            <SkeletonItem />
                                        </Skeleton>) : (<TimeSelector
                                            selected={selectedEventTime.endTime}
                                            onSelect={(value) => dispatch(updateSelectedEventTime({ endTime: value }))}
                                        />)}

                                    </div>

                                </div>
                            </div>
                            <div className='repeat-container'>
                                {isLoading ? (<Skeleton className='skeleton-dropdown'>
                                    <SkeletonItem />
                                </Skeleton>) :
                                    (
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
                                    )}
                                <RepeatComponent />
                            </div>

                            <div className='event-channel-container'>
                                <ChatMultiple24Regular />
                                <div className='event-channel-input-container'>
                                    <div className='event-input-channel-wrapper'>
                                        {isLoading ? (
                                            <Skeleton className='skeleton-dropdown'>
                                                <SkeletonItem />
                                            </Skeleton>
                                        ) : (
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
                                                    <Option
                                                        key='no-results'
                                                        text=''
                                                    >
                                                        No results found
                                                    </Option>
                                                ) : null}
                                            </Combobox>
                                        )}
                                    </div>
                                </div>
                                {CurrentTeam ? (<div className="current-team-tag">
                                    <Tag icon={<PeopleTeam24Regular />}>{CurrentTeam.display_name}</Tag>
                                </div>) : null}
                            </div>

                            {
                                isLoading ?
                                    <Skeleton className='skeleton-dropdown'>
                                        <SkeletonItem />
                                    </Skeleton>
                                    :
                                    <EventTypeSelect
                                        selected={selectedType}
                                        onSelected={(selected) => setSelectedType(selected)}
                                    />
                            }

                            {
                                isLoading ?
                                    <Skeleton className='skeleton-dropdown'>
                                        <SkeletonItem />
                                    </Skeleton>
                                    :
                                    <VisibilitySelect
                                        selected={selectedVisibility}
                                        onSelected={(selected) => setSelectedVisibility(selected)}
                                    />
                            }

                            {
                                isLoading ?
                                    <Skeleton className='skeleton-dropdown'>
                                        <SkeletonItem />
                                    </Skeleton>
                                    :
                                    <EventAlertSelect
                                        selected={selectedAlert}
                                        onSelected={(selected) => setSelectedAlert(selected)}
                                    />
                            }



                            <div className='event-add-users-container'>
                                <PersonAdd24Regular />
                                <div className='event-input-container'>
                                    <div className='event-input-users-wrapper'>
                                        {isLoading ? (<Skeleton className='skeleton-dropdown'>
                                            <SkeletonItem />
                                        </Skeleton>) : (<Combobox
                                            placeholder='Select a user'
                                            checked={false}
                                            selectedOptions={[]}
                                            onChange={onInputUserAction}
                                            onOptionSelect={(event, data) => {
                                                if (data.optionValue in usersMentionTags) {
                                                    dispatch(updateMembersAddedInEvent(profilesInChannel));
                                                }
                                                usersAutocomplete.map((user) => {
                                                    if (user.id === data.optionValue && !usersAddedInEvent.some((u) => u.id === data.optionValue)) {
                                                        dispatch(updateMembersAddedInEvent([...usersAddedInEvent, user]));
                                                    }
                                                });
                                                setSearchUsersInput('');
                                                setUsersAutocomplete([]);
                                            }}
                                            value={searchUsersInput}
                                        >
                                            <OptionGroup label='USERS'>

                                                {usersAutocomplete.map((user) => {
                                                    let stat = 'unknown';
                                                    if (UserStatusSelector[user.id] === 'online') {
                                                        stat = 'available';
                                                    }
                                                    return (<Option text={user.id}>
                                                        <Persona
                                                            name={getDisplayUserName(user)}
                                                            className='user-list-item'
                                                            as='div'
                                                            presence={{ status: stat }}
                                                        />
                                                    </Option>);
                                                })}

                                                {usersAutocomplete.length === 0 ? (
                                                    <Option
                                                        key='no-results'
                                                        text=''
                                                    >
                                                        No results found
                                                    </Option>
                                                ) : null}
                                            </OptionGroup>
                                            <OptionGroup label='SPECIAL'>
                                                {
                                                    Object.entries(usersMentionTags).map(([key, value]) => {
                                                        return (<Option
                                                            key={key}
                                                            text={key}
                                                        >
                                                            {value}
                                                        </Option>);
                                                    })
                                                }
                                            </OptionGroup>
                                        </Combobox>)}

                                    </div>
                                </div>
                            </div>
                            <div className='users-added-container'>
                                <UsersAddedComponent />
                            </div>

                            <div className='event-description-container'>
                                <TextDescription24Regular />
                                <div className='event-description-input-container'>
                                    {isLoading ? (<Skeleton className='event-description-input-textarea'><SkeletonItem /></Skeleton>) :
                                        <Textarea
                                            placeholder='Add description'
                                            className='event-description-input-textarea'
                                            resize='vertical'
                                            value={descriptionEvent}
                                            onChange={(event, data) => setDescriptionEvent(data.value)}
                                        />}
                                </div>

                            </div>

                            <div className='event-meeting-link-container'>
                                <Video24Regular />
                                <div className='event-meeting-link-input-container'>
                                    {isLoading ? (<Skeleton className='skeleton-dropdown'><SkeletonItem /></Skeleton>) : (
                                        <>
                                            <Input
                                                readOnly={true}
                                                type='text'
                                                className='event-meeting-link-input'
                                                placeholder='No meeting link'
                                                value={meetingLink}
                                            />
                                            <Button
                                                appearance='subtle'
                                                onClick={onGenerateMeetingLink}
                                            >
                                                {'Generate Jitsi link'}
                                            </Button>
                                        </>
                                    )}
                                </div>
                            </div>
                            <Toaster toasterId={toasterId} />
                        </DialogContent>
                        <RemoveEventButton />
                        <DialogActions position='end'>
                            <DialogTrigger disableButtonEnhancement={true}>
                                <Button
                                    appearance='secondary'
                                    onClick={viewEventModalHandleClose}
                                >
                                    {'Close'}
                                </Button>
                            </DialogTrigger>

                            <Button
                                appearance='primary'
                                onClick={onSaveEvent}
                                icon={isSaving ? (<Spinner size='tiny' />) : (<Save16Regular />)}
                                disabled={isSaving}
                            >
                                {'Save'}
                            </Button>

                        </DialogActions>
                    </DialogBody>
                </DialogSurface>
            </Dialog>
        </div>

    );
};

export default EventModalComponent;
// 24px to match every other row icon in the event form; a 20px one shifts the
// field next to it by 4px
import {CalendarLtr24Regular} from '@fluentui/react-icons';
import {Combobox, InfoLabel, Option} from '@fluentui/react-components';
import React from 'react';

interface EventTypeSelectProps {
    selected: string;
    onSelected: (selected: string) => void;
}

interface EventTypeOption {
    id: string;
    display_name: string;
}

const eventTypeOptions: Array<EventTypeOption> = [
    {id: 'call', display_name: 'Call'},
    {id: 'event', display_name: 'Event'},
];

const EventTypeSelect = (props: EventTypeSelectProps) => {
    const getSelectedOptionName = (selected: string) => {
        if (!selected) {
            return '';
        }
        return eventTypeOptions.find((option) => option.id === selected)?.display_name;
    };

    return (
        <div className='event-visibility-container'>
            <CalendarLtr24Regular/>
            <div className='event-visibility-input-container'>
                <div className='event-input-visibility-wrapper'>
                    <Combobox
                        placeholder='Select a calendar'
                        onOptionSelect={(event, data) => {
                            props.onSelected(data.optionValue!);
                        }}
                        value={getSelectedOptionName(props.selected)}
                    >
                        {eventTypeOptions.map((option) => (
                            <Option
                                key={option.id}
                                text={option.id}
                            >
                                {getSelectedOptionName(option.id)}
                            </Option>
                        ))}
                    </Combobox>
                </div>
            </div>
            <InfoLabel
                info={
                    <>
                        <p>Which calendar this belongs to. Calls and events are filtered separately in the toolbar and get their own colour in settings.</p>
                        <p>Only calls carry a video meeting link; events default to all day.</p>
                    </>
                }
            />
        </div>
    );
};

export default EventTypeSelect;

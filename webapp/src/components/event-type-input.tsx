import {CalendarLtr20Regular} from '@fluentui/react-icons';
import {Combobox, Option} from '@fluentui/react-components';
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
    {id: 'event', display_name: 'Meeting event'},
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
            <CalendarLtr20Regular/>
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
        </div>
    );
};

export default EventTypeSelect;

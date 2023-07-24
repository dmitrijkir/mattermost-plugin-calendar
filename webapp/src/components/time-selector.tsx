import {Combobox, ComboboxProps, Option} from '@fluentui/react-components';
import {format} from 'date-fns';
import * as React from 'react';

type SelectTimeOnChange = (time: string) => void;

interface TimeSelectorComponentProps {
    selected: string
    onSelect?: SelectTimeOnChange | undefined
}

const TimeSelector = (props: TimeSelectorComponentProps) => {
    const timeList = [];
    const initialDate = new Date();
    initialDate.setHours(0, 0, 0, 0);
    for (let i = 0; i < 48; i++) {
        timeList.push(new Date(initialDate.valueOf()));
        initialDate.setMinutes(initialDate.getMinutes() + 30);
    }

    const onChange: ComboboxProps['onChange'] = (event) => {
        if (props.onSelect) {
            const value = event.target.value.trim();
            props.onSelect(value);
        }
    };

    const onOptionSelect: ComboboxProps['onOptionSelect'] = (event, data) => {
        if (data.optionText && props.onSelect) {
            props.onSelect(data.optionText);
        }
    };

    return (<div>
        <Combobox
            value={props.selected}
            freeform={true}
            onOptionSelect={onOptionSelect}
            onChange={onChange}
            className='time-selector'
        >
            {timeList.map((option) => (
                <Option key={format(option, 'HH:mm')}>
                    {format(option, 'HH:mm')}
                </Option>
            ))}
        </Combobox>
    </div>);
};

export default TimeSelector;
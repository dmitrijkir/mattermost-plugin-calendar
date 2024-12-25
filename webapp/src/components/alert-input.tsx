import {Alert24Regular} from "@fluentui/react-icons";
import {Combobox, Option, SelectOnChangeData} from "@fluentui/react-components";
import React from "react";

interface EventAlertSelectProps {
    selected: string;
    onSelected: (selected: string) => void;
}

interface SelectOptionOnChangeData {
    optionValue: string;
}

const EventAlertSelect = (props: EventAlertSelectProps) => {
    const alertMapping: Record<string, string> = {
        '': '',
        '5_minutes': '5 minutes before',
        '15_minutes': '15 minutes before',
        '30_minutes': '30 minutes before',
        '1_hour': '1 hour before',
        '2_hours': '2 hours before',
        '1_day': '1 day before',
        '2_days': '2 days before',
        '1_week': '1 week before',
    };

    const availableAlerts = Object.keys(alertMapping);

    return (
        <div className="event-visibility-container">
            <Alert24Regular />
            <div className="event-channel-input-container">
                <div className="event-input-channel-wrapper">
                    <Combobox
                        placeholder="Select alert"
                        value={alertMapping[props.selected]}
                        onOptionSelect={(event, data: SelectOptionOnChangeData) => {
                            props.onSelected(data.optionValue); // update the selected value
                        }}
                    >
                        {availableAlerts.map((option) => (
                            <Option key={option} value={option}>
                                {alertMapping[option]}
                            </Option>
                        ))}
                    </Combobox>
                </div>
            </div>
        </div>
    );
};

export default EventAlertSelect;
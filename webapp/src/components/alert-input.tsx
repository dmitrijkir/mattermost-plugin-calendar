import {Alert24Regular} from "@fluentui/react-icons";
import {Combobox, Option} from "@fluentui/react-components";
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
        '': 'Not set',
        '5_minutes_before': '5 minutes before',
        '15_minutes_before': '15 minutes before',
        '30_minutes_before': '30 minutes before',
        '1_hour_before': '1 hour before',
        '2_hours_before': '2 hours before',
        '1_day_before': '1 day before',
        '2_days_before': '2 days before',
        '1_week_before': '1 week before',
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
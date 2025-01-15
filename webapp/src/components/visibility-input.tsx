import { Eye24Regular } from "@fluentui/react-icons";
import { Combobox, Option, InfoLabel } from "@fluentui/react-components";
import React from "react";


interface VisibilitySelectProps {
    selected: string;
    onSelected: (selected: string) => void;
}

interface VisibilityOption {
    id: string;
    display_name: string;
}

const VisibilitySelect = (props: VisibilitySelectProps) => {
    const visibilityOptions: Array<VisibilityOption> = [
        { id: 'private', display_name: 'Private' },
        { id: 'channel', display_name: 'Channel' },
        { id: 'team', display_name: 'Team' },
    ];

    const getSelectedOptionaName = (selected: string) => {
        if (!selected) {
            return '';
        }
        return visibilityOptions.find(option => option.id === selected)?.display_name;
    };

    return (
        <div className='event-visibility-container'>
            <Eye24Regular />
            <div className='event-visibility-input-container'>
                <div className='event-input-visibility-wrapper'>
                    <Combobox
                        placeholder='Select a visibility'
                        onOptionSelect={(event, data) => {
                            props.onSelected(data.optionValue!);
                        }}
                        value={getSelectedOptionaName(props.selected)}
                    >
                        {visibilityOptions.map((option) => (
                            <Option
                                key={option.id}
                                text={option.id}
                            >
                                {getSelectedOptionaName(option.id)}
                            </Option>
                        ))}
                    </Combobox>
                </div>
            </div>
            <InfoLabel
                info={
                    <>
                        <p>Private: Only invited users can see the event</p>
                        <p>Channel: All members of the channel can see the event</p>
                        <p>Team: All members of the team can see the event</p>
                    </>
                }
            />
        </div>
    );
};

export default VisibilitySelect;
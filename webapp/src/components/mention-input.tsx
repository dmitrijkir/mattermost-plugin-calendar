import {Megaphone24Regular} from '@fluentui/react-icons';
import {Combobox, Option, InfoLabel} from '@fluentui/react-components';
import React from 'react';

interface MentionSelectProps {
    selected: string;
    onSelected: (selected: string) => void;
}

interface MentionOption {
    id: string;
    display_name: string;
}

// Values match the server's EventMention tokens, which are stored without the
// leading "@" — the reminder post adds it.
const mentionOptions: Array<MentionOption> = [
    {id: '', display_name: 'No mention'},
    {id: 'here', display_name: '@here — online members'},
    {id: 'channel', display_name: '@channel — everyone in the channel'},
    {id: 'all', display_name: '@all — everyone in the channel'},
];

const MentionSelect = (props: MentionSelectProps) => {
    const getSelectedOptionName = (selected: string) => {
        return mentionOptions.find((option) => option.id === selected)?.display_name || '';
    };

    return (
        <div className='event-visibility-container'>
            <Megaphone24Regular/>
            <div className='event-visibility-input-container'>
                <div className='event-input-visibility-wrapper'>
                    <Combobox
                        placeholder='Select a mention'
                        onOptionSelect={(event, data) => {
                            props.onSelected(data.optionValue ?? '');
                        }}
                        value={getSelectedOptionName(props.selected)}
                    >
                        {mentionOptions.map((option) => (
                            <Option
                                key={option.id || 'none'}
                                value={option.id}
                                text={option.display_name}
                            >
                                {option.display_name}
                            </Option>
                        ))}
                    </Combobox>
                </div>
            </div>
            <InfoLabel
                info={
                    <>
                        <p>Added to the reminder post so Mattermost actually notifies people.</p>
                        <p>Only applies to events that post to a channel.</p>
                    </>
                }
            />
        </div>
    );
};

export default MentionSelect;

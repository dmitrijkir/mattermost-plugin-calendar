import React, {useEffect, useState} from 'react';

import {Button, Input, Spinner, Text} from '@fluentui/react-components';
import {
    CalendarSync20Regular,
    Copy20Regular,
    Checkmark20Regular,
    ArrowSync20Regular,
    Delete20Regular,
} from '@fluentui/react-icons';

import {ApiClient, ICalTokenResponse} from '../client';

const ICalSettings = () => {
    const [loading, setLoading] = useState<boolean>(true);
    const [tokenData, setTokenData] = useState<ICalTokenResponse | null>(null);
    const [copiedIcal, setCopiedIcal] = useState<boolean>(false);
    const [copiedCaldav, setCopiedCaldav] = useState<boolean>(false);
    const [actionLoading, setActionLoading] = useState<boolean>(false);

    useEffect(() => {
        loadToken();
    }, []);

    const loadToken = async () => {
        setLoading(true);
        try {
            const data = await ApiClient.getICalToken();
            setTokenData(data);
        } catch (error) {
            console.error('Failed to load iCal token:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleEnable = async () => {
        setActionLoading(true);
        try {
            const data = await ApiClient.generateICalToken();
            setTokenData(data);
        } catch (error) {
            console.error('Failed to generate iCal token:', error);
        } finally {
            setActionLoading(false);
        }
    };

    const handleRegenerate = async () => {
        setActionLoading(true);
        try {
            const data = await ApiClient.generateICalToken();
            setTokenData(data);
        } catch (error) {
            console.error('Failed to regenerate iCal token:', error);
        } finally {
            setActionLoading(false);
        }
    };

    const handleRevoke = async () => {
        setActionLoading(true);
        try {
            const data = await ApiClient.revokeICalToken();
            setTokenData(data);
        } catch (error) {
            console.error('Failed to revoke iCal token:', error);
        } finally {
            setActionLoading(false);
        }
    };

    const handleCopyIcal = async () => {
        if (tokenData?.url) {
            try {
                await navigator.clipboard.writeText(tokenData.url);
                setCopiedIcal(true);
                setTimeout(() => setCopiedIcal(false), 2000);
            } catch (error) {
                console.error('Failed to copy iCal URL:', error);
            }
        }
    };

    const handleCopyCaldav = async () => {
        if (tokenData?.caldavUrl) {
            try {
                await navigator.clipboard.writeText(tokenData.caldavUrl);
                setCopiedCaldav(true);
                setTimeout(() => setCopiedCaldav(false), 2000);
            } catch (error) {
                console.error('Failed to copy CalDAV URL:', error);
            }
        }
    };

    if (loading) {
        return (
            <div className='ical-settings-container'>
                <Spinner size='small'/>
            </div>
        );
    }

    return (
        <div className='ical-settings-container'>
            <div className='ical-settings-header'>
                <CalendarSync20Regular/>
                <Text
                    weight='semibold'
                    size={400}
                >
                    {'iCal Feed'}
                </Text>
            </div>
            <Text size={200}>
                {'Subscribe to your calendar from external apps like Apple Calendar, Google Calendar, or Outlook. Use iCal for read-only access or CalDAV for two-way sync.'}
            </Text>

            {tokenData?.enabled ? (
                <div className='ical-settings-enabled'>
                    <Text
                        size={200}
                        weight='semibold'
                    >
                        {'iCal Feed URL (read-only):'}
                    </Text>
                    <div className='ical-url-container'>
                        <Input
                            readOnly={true}
                            value={tokenData.url || ''}
                            className='ical-url-input'
                        />
                        <Button
                            appearance='subtle'
                            icon={copiedIcal ? <Checkmark20Regular/> : <Copy20Regular/>}
                            onClick={handleCopyIcal}
                            title='Copy iCal URL'
                        />
                    </div>
                    <Text
                        size={200}
                        weight='semibold'
                        style={{marginTop: '12px'}}
                    >
                        {'CalDAV URL (read/write sync):'}
                    </Text>
                    <div className='ical-url-container'>
                        <Input
                            readOnly={true}
                            value={tokenData.caldavUrl || ''}
                            className='ical-url-input'
                        />
                        <Button
                            appearance='subtle'
                            icon={copiedCaldav ? <Checkmark20Regular/> : <Copy20Regular/>}
                            onClick={handleCopyCaldav}
                            title='Copy CalDAV URL'
                        />
                    </div>
                    <div className='ical-actions'>
                        <Button
                            appearance='subtle'
                            icon={<ArrowSync20Regular/>}
                            onClick={handleRegenerate}
                            disabled={actionLoading}
                        >
                            {'Regenerate'}
                        </Button>
                        <Button
                            appearance='subtle'
                            icon={<Delete20Regular/>}
                            onClick={handleRevoke}
                            disabled={actionLoading}
                        >
                            {'Disable'}
                        </Button>
                    </div>
                </div>
            ) : (
                <Button
                    appearance='primary'
                    icon={<CalendarSync20Regular/>}
                    onClick={handleEnable}
                    disabled={actionLoading}
                >
                    {'Enable iCal Feed'}
                </Button>
            )}
        </div>
    );
};

export default ICalSettings;

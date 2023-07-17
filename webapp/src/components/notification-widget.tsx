import {useSelector} from 'react-redux';

import {Link, Toast, Toaster, ToastTitle, ToastTrigger, useId, useToastController} from '@fluentui/react-components';

import {useEffect} from 'react';

import {getEventNotification} from '../selectors';
import {CalendarEventNotification} from '../types/event';

const NotificationWidget = () => {
    const notification: CalendarEventNotification = useSelector(getEventNotification);
    const toasterId = useId('toaster');
    const {dispatchToast} = useToastController(toasterId);
    const notify = () =>
        dispatchToast(
            <Toast>
                <ToastTitle
                    action={
                        <ToastTrigger>
                            <Link>Dismiss</Link>
                        </ToastTrigger>
                    }
                >{notification.title}</ToastTitle>
            </Toast>,
            {position: 'top-end', intent: 'success'},
        );

    useEffect(() => {
        if (notification.id === undefined) {
            return;
        }
        notify();
    }, [notification]);

    return (
        <div>
            <Toaster toasterId={toasterId}/>
        </div>

    );
};

export default NotificationWidget;
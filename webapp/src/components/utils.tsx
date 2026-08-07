import React from 'react';

type WindowObject = {
    location: {
        origin: string;
        protocol: string;
        hostname: string;
        port: string;
    };
    basename?: string;
}

function getSiteURLFromWindowObject(obj: WindowObject): string {
    let siteURL = '';
    if (obj.location.origin) {
        siteURL = obj.location.origin;
    } else {
        siteURL = obj.location.protocol + '//' + obj.location.hostname + (obj.location.port ? ':' + obj.location.port : '');
    }

    if (siteURL[siteURL.length - 1] === '/') {
        siteURL = siteURL.substring(0, siteURL.length - 1);
    }

    if (obj.basename) {
        siteURL += obj.basename;
    }

    if (siteURL[siteURL.length - 1] === '/') {
        siteURL = siteURL.substring(0, siteURL.length - 1);
    }

    return siteURL;
}

function getSiteURL(): string {
    return getSiteURLFromWindowObject(window);
}

// Safari in private mode throws on localStorage access rather than just
// returning null, and an exception here would take the whole plugin down at
// import time — a remembered view is never worth that.
export const readStoredPreference = (key: string): string | null => {
    try {
        return window.localStorage.getItem(key);
    } catch (e) {
        return null;
    }
};

export const writeStoredPreference = (key: string, value: string): void => {
    try {
        window.localStorage.setItem(key, value);
    } catch (e) {
        // preference just won't stick, nothing else to do about it
    }
};

// kept in sync with the @media query in style.css
export const MOBILE_BREAKPOINT = 768;

export const isMobileViewport = (): boolean => {
    return typeof window !== 'undefined' && window.innerWidth <= MOBILE_BREAKPOINT;
};

// a seven column week grid is unreadable on a phone, so narrow viewports start
// on the single day view instead
export const getDefaultCalendarView = (): string => {
    return isMobileViewport() ? 'timeGridDay' : 'timeGridWeek';
};

export const STORAGE_KEY_CALENDAR_VIEW = 'calendarPlugin.selectedView';
export const STORAGE_KEY_CALENDAR_TYPE = 'calendarPlugin.selectedType';

// Anything can end up in localStorage — a stale key from an older build, or a
// hand-edited value. Feeding FullCalendar an unknown view name breaks the grid,
// so only known values are ever restored.
const CALENDAR_VIEWS = ['timeGridDay', 'timeGridWeek', 'dayGridMonth'];
const CALENDAR_TYPES = ['call', 'event', 'all'];

export const getInitialCalendarView = (): string => {
    const stored = readStoredPreference(STORAGE_KEY_CALENDAR_VIEW);
    return stored && CALENDAR_VIEWS.includes(stored) ? stored : getDefaultCalendarView();
};

export const getInitialCalendarType = (): string => {
    const stored = readStoredPreference(STORAGE_KEY_CALENDAR_TYPE);
    return stored && CALENDAR_TYPES.includes(stored) ? stored : 'call';
};

export default getSiteURL;
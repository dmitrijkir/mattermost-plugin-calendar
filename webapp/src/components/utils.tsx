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

export default getSiteURL;
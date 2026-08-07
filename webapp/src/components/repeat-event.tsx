import React, {useEffect, useRef, useState} from 'react';
import {
    Popover,
    PopoverSurface,
    PopoverTrigger,
    Select,
    SpinButton,
    SpinButtonChangeEvent,
    SpinButtonOnChangeData,
    Text,
    ToggleButton,
} from '@fluentui/react-components';
import {useBoolean} from '@fluentui/react-hooks';
import {format} from 'date-fns';
import {Calendar, DayOfWeek} from '@fluentui/react';
import {OnOpenChangeData, OpenPopoverEvents} from '@fluentui/react-popover';

type RepeatRuleOnChange = (rule: string) => void;

interface RepeatEventComponentProps {
    selected: string
    onSelect?: RepeatRuleOnChange | undefined
}

const RepeatFreq = {
    Monthly: 'MONTHLY',
    Weekly: 'WEEKLY',
};

const RepeatByWeekDay = {
    0: 'MO',
    1: 'TU',
    2: 'WE',
    3: 'TH',
    4: 'FR',
    5: 'SA',
    6: 'SU',
};

// RRULE UNTIL must be a real UTC instant (RFC5545). We treat the picked day as
// "include everything through the end of that local day" and convert it to UTC,
// instead of stamping local midnight and mislabeling it with a literal 'Z'.
const pad = (n: number): string => String(n).padStart(2, '0');

const toRruleUntilUtc = (date: Date): string => {
    const endOfDayLocal = new Date(date);
    endOfDayLocal.setHours(23, 59, 59, 0);
    return `${endOfDayLocal.getUTCFullYear()}${pad(endOfDayLocal.getUTCMonth() + 1)}${pad(endOfDayLocal.getUTCDate())}` +
        `T${pad(endOfDayLocal.getUTCHours())}${pad(endOfDayLocal.getUTCMinutes())}${pad(endOfDayLocal.getUTCSeconds())}Z`;
};

const fromRruleUntilUtc = (value: string): Date => {
    const year = Number(value.slice(0, 4));
    const month = Number(value.slice(4, 6)) - 1;
    const day = Number(value.slice(6, 8));
    const hours = Number(value.slice(9, 11));
    const minutes = Number(value.slice(11, 13));
    const seconds = Number(value.slice(13, 15));
    return new Date(Date.UTC(year, month, day, hours, minutes, seconds));
};

const RepeatEventCustom = (props: RepeatEventComponentProps) => {
    const [repeatType, setRepeatType] = useState(RepeatFreq.Weekly);
    const [selectedDays, setSelectedDays] = useState([]);
    const [repeatEvery, setRepeatEvery] = useState(1);
    const didMount = useRef(false);
    const [selectedUntil, setSelectedUntil] = useState<Date>();

    const [isShowCalendar, {
        toggle: toggleShowCalendar,
        setFalse: hideCalendar,
        setTrue: showCalendar,
    }] = useBoolean(false);
    const firstDayOfWeek = DayOfWeek.Monday;

    const buildRruleString = () => {
        let rruleString = `RRULE:FREQ=${repeatType};INTERVAL=${repeatEvery};`;
        if (repeatType === RepeatFreq.Weekly && selectedDays.length !== 0) {
            rruleString += `BYDAY=${selectedDays.join(',')};`;
        }
        if (selectedUntil) {
            rruleString += `UNTIL=${toRruleUntilUtc(selectedUntil)};`;
        }

        if (rruleString.endsWith(';')) {
            rruleString = rruleString.slice(0, -1);
        }
        return rruleString;
    };

    const onChangeTrigger = () => {
        if (props.onSelect === undefined) {
            return;
        }
        props.onSelect(buildRruleString());
    };

    const onDaySelected = (day: number) => {
        const selectedDay = RepeatByWeekDay[day];
        if (selectedDays.includes(selectedDay)) {
            setSelectedDays(selectedDays.filter((day) => day !== selectedDay));
        } else {
            setSelectedDays([...selectedDays, selectedDay]);
        }
    };

    // show active selected day
    const isSelectedDay = (day: string): boolean => {
        const found = selectedDays.find((elem) => elem === day);
        return Boolean(found); // Return true if found is truthy, false otherwise
    };

    // change rrule Type.
    const onRepeatTypeSelected = (event: React.ChangeEvent<HTMLSelectElement>, data: any) => {
        setRepeatType(event.target.value);
    };

    const onUntilDaySelected = (day: Date, dateRangeArray: Date[]) => {
        setSelectedUntil(day);
        hideCalendar();
    };
    const InputEveryElem = useRef<HTMLInputElement>(null);

    const RepeatTypeSettingsView = () => {
        if (repeatType === RepeatFreq.Weekly) {
            return (<div className='recurrence-container'>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[0])}
                    onClick={() => onDaySelected(0)}
                    shape='circular'
                    size='small'
                >
                    Mo
                </ToggleButton>

                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[1])}
                    onClick={() => onDaySelected(1)}
                    shape='circular'
                    size='small'
                >
                    Tu
                </ToggleButton>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[2])}
                    onClick={() => onDaySelected(2)}
                    shape='circular'
                    size='small'
                >
                    We
                </ToggleButton>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[3])}
                    onClick={() => onDaySelected(3)}
                    shape='circular'
                    size='small'
                >
                    Th
                </ToggleButton>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[4])}
                    onClick={() => onDaySelected(4)}
                    shape='circular'
                    size='small'
                >
                    Fr
                </ToggleButton>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[5])}
                    onClick={() => onDaySelected(5)}
                    shape='circular'
                    size='small'
                >
                    Sa
                </ToggleButton>
                <ToggleButton
                    className='weekday-day-button'
                    checked={isSelectedDay(RepeatByWeekDay[6])}
                    onClick={() => onDaySelected(6)}
                    shape='circular'
                    size='small'
                >
                    Su
                </ToggleButton>

            </div>);
        }
        return <div>Default</div>;
    };

    useEffect(() => {
        if (!didMount.current) {
            didMount.current = true;
            if (props.selected.length <= 0) {
                setSelectedDays([]);
                setRepeatEvery(1);
                setRepeatType(RepeatFreq.Weekly);
                return;
            }

            const ruleString = props.selected.split(':')[1];
            const params = ruleString.split(';');

            const updateState = () => {
                params.forEach((elem) => {
                    const kv = elem.split('=');
                    switch (kv[0]) {
                    case 'FREQ':
                        setRepeatType(kv[1]);
                        break;
                    case 'INTERVAL':
                        setRepeatEvery(kv[1]);
                        break;
                    case 'BYDAY':
                        setSelectedDays(kv[1].split(','));
                        break;
                    case 'UNTIL':
                        setSelectedUntil(fromRruleUntilUtc(kv[1]));
                        break;
                    default:
                        break;
                    }
                });
            };

            updateState();
        } else {
            onChangeTrigger();
        }
    }, [props.selected, selectedDays, repeatEvery, selectedUntil]);

    return (<div className='repeat-type-container'>
        <div className='repeat-type-select-container'>
            <Select
                className='repeat-type-selector'
                onChange={onRepeatTypeSelected}
            >
                <option value={RepeatFreq.Weekly}>Weekly</option>
                {/* <option value="custom_monthly">Monthly</option> */}
            </Select>
            <p><Text className='repeat-every-label'>every:</Text><SpinButton
                ref={InputEveryElem}
                defaultValue={1}
                min={1}
                max={10}
                value={repeatEvery}
                onChange={(event: SpinButtonChangeEvent, data: SpinButtonOnChangeData) => {
                    if (data.value === null || data.value === undefined || !Number.isInteger(data.value)) {
                        return;
                    }
                    setRepeatEvery(data.value!);
                }}
            /></p>
        </div>
        <RepeatTypeSettingsView/>

        <div className='event-until-container'>
            <Popover
                trapFocus={true}
                open={isShowCalendar}
                closeOnScroll={true}
                unstable_disableAutoFocus={true}
                onOpenChange={(e: OpenPopoverEvents, data: OnOpenChangeData) => (data.open ? showCalendar() : hideCalendar())}
            >
                <PopoverTrigger disableButtonEnhancement={true}>
                    <Text
                        className='event-until-open-calendar'
                        onClick={toggleShowCalendar}
                    >Choose an end date
                    </Text>
                </PopoverTrigger>

                <PopoverSurface className='repeat-date-until-popover'>
                    <Calendar
                        showMonthPickerAsOverlay={true}
                        highlightSelectedMonth={true}
                        showGoToToday={true}
                        onSelectDate={onUntilDaySelected}
                        firstDayOfWeek={firstDayOfWeek}
                    />
                </PopoverSurface>
            </Popover>

            <span className='selected-date-until'>{selectedUntil ? format(selectedUntil, 'PP') : ''}</span>
            <span className='remove-selected-date-until'>{selectedUntil ?
                <Text onClick={() => setSelectedUntil(undefined)}>Remove end date</Text> : ''}</span>
        </div>
    </div>);
};

export default RepeatEventCustom;
// Pure helpers for the Calendar page's event form (EventForm.svelte):
// parsing typed times, and converting between the form's flat state and
// the backend's calendar.EventInput / calendar.Event shapes.

import { toDateKey } from "./habitDisplay.js";
import { isValidDateKey, parseDateKey } from "./workoutProgram.js";

export const TIME_MODES = [
    { value: "at", label: "At a time" },
    { value: "range", label: "From – to" },
    { value: "allday", label: "All day" },
];

export const REPEAT_OPTIONS = [
    { value: "none", label: "Doesn't repeat" },
    { value: "daily", label: "Daily" },
    { value: "weekly", label: "Weekly" },
    { value: "monthly", label: "Monthly" },
    { value: "yearly", label: "Yearly" },
];

export const REPEAT_UNITS = { daily: "day", weekly: "week", monthly: "month", yearly: "year" };

export const END_MODES = [
    { value: "never", label: "Never" },
    { value: "count", label: "After N times" },
    { value: "until", label: "On a date" },
];

// Values are lead times in seconds; "" means no reminder.
export const REMINDER_OPTIONS = [
    { value: "", label: "No reminder" },
    { value: "0", label: "At start time" },
    { value: "300", label: "5 min before" },
    { value: "600", label: "10 min before" },
    { value: "900", label: "15 min before" },
    { value: "1800", label: "30 min before" },
    { value: "3600", label: "1 hour before" },
    { value: "86400", label: "1 day before" },
    { value: "604800", label: "1 week before" },
];

// Monday-first, matching calendarGrid.js's WEEKDAY_LABELS. Go's
// time.Weekday is Sunday-indexed, hence the conversions below.
const goWeekday = (mondayIndex) => (mondayIndex + 1) % 7;
const mondayIndex = (goDay) => (goDay + 6) % 7;

// Parses a typed time of day: "9", "9am", "9:30 pm", "21:30", "930",
// "2130". Without am/pm the hour is read as 24-hour. Returns { h, m } or
// null if the text isn't a valid time.
export function parseTimeOfDay(text) {
    const match = text
        .trim()
        .toLowerCase()
        .match(/^(\d{1,2})(?::?(\d{2}))?\s*(a|p|am|pm|a\.m\.|p\.m\.)?$/);
    if (!match) {
        return null;
    }

    let h = Number(match[1]);
    const m = match[2] ? Number(match[2]) : 0;
    const meridiem = match[3]?.[0];

    if (m > 59) {
        return null;
    }
    if (meridiem) {
        if (h < 1 || h > 12) {
            return null;
        }
        h = (h % 12) + (meridiem === "p" ? 12 : 0);
    } else if (h > 23) {
        return null;
    }

    return { h, m };
}

export function formatTimeOfDay(date) {
    return date.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
}

// Form state for a brand-new event on `date`.
export function emptyForm(date) {
    const weekdays = Array(7).fill(false);
    weekdays[mondayIndex(date.getDay())] = true;

    return {
        title: "",
        date: toDateKey(date),
        timeMode: "at",
        startTime: "9:00 AM",
        endTime: "10:00 AM",
        repeat: "none",
        interval: 1,
        weekdays,
        endMode: "never",
        count: 10,
        untilDate: "",
        reminder: "",
        // Only shown when editing.
        description: "",
        location: "",
        important: false,
        // Reminders past the first one (the form edits only one); kept so
        // an edit doesn't silently drop them.
        extraReminderSeconds: [],
    };
}

// Form state for editing an existing calendar.Event.
export function eventToForm(event) {
    const start = new Date(event.start);
    const end = new Date(event.end);
    const form = emptyForm(start);

    form.title = event.title;
    form.description = event.description ?? "";
    form.location = event.location ?? "";
    form.important = event.important;

    if (event.allDay) {
        form.timeMode = "allday";
    } else {
        form.startTime = formatTimeOfDay(start);
        form.endTime = formatTimeOfDay(end);
        form.timeMode = end.getTime() === start.getTime() ? "at" : "range";
    }

    const rule = event.recurrence;
    if (rule) {
        form.repeat = rule.frequency;
        form.interval = rule.interval || 1;
        if (rule.weekdays?.length) {
            form.weekdays = Array(7).fill(false);
            for (const d of rule.weekdays) {
                form.weekdays[mondayIndex(d)] = true;
            }
        }
        if (rule.count) {
            form.endMode = "count";
            form.count = rule.count;
        } else if (rule.until) {
            form.endMode = "until";
            form.untilDate = toDateKey(new Date(rule.until));
        }
    }

    const leadSeconds = (event.reminders ?? []).map((r) => Math.round(r.leadTime / 1e9));
    if (leadSeconds.length > 0) {
        form.reminder = String(leadSeconds[0]);
        form.extraReminderSeconds = leadSeconds.slice(1);
    }

    return form;
}

function atTime(dateKey, text, fieldName) {
    const time = parseTimeOfDay(text);
    if (!time) {
        throw new Error(`${fieldName} isn't a time (try 9am, 2:30pm or 14:30)`);
    }
    const day = parseDateKey(dateKey);
    return new Date(day.getFullYear(), day.getMonth(), day.getDate(), time.h, time.m);
}

function positiveInt(value, fieldName) {
    const n = Number(value);
    if (!Number.isInteger(n) || n < 1) {
        throw new Error(`${fieldName} must be a whole number of at least 1`);
    }
    return n;
}

// Converts form state to a calendar.EventInput. Throws an Error with a
// user-facing message if a field is invalid.
export function formToInput(form) {
    const title = form.title.trim();
    if (!title) {
        throw new Error("Title is required");
    }
    if (!isValidDateKey(form.date)) {
        throw new Error("Date must be YYYY-MM-DD");
    }

    let start;
    let end;
    if (form.timeMode === "allday") {
        start = parseDateKey(form.date);
        end = start;
    } else {
        start = atTime(form.date, form.startTime, "Start time");
        end = start;
        if (form.timeMode === "range") {
            end = atTime(form.date, form.endTime, "End time");
            if (end <= start) {
                throw new Error("End time must be after the start time");
            }
        }
    }

    let recurrence;
    if (form.repeat !== "none") {
        recurrence = {
            frequency: form.repeat,
            interval: positiveInt(form.interval, "Repeat interval"),
        };

        if (form.repeat === "weekly") {
            const weekdays = form.weekdays.flatMap((on, i) => (on ? [goWeekday(i)] : []));
            if (weekdays.length === 0) {
                throw new Error("Pick at least one day of the week");
            }
            recurrence.weekdays = weekdays;
        }

        if (form.endMode === "count") {
            recurrence.count = positiveInt(form.count, "Number of times");
        } else if (form.endMode === "until") {
            if (!isValidDateKey(form.untilDate)) {
                throw new Error("End date must be YYYY-MM-DD");
            }
            const until = parseDateKey(form.untilDate);
            // Inclusive: an occurrence on the end date itself still happens.
            until.setHours(23, 59, 59, 999);
            if (until < start) {
                throw new Error("Repeat end date must not be before the start");
            }
            recurrence.until = until;
        }
    }

    const reminderLeadSeconds = [
        ...(form.reminder === "" ? [] : [Number(form.reminder)]),
        ...form.extraReminderSeconds,
    ];

    return {
        title,
        description: form.description.trim(),
        location: form.location.trim(),
        start,
        end,
        allDay: form.timeMode === "allday",
        important: form.important,
        recurrence,
        reminderLeadSeconds,
    };
}

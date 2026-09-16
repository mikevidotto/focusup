const DAY_MS = 24 * 60 * 60 * 1000;

function startOfDay(date) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

// Formats an already-computed occurrence time for display (e.g. "Today",
// "Tomorrow · 9:00 AM"). This only labels a timestamp the backend already
// derived — it never decides *which* dates an event falls on.
export function formatOccurrenceWhen(occ, now = new Date()) {
    const start = new Date(occ.start);
    const diffDays = Math.round((startOfDay(start) - startOfDay(now)) / DAY_MS);

    let dayLabel;
    if (diffDays === 0) {
        dayLabel = "Today";
    } else if (diffDays === 1) {
        dayLabel = "Tomorrow";
    } else {
        dayLabel = start.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" });
    }

    if (occ.allDay) {
        return dayLabel;
    }

    const timeLabel = start.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
    return `${dayLabel} · ${timeLabel}`;
}

// Formats just the time portion of an occurrence (day is implied by context,
// e.g. a month-grid day cell already showing the date).
export function formatOccurrenceTime(occ) {
    if (occ.allDay) {
        return "All day";
    }

    return new Date(occ.start).toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
}

// A stable React/Svelte {#each} key for an occurrence: the same event
// recurring many times shares an eventId, so the key needs originalStart too.
export function occurrenceKey(occ) {
    return `${occ.eventId}-${occ.originalStart}`;
}

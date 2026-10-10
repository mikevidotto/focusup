// Formats just the time portion of an occurrence — its start, or a
// "start – end" range — since the day is implied by context (e.g. a
// month-grid day cell already showing the date).
export function formatOccurrenceTime(occ) {
    if (occ.allDay) {
        return "All day";
    }

    const start = new Date(occ.start);
    const end = new Date(occ.end);
    const format = (date) => date.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });

    // A point-in-time event (end == start) shows just its start.
    if (end.getTime() <= start.getTime()) {
        return format(start);
    }

    return `${format(start)} – ${format(end)}`;
}

// A stable React/Svelte {#each} key for an occurrence: the same event
// recurring many times shares an eventId, so the key needs originalStart too.
export function occurrenceKey(occ) {
    return `${occ.eventId}-${occ.originalStart}`;
}

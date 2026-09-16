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

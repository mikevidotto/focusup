// Pure helpers for the Jobs page and widget. Dates in the ai-job-search
// files are plain "YYYY-MM-DD" days: the tracker's date column is when a row
// was added (drafted), and "applied YYYY-MM-DD" is appended to its notes (and
// written as the queue line's outcome) when the application is submitted.

const DATE_IN_TEXT = /(\d{4}-\d{2}-\d{2})/;

// When the application was submitted, or null if it hasn't been. Falls back
// to the row date for rows marked applied without a dated note.
export function appliedDate(app) {
    if (app.status !== "applied") {
        return null;
    }

    const match = app.notes.match(/applied (\d{4}-\d{2}-\d{2})/);
    return match ? match[1] : app.date || null;
}

// The date in a finished queue line's outcome ("applied 2026-10-08"), if any.
export function queueOutcomeDate(item) {
    const match = (item.outcome ?? "").match(DATE_IN_TEXT);
    return match ? match[1] : null;
}

function fromKey(dateKey) {
    const [y, m, d] = dateKey.split("-").map(Number);
    return new Date(y, m - 1, d);
}

export function isThisWeek(dateKey) {
    if (!dateKey) {
        return false;
    }

    return fromKey(dateKey) >= mondayOf(new Date());
}

// "2026-10-08" -> "Oct 8"; anything that isn't a date is returned as-is.
export function formatShortDate(dateKey) {
    if (!dateKey || !DATE_IN_TEXT.test(dateKey)) {
        return dateKey ?? "";
    }

    return fromKey(dateKey.match(DATE_IN_TEXT)[1]).toLocaleDateString(
        "en-US",
        { month: "short", day: "numeric" },
    );
}

function mondayOf(date) {
    const monday = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    monday.setDate(monday.getDate() - ((monday.getDay() + 6) % 7));
    return monday;
}

function keyOf(date) {
    const m = String(date.getMonth() + 1).padStart(2, "0");
    const d = String(date.getDate()).padStart(2, "0");
    return `${date.getFullYear()}-${m}-${d}`;
}

// Applications submitted per Monday–Sunday week, for the last `n` weeks
// (oldest first, ending with this week), shaped for MiniColumnChart.
export function weeklyApplied(apps, today, n = 8) {
    const dates = apps.map(appliedDate).filter(Boolean);
    const thisMonday = mondayOf(today);
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const start = new Date(thisMonday);
        start.setDate(start.getDate() - 7 * i);
        const end = new Date(start);
        end.setDate(end.getDate() + 6);

        const startKey = keyOf(start);
        const endKey = keyOf(end);
        const value = dates.filter((d) => d >= startKey && d <= endKey).length;
        const range = `${formatShortDate(startKey)} – ${formatShortDate(endKey)}`;

        buckets.push({
            key: startKey,
            label: formatShortDate(startKey),
            value,
            title: `${value} applied`,
            sub: i === 0 ? `${range} (so far)` : range,
            current: i === 0,
        });
    }

    return buckets;
}

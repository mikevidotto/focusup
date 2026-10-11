// Pure helpers for the Jobs page and widget. Dates in the ai-job-search
// files are plain "YYYY-MM-DD" days: the tracker's date column is when a row
// was added (drafted), and "applied YYYY-MM-DD" is appended to its notes (and
// written as the queue line's outcome) when the application is submitted.

// The Jobs page stats chart cycles through these with "v".
export const JOB_STATS_VIEWS = ["day", "week", "month"];

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

function addDays(date, n) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() + n);
}

function appliedDates(apps) {
    return apps.map(appliedDate).filter(Boolean);
}

function countBetween(dates, startKey, endKey) {
    return dates.filter((d) => d >= startKey && d <= endKey).length;
}

// Applications submitted per day for the last `n` days (oldest first,
// ending with today), shaped for MiniColumnChart. Columns are labelled by
// weekday for a week or less, otherwise by day of month so every column
// keeps a label; the tooltip carries the full date.
export function dailyApplied(apps, today, n = 14) {
    const dates = appliedDates(apps);
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const day = addDays(today, -i);
        const key = keyOf(day);
        const value = countBetween(dates, key, key);
        const weekday = day.toLocaleDateString("en-US", { weekday: "short" });

        buckets.push({
            key,
            label: n <= 7 ? weekday : String(day.getDate()),
            value,
            title: `${value} applied`,
            sub: `${weekday} ${formatShortDate(key)}${i === 0 ? " (today)" : ""}`,
            current: i === 0,
        });
    }

    return buckets;
}

// Applications submitted per Monday–Sunday week, for the last `n` weeks
// (oldest first, ending with this week), shaped for MiniColumnChart.
export function weeklyApplied(apps, today, n = 8) {
    const dates = appliedDates(apps);
    const thisMonday = mondayOf(today);
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const start = addDays(thisMonday, -7 * i);
        const startKey = keyOf(start);
        const endKey = keyOf(addDays(start, 6));
        const value = countBetween(dates, startKey, endKey);
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

// Applications submitted per calendar month, for the last `n` months
// (oldest first, ending with this month), shaped for MiniColumnChart.
export function monthlyApplied(apps, today, n = 12) {
    const dates = appliedDates(apps);
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const start = new Date(today.getFullYear(), today.getMonth() - i, 1);
        const end = new Date(start.getFullYear(), start.getMonth() + 1, 0);
        const value = countBetween(dates, keyOf(start), keyOf(end));
        const month = start.toLocaleDateString("en-US", {
            month: "long",
            year: "numeric",
        });

        buckets.push({
            key: keyOf(start),
            label: start.toLocaleDateString("en-US", { month: "short" }),
            value,
            title: `${value} applied`,
            sub: i === 0 ? `${month} (so far)` : month,
            current: i === 0,
        });
    }

    return buckets;
}

// Applications submitted in the `days` days ending today.
export function appliedInLastDays(apps, today, days) {
    return countBetween(
        appliedDates(apps),
        keyOf(addDays(today, -(days - 1))),
        keyOf(today),
    );
}

// Runs of consecutive days with at least one application. The current run
// still stands as of yesterday when nothing has gone out yet today, so it
// doesn't reset each morning.
export function applyStreak(apps, today) {
    const days = new Set(appliedDates(apps));

    let cursor = days.has(keyOf(today)) ? today : addDays(today, -1);
    let current = 0;
    while (days.has(keyOf(cursor))) {
        current++;
        cursor = addDays(cursor, -1);
    }

    let best = 0;
    let run = 0;
    let prev = null;
    for (const key of [...days].sort()) {
        const date = fromKey(key);
        run = prev && keyOf(addDays(prev, 1)) === key ? run + 1 : 1;
        best = Math.max(best, run);
        prev = date;
    }

    return { current, best };
}

// Where the search stands: tracked rows by status, finished queue lines that
// were skipped, and the share of tracked rows that went out (null when none).
export function jobFunnel(apps, queue) {
    const applied = apps.filter((app) => app.status === "applied").length;
    const drafted = apps.filter((app) => app.status === "drafted").length;

    return {
        applied,
        drafted,
        queued: queue.filter((item) => !item.done).length,
        skipped: queue.filter(
            (item) => item.done && (item.outcome ?? "").startsWith("skipped"),
        ).length,
        appliedRate:
            applied + drafted > 0 ? applied / (applied + drafted) : null,
    };
}

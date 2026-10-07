// Pure helpers that bucket a habit's completion dates ("YYYY-MM-DD") into
// the periods the Habits page history chart plots. All dates are local
// calendar days; keys compare correctly as plain strings.

import { toDateKey } from "./habitDisplay.js";
import { startOfDay } from "./calendarGrid.js";

export const HISTORY_VIEWS = ["week", "month", "year"];

export const HISTORY_VIEW_LABELS = {
    week: "Weekly",
    month: "Monthly",
    year: "Year",
};

const DAY_MS = 24 * 60 * 60 * 1000;

function addDays(date, days) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() + days);
}

function mondayOf(date) {
    return addDays(date, -((date.getDay() + 6) % 7));
}

function daysBetweenInclusive(start, end) {
    return Math.round((startOfDay(end) - startOfDay(start)) / DAY_MS) + 1;
}

function shortDate(date) {
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

// The first day a habit counts as "tracked": its creation day, or an
// earlier day if one was ticked retroactively. Periods entirely before this
// are drawn as untracked rather than as zeros.
export function trackedSince(habit) {
    const created = habit.createdAt ? toDateKey(new Date(habit.createdAt)) : null;
    const first = habit.completions?.[0] ?? null;
    const candidates = [created, first].filter(Boolean).sort();

    return candidates[0] ?? toDateKey(new Date());
}

function countInRange(completions, startKey, endKey) {
    let count = 0;
    for (const key of completions) {
        if (key >= startKey && key <= endKey) {
            count++;
        }
    }
    return count;
}

// Builds one bucket per period. `possible` is how many days of the period
// could have been completed: today caps the current period, and the
// tracked-since day caps the first one.
function bucket(habit, start, end, today, label, rangeLabel) {
    const startKey = toDateKey(start);
    const endKey = toDateKey(end);
    const sinceKey = trackedSince(habit);
    const todayKey = toDateKey(today);

    const tracked = endKey >= sinceKey;
    const effectiveStart = startKey < sinceKey ? new Date(`${sinceKey}T00:00`) : start;
    const effectiveEnd = endKey > todayKey ? today : end;
    const possible = tracked ? Math.max(daysBetweenInclusive(effectiveStart, effectiveEnd), 0) : 0;

    return {
        key: startKey,
        label,
        rangeLabel,
        capacity: daysBetweenInclusive(start, end),
        count: countInRange(habit.completions ?? [], startKey, endKey),
        possible,
        tracked,
        current: startKey <= todayKey && todayKey <= endKey,
    };
}

// The last `n` Monday–Sunday weeks, oldest first, ending with this week.
export function weeklyBuckets(habit, today, n = 12) {
    const thisMonday = mondayOf(today);
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const start = addDays(thisMonday, -7 * i);
        const end = addDays(start, 6);
        buckets.push(
            bucket(habit, start, end, today, shortDate(start), `${shortDate(start)} – ${shortDate(end)}`),
        );
    }

    return buckets;
}

// The last `n` calendar months, oldest first, ending with this month.
export function monthlyBuckets(habit, today, n = 12) {
    const buckets = [];

    for (let i = n - 1; i >= 0; i--) {
        const start = new Date(today.getFullYear(), today.getMonth() - i, 1);
        const end = new Date(start.getFullYear(), start.getMonth() + 1, 0);
        const label = start.toLocaleDateString("en-US", { month: "short" });
        const rangeLabel = start.toLocaleDateString("en-US", { month: "long", year: "numeric" });
        buckets.push(bucket(habit, start, end, today, label, rangeLabel));
    }

    return buckets;
}

// A contribution-style grid: `weeks` Monday-first columns ending with this
// week, each holding seven day cells.
export function yearGrid(habit, today, weeks = 53) {
    const done = new Set(habit.completions ?? []);
    const sinceKey = trackedSince(habit);
    const todayKey = toDateKey(today);
    const firstMonday = addDays(mondayOf(today), -7 * (weeks - 1));

    const columns = [];
    for (let w = 0; w < weeks; w++) {
        const days = [];
        for (let d = 0; d < 7; d++) {
            const date = addDays(firstMonday, w * 7 + d);
            const key = toDateKey(date);
            days.push({
                key,
                date,
                done: done.has(key),
                future: key > todayKey,
                tracked: key >= sinceKey,
            });
        }
        columns.push({ key: toDateKey(days[0].date), days });
    }

    return columns;
}

// Current streak counts back from today; if today isn't ticked yet the
// streak still stands as of yesterday, so it doesn't reset each morning.
export function streaks(habit, today) {
    const done = new Set(habit.completions ?? []);

    let cursor = done.has(toDateKey(today)) ? today : addDays(today, -1);
    let current = 0;
    while (done.has(toDateKey(cursor))) {
        current++;
        cursor = addDays(cursor, -1);
    }

    let best = 0;
    let run = 0;
    let prev = null;
    for (const key of habit.completions ?? []) {
        const date = new Date(`${key}T00:00`);
        run = prev && daysBetweenInclusive(prev, date) === 2 ? run + 1 : 1;
        best = Math.max(best, run);
        prev = date;
    }

    return { current, best };
}

// Share of possible days completed across the given buckets, 0..1, or null
// when nothing in range was trackable yet.
export function completionRate(buckets) {
    const possible = buckets.reduce((sum, b) => sum + b.possible, 0);
    if (possible === 0) {
        return null;
    }
    const count = buckets.reduce((sum, b) => sum + b.count, 0);
    return count / possible;
}

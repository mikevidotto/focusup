// Pure 5/3/1 helpers for the Workouts page. The backend stores cycles
// (start date, training maxes, logged workouts); everything shown per set,
// plus the projected schedule, is derived here.
//
// A cycle is 4 weeks × 4 lifts = 16 workouts, addressed by index
// (week - 1) * 4 + liftIndex, matching internal/workouts.

import { toDateKey } from "./habitDisplay.js";
import { startOfDay } from "./calendarGrid.js";

export const LIFTS = ["squat", "bench", "deadlift", "press"];

export const LIFT_META = {
    squat: { label: "Squat", day: "Mon" },
    bench: { label: "Bench", day: "Tue" },
    deadlift: { label: "Deadlift", day: "Thu" },
    press: { label: "Press", day: "Fri" },
};

// Rest days after each lift before the next one is due (Mon→Tue→Thu→Fri→Mon).
export const GAP_AFTER = { squat: 1, bench: 2, deadlift: 1, press: 3 };

export const WORKOUTS_PER_CYCLE = 16;
export const DELOAD_WEEK = 4;

const WARMUPS = [
    { pct: 40, reps: 5 },
    { pct: 50, reps: 5 },
    { pct: 60, reps: 3 },
];

// Main sets per week; the last set of weeks 1–3 is AMRAP ("5+").
export const WEEKS = {
    1: { label: "5s", main: [[65, 5], [75, 5], [85, 5]] },
    2: { label: "3s", main: [[70, 3], [80, 3], [90, 3]] },
    3: { label: "5/3/1", main: [[75, 5], [85, 3], [95, 1]] },
    4: { label: "Deload", main: [[40, 5], [50, 5], [60, 5]] },
};

const VOLUME = { pct: 60, sets: 5, reps: 10 };

export function roundTo5(weight) {
    return Math.round(weight / 5) * 5;
}

export function trainingMaxFrom1RM(oneRepMax) {
    return roundTo5(0.9 * oneRepMax);
}

export function workoutAt(index) {
    return { week: Math.floor(index / 4) + 1, lift: LIFTS[index % 4] };
}

export function isDeload(index) {
    return workoutAt(index).week === DELOAD_WEEK;
}

// Every set for one lift in one week. Deload week is only its three light
// sets: they already sit at warm-up weights, and there's no AMRAP or 5×10.
export function setsFor(tm, week) {
    const at = (pct) => roundTo5((pct / 100) * tm);
    const deload = week === DELOAD_WEEK;
    const sets = [];

    if (!deload) {
        for (const w of WARMUPS) {
            sets.push({ kind: "warmup", pct: w.pct, reps: w.reps, amrap: false, weight: at(w.pct) });
        }
    }

    WEEKS[week].main.forEach(([pct, reps], i) => {
        const amrap = !deload && i === WEEKS[week].main.length - 1;
        sets.push({ kind: "main", pct, reps, amrap, weight: at(pct) });
    });

    if (!deload) {
        for (let i = 0; i < VOLUME.sets; i++) {
            sets.push({ kind: "volume", pct: VOLUME.pct, reps: VOLUME.reps, amrap: false, weight: at(VOLUME.pct) });
        }
    }

    return sets;
}

export function topSet(tm, week) {
    return setsFor(tm, week).filter((s) => s.kind === "main").at(-1);
}

// Epley estimate; null when there's nothing to estimate from.
export function estimated1RM(weight, reps) {
    if (!weight || !reps) {
        return null;
    }
    return Math.round(weight * (1 + reps / 30));
}

export function workoutEstimate(cycle, log) {
    const { week, lift } = workoutAt(log.index);
    if (week === DELOAD_WEEK) {
        return null;
    }
    return estimated1RM(topSet(cycle.trainingMax[lift], week).weight, log.amrapReps);
}

export function bestEstimates(cycle) {
    const best = {};
    for (const log of cycle.logs ?? []) {
        const e = workoutEstimate(cycle, log);
        const { lift } = workoutAt(log.index);
        if (e && (!best[lift] || e > best[lift])) {
            best[lift] = e;
        }
    }
    return best;
}

export function parseDateKey(key) {
    const [y, m, d] = key.split("-").map(Number);
    return new Date(y, m - 1, d);
}

function addDays(date, days) {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate() + days);
}

function daysBetween(a, b) {
    return Math.round((startOfDay(b) - startOfDay(a)) / 86400000);
}

// Status for each of the cycle's 16 workouts. Logged ones are "done" on
// their date. The first unlogged one is "next": due a rest gap after the
// last workout actually done (or on the cycle's start date), but never
// before today — if that day has passed it's due today and `lateDays`
// says by how much the rest of the cycle slid. Later ones follow on
// from it by their rest gaps.
export function schedule(cycle, today) {
    const logs = cycle.logs ?? [];
    const out = logs.map((log) => ({ status: "done", date: parseDateKey(log.date) }));

    if (logs.length >= WORKOUTS_PER_CYCLE) {
        return out;
    }

    const n = logs.length;
    const planned =
        n > 0
            ? addDays(parseDateKey(logs[n - 1].date), GAP_AFTER[workoutAt(n - 1).lift])
            : parseDateKey(cycle.startDate);
    const lateDays = Math.max(0, daysBetween(planned, today));
    let due = lateDays > 0 ? startOfDay(today) : planned;

    out.push({ status: "next", due, lateDays });

    for (let i = n + 1; i < WORKOUTS_PER_CYCLE; i++) {
        due = addDays(due, GAP_AFTER[workoutAt(i - 1).lift]);
        out.push({ status: "upcoming", due });
    }

    return out;
}

export function dueLabel(due, today) {
    const diff = daysBetween(today, due);
    if (diff === 0) return "today";
    if (diff === 1) return "tomorrow";
    return `in ${diff} days`;
}

export function formatDay(date) {
    return date.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" });
}

export function isValidDateKey(key) {
    return /^\d{4}-\d{2}-\d{2}$/.test(key) && toDateKey(parseDateKey(key)) === key;
}

export { toDateKey };

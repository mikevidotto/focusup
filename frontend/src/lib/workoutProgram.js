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

// ---- projections ----
// If the program is followed on schedule, every cycle adds a fixed amount
// to each training max (mirrors internal/workouts TMIncrement) and starts
// 28 days after the previous one (Mon of week 1 → Mon of the next week 1).

export const TM_INCREMENT = { squat: 10, bench: 5, deadlift: 10, press: 5 };
export const CYCLE_DAYS = 28;

// Lifts that count toward the powerlifting total.
export const TOTAL_LIFTS = ["squat", "bench", "deadlift"];

export const PLATE_MILESTONES = [135, 225, 315, 405, 495, 585];
export const TOTAL_MILESTONES = [1000, 1200, 1500];

// The 1RM a training max stands for, since TM = 90% of 1RM.
export function implied1RM(tm) {
    return Math.round(tm / 0.9);
}

// When the cycle after this one should start: a rest gap after its last
// workout, done or still scheduled. Missed days push this later, just
// like they shift the rest of the cycle in schedule().
export function nextCycleStart(cycle, today) {
    const last = schedule(cycle, today).at(-1);
    return addDays(last.status === "done" ? last.date : last.due, GAP_AFTER.press);
}

// The cycle `k` cycles after `latest` (k = 0 is `latest` itself, on its
// real start date). `nextStart` is nextCycleStart(latest, today).
export function projectedCycle(latest, nextStart, k) {
    const trainingMax = Object.fromEntries(
        LIFTS.map((lift) => [lift, latest.trainingMax[lift] + k * TM_INCREMENT[lift]]),
    );
    const implied = Object.fromEntries(LIFTS.map((lift) => [lift, implied1RM(trainingMax[lift])]));
    const topSingle = Object.fromEntries(LIFTS.map((lift) => [lift, topSet(trainingMax[lift], 3).weight]));

    return {
        k,
        number: latest.number + k,
        start: k === 0 ? parseDateKey(latest.startDate) : addDays(nextStart, (k - 1) * CYCLE_DAYS),
        trainingMax,
        implied,
        topSingle,
        total: TOTAL_LIFTS.reduce((sum, lift) => sum + implied[lift], 0),
    };
}

// The latest cycle plus the next `count` cycles.
export function projectCycles(latest, today, count) {
    const nextStart = nextCycleStart(latest, today);
    return Array.from({ length: count + 1 }, (_, k) => projectedCycle(latest, nextStart, k));
}

// The projected cycle in effect on `date` (the latest one for any earlier date).
export function projectionAt(latest, today, date) {
    const nextStart = nextCycleStart(latest, today);
    const k = date < nextStart ? 0 : Math.floor(daysBetween(nextStart, date) / CYCLE_DAYS) + 1;
    return projectedCycle(latest, nextStart, k);
}

// The first projected cycle whose implied 1RM for `lift` (or the total,
// when lift is "total") reaches `target`. Gives up after ~40 years.
export function cycleToReach(latest, today, lift, target) {
    const nextStart = nextCycleStart(latest, today);
    for (let k = 0; k <= 520; k++) {
        const c = projectedCycle(latest, nextStart, k);
        if ((lift === "total" ? c.total : c.implied[lift]) >= target) {
            return c;
        }
    }
    return null;
}

// The next milestone above where `lift` (or "total") stands in the latest
// cycle, and when it's reached; null once past the last one.
export function nextMilestone(latest, today, lift) {
    const current = lift === "total" ? projectedCycle(latest, null, 0).total : implied1RM(latest.trainingMax[lift]);
    const target = (lift === "total" ? TOTAL_MILESTONES : PLATE_MILESTONES).find((m) => m > current);
    if (!target) {
        return null;
    }
    return { target, cycle: cycleToReach(latest, today, lift, target) };
}

// Chart series for one lift, each a list of { date, value, number }.
// - program: the implied 1RM of every past cycle's TM, then the projection
// - pace: starts from the most recent cycle with an AMRAP e1RM and adds the
//   same increment per cycle, so it reflects how you're actually lifting
// - actual: the best AMRAP e1RM logged in each cycle
export function projectionSeries(cycles, rows, lift) {
    const latest = cycles.at(-1);
    const program = [
        ...cycles.slice(0, -1).map((c) => ({
            date: parseDateKey(c.startDate),
            value: implied1RM(c.trainingMax[lift]),
            number: c.number,
        })),
        ...rows.map((r) => ({ date: r.start, value: r.implied[lift], number: r.number })),
    ];

    const actual = cycles
        .map((c) => ({ date: parseDateKey(c.startDate), value: bestEstimates(c)[lift], number: c.number }))
        .filter((p) => p.value);

    let pace = null;
    const from = actual.at(-1);
    if (from) {
        pace = [
            from,
            ...rows
                .filter((r) => r.number > from.number)
                .map((r) => ({
                    date: r.start,
                    value: from.value + (r.number - from.number) * TM_INCREMENT[lift],
                    number: r.number,
                })),
        ];
    }

    return { program, pace, actual };
}

export function formatLongDay(date) {
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

// "in 3 months", "in 2 weeks", "today", "now" — rough distance for labels.
export function fromNow(date, today) {
    const days = daysBetween(today, date);
    if (days <= 0) return "now";
    if (days < 14) return `in ${days} day${days === 1 ? "" : "s"}`;
    if (days < 60) return `in ${Math.round(days / 7)} weeks`;
    const months = Math.round(days / 30.44);
    if (months < 24) return `in ${months} months`;
    return `in ${(days / 365.25).toFixed(1)} years`;
}

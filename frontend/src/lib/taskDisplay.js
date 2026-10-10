import { buildWeekGrid, isSameDay, startOfDay } from "./calendarGrid.js";

// GTD lists, in the order the Tasks page cycles through them with [ / ].
export const LISTS = [
    { id: "inbox", label: "Inbox" },
    { id: "next", label: "Next Actions" },
    { id: "someday", label: "Someday/Maybe" },
    { id: "done", label: "Done" }
];

const DAY_MS = 24 * 60 * 60 * 1000;
const REVIEW_INTERVAL_DAYS = 7;

export function splitTasks(tasks) {
    const active = tasks.filter(t => !t.done);
    const completed = tasks
        .filter(t => t.done)
        .slice()
        .sort((a, b) => new Date(b.completedAt) - new Date(a.completedAt));

    return { active, completed };
}

export function formatCompletedDate(iso) {
    if (!iso) {
        return "";
    }

    return new Date(iso).toLocaleDateString("en-US", {
        month: "short",
        day: "numeric"
    });
}

// Splits captured text like "call mom @phone @errands" into its title and
// contexts, so tasks can be tagged inline while typing.
export function parseCapture(text) {
    const contexts = [];
    const words = [];

    for (const word of text.trim().split(/\s+/)) {
        if (word.length > 1 && word.startsWith("@")) {
            contexts.push(word.slice(1).toLowerCase());
        } else if (word) {
            words.push(word);
        }
    }

    return { title: words.join(" "), contexts };
}

// The inverse of parseCapture, used to prefill the edit input.
export function formatForEdit(task) {
    return [task.title, ...(task.contexts ?? []).map(c => "@" + c)].join(" ");
}

export function allContexts(tasks) {
    return [...new Set(tasks.flatMap(t => t.contexts ?? []))].sort();
}

// Open tasks per list, plus every done task under "done", for the list
// tab counts.
export function countByList(tasks) {
    const counts = Object.fromEntries(LISTS.map(l => [l.id, 0]));

    for (const task of tasks) {
        const list = task.done ? "done" : task.list;
        counts[list] = (counts[list] ?? 0) + 1;
    }

    return counts;
}

export function isDoneToday(task, now = new Date()) {
    return task.done && !!task.completedAt && isSameDay(new Date(task.completedAt), now);
}

// Tasks completed since Monday (weeks start on Monday, as in calendarGrid.js).
export function doneThisWeek(tasks, now = new Date()) {
    const monday = buildWeekGrid(now)[0].date;
    return tasks.filter(t => t.done && t.completedAt && new Date(t.completedAt) >= monday).length;
}

// Done tasks from before today, newest first, grouped under day labels
// ("Yesterday", "Mon, Oct 6") for the Done list. Today's completions stay
// inline in their own lists instead.
export function groupDoneByDay(tasks, now = new Date()) {
    const today = startOfDay(now);
    const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1);
    const groups = [];

    const older = splitTasks(tasks).completed.filter(t => !t.completedAt || new Date(t.completedAt) < today);

    for (const task of older) {
        const day = task.completedAt ? startOfDay(new Date(task.completedAt)) : null;
        const label = !day
            ? "Unknown"
            : isSameDay(day, yesterday)
              ? "Yesterday"
              : day.toLocaleDateString("en-US", {
                    weekday: "short",
                    month: "short",
                    day: "numeric",
                    year: day.getFullYear() === today.getFullYear() ? undefined : "numeric"
                });

        const last = groups[groups.length - 1];
        if (last && last.label === label) {
            last.tasks.push(task);
        } else {
            groups.push({ label, tasks: [task] });
        }
    }

    return groups;
}

// A short age like "today", "3d", "2w" or "4mo".
export function formatAge(iso, now = new Date()) {
    if (!iso) {
        return "";
    }

    const days = Math.floor((startOfDay(now) - startOfDay(new Date(iso))) / DAY_MS);
    if (days <= 0) return "today";
    if (days < 14) return `${days}d`;
    if (days < 60) return `${Math.floor(days / 7)}w`;
    return `${Math.floor(days / 30)}mo`;
}

export function nextActionCount(project, tasks) {
    return tasks.filter(t => t.projectId === project.id && !t.done && t.list === "next").length;
}

// Active projects with no open next action — GTD's "stalled" projects.
export function stalledProjects(projects, tasks) {
    return projects.filter(p => !p.done && nextActionCount(p, tasks) === 0);
}

// True when the weekly review was never done or is a week or more old.
export function isReviewDue(lastReviewAt, now = new Date()) {
    if (!lastReviewAt) {
        return true;
    }

    return now - new Date(lastReviewAt) >= REVIEW_INTERVAL_DAYS * DAY_MS;
}

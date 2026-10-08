// GTD lists, in the order the Tasks page cycles through them with [ / ].
export const LISTS = [
    { id: "inbox", label: "Inbox" },
    { id: "next", label: "Next Actions" },
    { id: "someday", label: "Someday/Maybe" }
];

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

// Open tasks per list, for the list tab counts.
export function countByList(tasks) {
    const counts = Object.fromEntries(LISTS.map(l => [l.id, 0]));

    for (const task of tasks) {
        if (!task.done) {
            counts[task.list] = (counts[task.list] ?? 0) + 1;
        }
    }

    return counts;
}

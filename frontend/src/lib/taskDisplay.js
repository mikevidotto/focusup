export const PRIORITY_ORDER = ["low", "medium", "high"];

export const PRIORITY_META = {
    high: { icon: "↑", color: "var(--accent)", label: "High" },
    medium: { icon: "–", color: "var(--warning)", label: "Medium" },
    low: { icon: "↓", color: "var(--success)", label: "Low" }
};

export function nextPriority(priority) {
    const index = PRIORITY_ORDER.indexOf(priority);
    return PRIORITY_ORDER[(index + 1) % PRIORITY_ORDER.length];
}

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

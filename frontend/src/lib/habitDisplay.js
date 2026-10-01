export function splitHabits(habits) {
    const active = habits.filter(t => !t.done);
    const completed = habits
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

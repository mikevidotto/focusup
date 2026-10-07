// Pure helpers for the Habits page. A habit stores the days it was
// completed on as "YYYY-MM-DD" strings (local calendar days), so rendering
// any week is just checking which of its seven dates are in that list.

// Formats a Date as a local "YYYY-MM-DD" key, matching the backend's
// habits.DateLayout. Deliberately not toISOString(), which is UTC and would
// shift the day for anyone east/west of Greenwich late at night.
export function toDateKey(date) {
    const y = date.getFullYear();
    const m = String(date.getMonth() + 1).padStart(2, "0");
    const d = String(date.getDate()).padStart(2, "0");

    return `${y}-${m}-${d}`;
}

export function isDoneOn(habit, date) {
    return (habit.completions ?? []).includes(toDateKey(date));
}

export function countDoneInWeek(habit, weekCells) {
    return weekCells.filter((cell) => isDoneOn(habit, cell.date)).length;
}

// "Oct 5 – Oct 11, 2026"
export function formatWeekRange(weekCells) {
    const first = weekCells[0].date;
    const last = weekCells[weekCells.length - 1].date;
    const fmt = (date) =>
        date.toLocaleDateString("en-US", { month: "short", day: "numeric" });

    return `${fmt(first)} – ${fmt(last)}, ${last.getFullYear()}`;
}

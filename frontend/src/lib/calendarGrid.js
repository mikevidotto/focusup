// Pure spatial-navigation helpers for calendar day grids, mirroring
// keyboardGrid.js's shape (explicit row/col metadata per cell, a pure move
// function) even though the "widgets" here are calendar days.
//
// Weeks start on Monday and end on Sunday throughout (month grid, week
// strip, and the weekday labels below) — a deliberate choice, not the
// Sunday-start JS Date.getDay() default.

export const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

// Converts JS's Sunday-indexed getDay() (0=Sun..6=Sat) to a Monday-indexed
// column (0=Mon..6=Sun).
function mondayIndex(date) {
    return (date.getDay() + 6) % 7;
}

// Returns every cell for the month grid containing `year`/`month`
// (JS Date's 0-indexed month), padded with leading/trailing days from
// adjacent months. Always 6 rows (42 cells) so the grid's height doesn't
// jump between 5 and 6 rows as the month changes.
export function buildMonthGrid(year, month) {
    const firstOfMonth = new Date(year, month, 1);
    const gridStart = new Date(year, month, 1 - mondayIndex(firstOfMonth));

    const totalCells = 42;
    const cells = [];
    for (let i = 0; i < totalCells; i++) {
        const date = new Date(gridStart.getFullYear(), gridStart.getMonth(), gridStart.getDate() + i);
        cells.push({
            date,
            row: Math.floor(i / 7),
            col: i % 7,
            inCurrentMonth: date.getMonth() === month
        });
    }

    return cells;
}

// Returns the 7 cells (Mon..Sun) for the week containing `date`.
export function buildWeekGrid(date) {
    const monday = new Date(date.getFullYear(), date.getMonth(), date.getDate() - mondayIndex(date));

    const cells = [];
    for (let i = 0; i < 7; i++) {
        const cellDate = new Date(monday.getFullYear(), monday.getMonth(), monday.getDate() + i);
        cells.push({ date: cellDate, row: 0, col: i, inCurrentMonth: true });
    }

    return cells;
}

export function isSameDay(a, b) {
    return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

// Moves the day cursor by one cell in the given direction. left/right/down
// clamp at the grid's edges (no wrap), same as moveSelection in
// keyboardGrid.js. up returns null instead of clamping when already in the
// top row — same signal moveSelection uses to tell the caller to hand focus
// back to the tab bar.
export function moveDayCursor(cells, currentIndex, direction) {
    switch (direction) {
        case "left":
            return Math.max(currentIndex - 1, 0);
        case "right":
            return Math.min(currentIndex + 1, cells.length - 1);
        case "up": {
            const next = currentIndex - 7;
            return next >= 0 ? next : null;
        }
        case "down": {
            const next = currentIndex + 7;
            return next < cells.length ? next : currentIndex;
        }
        default:
            return currentIndex;
    }
}

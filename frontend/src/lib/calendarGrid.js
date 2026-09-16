// Pure spatial-navigation helpers for the month page's day grid, mirroring
// keyboardGrid.js's shape (explicit row/col metadata per cell, a pure move
// function) even though the "widgets" here are calendar days.
//
// Note: unlike the dashboard grid, this can't use h/l for movement — h/l are
// reserved globally in App.svelte for tab switching in "tabs" mode, which is
// the mode every full page (including this one) runs in. So day-to-day
// movement is a single j/k axis (which still reads naturally left-to-right,
// top-to-bottom through the grid, matching how a calendar is read); month
// switching uses "[" / "]" instead.

// Returns every cell for the month grid containing `year`/`month`
// (JS Date's 0-indexed month), padded with leading/trailing days from
// adjacent months so every displayed week is a full row of 7.
export function buildMonthGrid(year, month) {
    const firstOfMonth = new Date(year, month, 1);
    const startWeekday = firstOfMonth.getDay(); // 0 = Sunday
    const gridStart = new Date(year, month, 1 - startWeekday);

    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const totalCellsNeeded = startWeekday + daysInMonth;
    const weeks = Math.ceil(totalCellsNeeded / 7);
    const totalCells = weeks * 7;

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

export function isSameDay(a, b) {
    return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate();
}

// Moves the day cursor by one cell, clamped at the grid's edges (no wrap),
// matching moveSelection's clamp behavior in keyboardGrid.js.
export function moveDayCursor(cells, currentIndex, direction) {
    switch (direction) {
        case "next":
            return Math.min(currentIndex + 1, cells.length - 1);
        case "prev":
            return Math.max(currentIndex - 1, 0);
        default:
            return currentIndex;
    }
}

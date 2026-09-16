// Pure spatial-navigation helpers over a list of { id, row, col } widgets.
// row/col are explicit logical grid positions (not derived from the CSS grid),
// so navigation stays correct as more widgets/rows are added later.

export function firstWidget(widgets) {
    if (widgets.length === 0) {
        return null;
    }

    return widgets.slice().sort((a, b) => a.row - b.row || a.col - b.col)[0];
}

function closestInRow(widgets, row, col) {
    const inRow = widgets.filter(w => w.row === row);

    return inRow.reduce((best, w) =>
        Math.abs(w.col - col) < Math.abs(best.col - col) ? w : best
    );
}

// Returns the next widget id for the given direction, or null when
// direction is "up" and there is no row above (caller should exit to tab-bar mode).
// Clamps at all other edges (no wrap).
export function moveSelection(widgets, currentId, direction) {
    const current = widgets.find(w => w.id === currentId);

    if (!current) {
        return currentId;
    }

    const sameRow = widgets.filter(w => w.row === current.row);
    const rows = [...new Set(widgets.map(w => w.row))].sort((a, b) => a - b);

    switch (direction) {
        case "left": {
            const candidates = sameRow
                .filter(w => w.col < current.col)
                .sort((a, b) => b.col - a.col);

            return candidates[0]?.id ?? current.id;
        }

        case "right": {
            const candidates = sameRow
                .filter(w => w.col > current.col)
                .sort((a, b) => a.col - b.col);

            return candidates[0]?.id ?? current.id;
        }

        case "up": {
            const rowIndex = rows.indexOf(current.row);

            if (rowIndex <= 0) {
                return null;
            }

            return closestInRow(widgets, rows[rowIndex - 1], current.col).id;
        }

        case "down": {
            const rowIndex = rows.indexOf(current.row);

            if (rowIndex === rows.length - 1) {
                return current.id;
            }

            return closestInRow(widgets, rows[rowIndex + 1], current.col).id;
        }

        default:
            return current.id;
    }
}

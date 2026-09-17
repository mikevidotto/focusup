<script>
    import { onMount, onDestroy } from "svelte";

    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import { buildMonthGrid, isSameDay, moveDayCursor, WEEKDAY_LABELS } from "../calendarGrid.js";
    import { formatOccurrenceTime, occurrenceKey } from "../calendarDisplay.js";
    import { ListCalendarOccurrences, ToggleEventCompletion } from "../../../wailsjs/go/main/App.js";

    const today = new Date();

    let viewedYear = today.getFullYear();
    let viewedMonth = today.getMonth();
    let cells = buildMonthGrid(viewedYear, viewedMonth);
    let cursor = Math.max(0, cells.findIndex(c => isSameDay(c.date, today)));
    let occurrences = [];
    let loading = true;
    let error = null;

    async function fetchOccurrences() {
        loading = true;

        try {
            const rangeStart = cells[0].date;
            const rangeEnd = cells[cells.length - 1].date;
            occurrences = await ListCalendarOccurrences(rangeStart, rangeEnd);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    }

    async function changeMonth(delta) {
        let month = viewedMonth + delta;
        let year = viewedYear;

        if (month < 0) {
            month = 11;
            year -= 1;
        } else if (month > 11) {
            month = 0;
            year += 1;
        }

        viewedYear = year;
        viewedMonth = month;
        cells = buildMonthGrid(viewedYear, viewedMonth);
        cursor = Math.max(0, cells.findIndex(c => c.inCurrentMonth));

        await fetchOccurrences();
    }

    async function toggleCompletion(occ) {
        try {
            await ToggleEventCompletion(occ.eventId, occ.originalStart);
            await fetchOccurrences();
        } catch (e) {
            error = String(e);
        }
    }

    function occurrencesForDay(date) {
        return occurrences
            .filter(o => isSameDay(new Date(o.start), date))
            .sort((a, b) => new Date(a.start) - new Date(b.start));
    }

    function handleKey(event) {
        switch (event.key) {
            case "h":
                event.preventDefault();
                cursor = moveDayCursor(cells, cursor, "left");
                break;

            case "l":
                event.preventDefault();
                cursor = moveDayCursor(cells, cursor, "right");
                break;

            case "j":
                event.preventDefault();
                cursor = moveDayCursor(cells, cursor, "down");
                break;

            case "k": {
                event.preventDefault();
                const next = moveDayCursor(cells, cursor, "up");

                if (next === null) {
                    mode.set("tabs");
                } else {
                    cursor = next;
                }

                break;
            }

            case "[":
                event.preventDefault();
                changeMonth(-1);
                break;

            case "]":
                event.preventDefault();
                changeMonth(1);
                break;
        }
    }

    onMount(async () => {
        await fetchOccurrences();
        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
    });

    $: monthLabel = new Date(viewedYear, viewedMonth, 1).toLocaleDateString("en-US", {
        month: "long",
        year: "numeric"
    });
    $: selectedCell = cells[cursor];
    $: selectedDayOccurrences = selectedCell ? occurrencesForDay(selectedCell.date) : [];
</script>

<div class="page-placeholder calendar-page">
    <span class="eyebrow">FOCUSUP / CALENDAR</span>

    <div class="tasks-heading-row">
        <h1>{monthLabel}</h1>
        {#if loading}
            <span class="todo-header-count">Loading…</span>
        {/if}
    </div>

    <div class="tasks-hint">
        <kbd>h</kbd><kbd>j</kbd><kbd>k</kbd><kbd>l</kbd> move day
        <kbd>[</kbd><kbd>]</kbd> change month
    </div>

    <div class="calendar-grid">
        {#each WEEKDAY_LABELS as label}
            <div class="calendar-weekday">{label}</div>
        {/each}

        {#each cells as cell, index (cell.date.toISOString())}
            {@const dayOccurrences = occurrencesForDay(cell.date)}
            <div
                class="calendar-day-cell"
                class:dimmed={!cell.inCurrentMonth}
                class:today={isSameDay(cell.date, today)}
                class:cursor={index === cursor}
            >
                <span class="calendar-day-number">{cell.date.getDate()}</span>

                {#if dayOccurrences.length > 0}
                    <ul class="calendar-day-events">
                        {#each dayOccurrences.slice(0, 2) as occ (occurrenceKey(occ))}
                            <li>
                                <button
                                    type="button"
                                    class="calendar-day-event"
                                    class:done={occ.done}
                                    title={occ.title}
                                    on:click={() => toggleCompletion(occ)}
                                >
                                    {occ.done ? "✓ " : ""}{occ.title}
                                </button>
                            </li>
                        {/each}

                        {#if dayOccurrences.length > 2}
                            <li class="calendar-day-more">+{dayOccurrences.length - 2} more</li>
                        {/if}
                    </ul>
                {/if}
            </div>
        {/each}
    </div>

    <div class="calendar-detail-panel">
        <div class="calendar-detail-title">
            {selectedCell
                ? selectedCell.date.toLocaleDateString("en-US", { weekday: "long", month: "long", day: "numeric" })
                : ""}
        </div>

        {#if selectedDayOccurrences.length === 0}
            <p class="calendar-detail-empty">Nothing scheduled</p>
        {:else}
            <ul class="calendar-detail-list">
                {#each selectedDayOccurrences as occ (occurrenceKey(occ))}
                    <li class="calendar-detail-item" class:done={occ.done}>
                        <button
                            type="button"
                            class="calendar-detail-check"
                            aria-label={occ.done ? "Mark not done" : "Mark done"}
                            on:click={() => toggleCompletion(occ)}
                        >
                            {occ.done ? "☑" : "☐"}
                        </button>
                        <span class="calendar-detail-time">{formatOccurrenceTime(occ)}</span>
                        <span class="calendar-detail-event-title">{occ.title}</span>
                    </li>
                {/each}
            </ul>
        {/if}
    </div>

    {#if error}
        <p class="todo-error">{error}</p>
    {/if}
</div>

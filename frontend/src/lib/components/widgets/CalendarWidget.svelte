<script>
    import { onMount } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import {
        buildWeekGrid,
        buildMonthGrid,
        isSameDay,
        startOfDay,
        endOfDay,
        WEEKDAY_LABELS
    } from "../../calendarGrid.js";
    import { formatOccurrenceTime, occurrenceKey } from "../../calendarDisplay.js";
    import { ListCalendarOccurrences, ToggleEventCompletion } from "../../../../wailsjs/go/main/App.js";

    export let focused = false;

    const MODES = ["today", "week", "month"];

    const today = new Date();
    const weekCells = buildWeekGrid(today);
    const monthCells = buildMonthGrid(today.getFullYear(), today.getMonth());

    let mode = "week";
    let occurrences = [];
    let loading = true;
    let error = null;

    async function fetchOccurrences() {
        loading = true;
        error = null;

        try {
            let rangeStart;
            let rangeEnd;

            if (mode === "today") {
                rangeStart = startOfDay(today);
                rangeEnd = endOfDay(today);
            } else if (mode === "week") {
                rangeStart = weekCells[0].date;
                rangeEnd = endOfDay(weekCells[weekCells.length - 1].date);
            } else {
                rangeStart = monthCells[0].date;
                rangeEnd = endOfDay(monthCells[monthCells.length - 1].date);
            }

            occurrences = await ListCalendarOccurrences(rangeStart, rangeEnd);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    }

    async function toggleCompletion(occ) {
        try {
            await ToggleEventCompletion(occ.eventId, occ.originalStart);
            await fetchOccurrences();
        } catch (e) {
            error = String(e);
        }
    }

    function cycleMode(delta) {
        const index = MODES.indexOf(mode);
        mode = MODES[(index + delta + MODES.length) % MODES.length];
        fetchOccurrences();
    }

    function handleKey(event) {
        switch (event.key) {
            case "h":
                event.preventDefault();
                cycleMode(-1);
                break;

            case "l":
                event.preventDefault();
                cycleMode(1);
                break;
        }
    }

    onMount(async () => {
        await fetchOccurrences();
    });

    $: if (focused) {
        activeWidgetKeyHandler.set(handleKey);
    } else {
        activeWidgetKeyHandler.set(null);
    }

    // Reactive *functions*, not plain ones: Svelte's dependency tracking for
    // `$:`/`{@const}` only sees identifiers referenced directly in that
    // expression, not inside a called function's body — a plain function
    // closing over `occurrences` would silently stop updating callers
    // whenever `occurrences` changes without some *other* tracked variable
    // also happening to change. Declaring these with `$:` makes each
    // function itself a tracked dependency wherever it's called.
    $: occurrencesForDay = date => occurrences.filter(o => isSameDay(new Date(o.start), date));

    // null = no events that day, "done" = every event that day is done,
    // "important" = not every event is done and at least one is important,
    // "normal" = events but none done/important.
    $: dayIndicator = date => {
        const dayOccs = occurrencesForDay(date);
        if (dayOccs.length === 0) {
            return null;
        }
        if (dayOccs.every(o => o.done)) {
            return "done";
        }
        return dayOccs.some(o => o.important) ? "important" : "normal";
    };

    $: todayOccurrences = occurrencesForDay(today).sort((a, b) => new Date(a.start) - new Date(b.start));
    $: modeLabel =
        mode === "today"
            ? "Today"
            : mode === "week"
            ? "This Week"
            : today.toLocaleDateString("en-US", { month: "long", year: "numeric" });
</script>

<div class="calendar-widget">
    <header class="calendar-widget-header">
        <span class="calendar-widget-title">Calendar</span>
        <span class="calendar-widget-count">{modeLabel}</span>
    </header>

    <div class="calendar-widget-body">
        {#if loading}
            <div class="empty-widget">
                <span>…</span>
                <p>Loading events</p>
            </div>
        {:else if mode === "today"}
            {#if todayOccurrences.length === 0}
                <div class="empty-widget">
                    <span>空</span>
                    <p>Nothing scheduled today</p>
                </div>
            {:else}
                <ul class="calendar-widget-list">
                    {#each todayOccurrences as occ (occurrenceKey(occ))}
                        <li class="calendar-widget-item" class:done={occ.done}>
                            <button
                                type="button"
                                class="calendar-detail-check"
                                aria-label={occ.done ? "Mark not done" : "Mark done"}
                                on:click={() => toggleCompletion(occ)}
                            >
                                {occ.done ? "☑" : "☐"}
                            </button>
                            <span class="calendar-widget-time">{formatOccurrenceTime(occ)}</span>
                            <span class="calendar-widget-event-title">{occ.title}</span>
                        </li>
                    {/each}
                </ul>
            {/if}
        {:else if mode === "week"}
            <div class="calendar-widget-week">
                {#each weekCells as cell (cell.date.toISOString())}
                    {@const indicator = dayIndicator(cell.date)}
                    <div class="calendar-widget-week-cell" class:today={isSameDay(cell.date, today)}>
                        <span class="calendar-widget-weekday">{WEEKDAY_LABELS[cell.col]}</span>
                        <span class="calendar-widget-day-number">{cell.date.getDate()}</span>
                        <span
                            class="calendar-widget-dot"
                            class:visible={indicator !== null}
                            class:important={indicator === "important"}
                            class:done={indicator === "done"}
                        ></span>
                    </div>
                {/each}
            </div>
        {:else}
            <div class="calendar-widget-month-grid">
                {#each WEEKDAY_LABELS as label}
                    <span class="calendar-widget-month-weekday">{label[0]}</span>
                {/each}

                {#each monthCells as cell (cell.date.toISOString())}
                    {@const indicator = dayIndicator(cell.date)}
                    <div
                        class="calendar-widget-month-cell"
                        class:dimmed={!cell.inCurrentMonth}
                        class:today={isSameDay(cell.date, today)}
                    >
                        <span class="calendar-widget-day-number">{cell.date.getDate()}</span>
                        <span
                            class="calendar-widget-dot"
                            class:visible={indicator !== null}
                            class:important={indicator === "important"}
                            class:done={indicator === "done"}
                        ></span>
                    </div>
                {/each}
            </div>
        {/if}

        {#if error}
            <p class="calendar-widget-error">{error}</p>
        {/if}
    </div>
</div>

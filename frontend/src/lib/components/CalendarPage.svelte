<script>
    //import {svelte} from '@sveltejs/vite-plugin-svelte'

    import { onMount, onDestroy } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import { calendar } from "../../../wailsjs/go/models";
    import {
        buildMonthGrid,
        isSameDay,
        moveDayCursor,
        WEEKDAY_LABELS,
    } from "../calendarGrid.js";
    import { formatOccurrenceTime, occurrenceKey } from "../calendarDisplay.js";
    import {
        ListCalendarOccurrences,
        ToggleEventCompletion,
        AddEvent,
        DeleteEvent,
    } from "../../../wailsjs/go/main/App.js";

    const today = new Date();

    let now = new Date();
    let insertMode = false;
    let viewedYear = today.getFullYear();
    let viewedMonth = today.getMonth();
    let cells = buildMonthGrid(viewedYear, viewedMonth);
    let cursor = Math.max(
        0,
        cells.findIndex((c) => isSameDay(c.date, today)),
    );
    let eventCursor = 0;
    let occurrences = [];
    let newTitle = "";
    let loading = true;
    let error = null;
    let listMode = false;
    let inputEl;
    var selectedOcc;
    let occEventId = "yo";

    $: timeStr = now.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit",
    });

    function exitInsertMode() {
        insertMode = false;
        newTitle = "";
        inputEl?.blur();
    }

    async function enterInsertMode() {
        insertMode = true;
        await tick();
        inputEl?.focus();
    }

    async function onInputKeydown(event) {
        if (event.key === "Enter") {
            event.preventDefault();

            const title = newTitle.trim();

            if (title) {
                try {
                    const ruleWithUntil = {
                        frequency: "yearly",
                        interval: 2,
                        count: 10,
                        until: new Date("2026-12-31T23:59:59Z").toISOString(),
                    };

                    const rule =
                        calendar.RecurrenceRule.createFrom(ruleWithUntil);
                    const created = await AddEvent(
                        title,
                        "",
                        "",
                        cells[cursor].date,
                        cells[cursor].date,
                        false,
                        rule,
                        false,
                    );
                    occurrences = [...occurrences, created];
                } catch (e) {
                    error = String(e);
                }
            }
            exitInsertMode();
        } else if (event.key === "Escape") {
            event.preventDefault();
            exitInsertMode();
        }
        await fetchOccurrences();
    }

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
        cursor = Math.max(
            0,
            cells.findIndex((c) => c.inCurrentMonth),
        );

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

    function handleKey(event) {
        if (!listMode) {
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
                case "Enter":
                    event.preventDefault();
                    listMode = true;
                    break;
                case "[":
                    event.preventDefault();
                    changeMonth(-1);
                    break;

                case "]":
                    event.preventDefault();
                    changeMonth(1);
                    break;
            }
        } else if (listMode) {
            if (event.key === "a") {
                event.preventDefault();
                inputEl?.focus();
                enterInsertMode();
                return;
            }

            if (!insertMode) {
                switch (event.key) {
                    case "q":
                        listMode = false;
                        break;

                    case "Enter":
                        toggleCompletion(selectedDayOccurrences[eventCursor]);
                        break;

                    case "k":
                        if (eventCursor === 0) {
                        } else {
                            eventCursor = Math.min(
                                eventCursor - 1,
                                selectedDayOccurrences.length - 1,
                            );
                        }
                        break;
                    case "j":
                        eventCursor = Math.min(
                            eventCursor + 1,
                            selectedDayOccurrences.length - 1,
                        );
                        break;
                    case "x":
                        DeleteOccurrence(selectedDayOccurrences[eventCursor]);
                        break;
                }
            }
        }
    }

    async function DeleteOccurrence(occ) {
        DeleteEvent(occ.eventId);
        await fetchOccurrences();
    }

    onMount(async () => {
        await fetchOccurrences();
        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
    });

    $: monthLabel = new Date(viewedYear, viewedMonth, 1).toLocaleDateString(
        "en-US",
        {
            month: "long",
            year: "numeric",
        },
    );
    $: selectedCell = cells[cursor];
    // A reactive *function*, not a plain one: Svelte's dependency tracking
    // for `$:`/`{@const}` only sees identifiers referenced directly in the
    // expression, not inside a called function's body — a plain function
    // closing over `occurrences` would silently stop updating callers
    // whenever `occurrences` changes without some *other* dependency (like
    // `cursor`) also happening to change. Declaring it with `$:` makes the
    // function itself a tracked dependency wherever it's called.
    $: occurrencesForDay = (date) =>
        occurrences
            .filter((o) => isSameDay(new Date(o.start), date))
            .sort((a, b) => new Date(a.start) - new Date(b.start));
    $: selectedDayOccurrences = selectedCell
        ? occurrencesForDay(selectedCell.date)
        : [];
</script>

<div class="page-placeholder calendar-page" style="margin-top:0">
    <div class="calendar-page-left">
        <!--
    <span class="eyebrow">FOCUSUP / CALENDAR</span>
    -->

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
                    <span class="calendar-day-number"
                        >{cell.date.getDate()}</span
                    >

                    {#if dayOccurrences.length > 0}
                        <ul class="calendar-day-events">
                            {#each dayOccurrences.slice(0, 2) as occ (occurrenceKey(occ))}
                                <li>
                                    <button
                                        type="button"
                                        class="calendar-day-event"
                                        class:done={occ.done}
                                        title={occ.title}
                                    >
                                        {occ.done ? "[✓ ]" : "[ ] "}
                                    </button>
                                </li>
                            {/each}

                            {#if dayOccurrences.length > 2}
                                <li class="calendar-day-more">
                                    +{dayOccurrences.length - 2} more
                                </li>
                            {/if}
                        </ul>
                    {/if}
                </div>
            {/each}
        </div>
    </div>

    <div class="calendar-page-right">
        <div class="calendar-detail-title">
            {selectedCell
                ? selectedCell.date.toLocaleDateString("en-US", {
                      weekday: "long",
                      month: "long",
                      day: "numeric",
                  })
                : ""}
        </div>
        <div class="calendar-detail-panel" class:cursor={listMode === true}>
            {#if selectedDayOccurrences.length === 0}
                <p class="calendar-detail-empty">Nothing scheduled</p>
            {:else}
                <ul class="calendar-detail-list" class:cursor={listMode}>
                    {#each selectedDayOccurrences as occ, index (occurrenceKey(occ))}
                        <li
                            class="calendar-detail-item"
                            class:done={occ.done}
                            class:cursor={index === eventCursor &&
                                listMode === true}
                        >
                            <button
                                type="button"
                                class="calendar-detail-check"
                                aria-label={occ.done
                                    ? "Mark not done"
                                    : "Mark done"}
                            >
                                {occ.done ? "☑" : "☐"}
                            </button>
                            <span class="calendar-detail-time"
                                >{formatOccurrenceTime(occ)}</span
                            >
                            <span class="calendar-detail-event-title"
                                >{occ.title}</span
                            >
                        </li>
                    {/each}
                </ul>
            {/if}
            <div class="tasks-add-row">
                <input
                    class="todo-input"
                    type="text"
                    bind:this={inputEl}
                    bind:value={newTitle}
                    placeholder="press a to add a task…"
                    on:keydown={onInputKeydown}
                    on:focus={() => (insertMode = true)}
                    on:blur={exitInsertMode}
                />
            </div>
        </div>

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    </div>
</div>

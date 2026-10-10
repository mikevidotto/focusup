<script>
    // Dashboard calendar: a week strip with the rest of the week as an
    // agenda (plus anything overdue from earlier in the week), or a mini
    // month grid with the next few pending events.
    //
    // Keyboard while focused: h/l switch week <-> month, j/k move through
    // the listed events, enter toggles the selected one done.
    import { tick } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import {
        buildWeekGrid,
        buildMonthGrid,
        isSameDay,
        startOfDay,
        endOfDay,
        WEEKDAY_LABELS,
    } from "../../calendarGrid.js";
    import {
        formatOccurrenceTime,
        occurrenceKey,
    } from "../../calendarDisplay.js";
    import {
        ListCalendarOccurrences,
        ToggleEventCompletion,
    } from "../../../../wailsjs/go/main/App.js";

    export let focused = false;

    const MODES = ["week", "month"];
    const MAX_PIPS = 3;
    const NEXT_UP_COUNT = 3;

    const today = startOfDay(new Date());
    const tomorrow = new Date(
        today.getFullYear(),
        today.getMonth(),
        today.getDate() + 1,
    );
    const weekCells = buildWeekGrid(today);
    const monthCells = buildMonthGrid(today.getFullYear(), today.getMonth());
    // The agenda always reaches tomorrow, even on a Sunday when tomorrow
    // is already next week.
    const agendaEnd = endOfDay(
        new Date(
            Math.max(weekCells[weekCells.length - 1].date.getTime(), tomorrow.getTime()),
        ),
    );

    let mode = "week";
    let occurrences = [];
    let loading = true;
    let error = null;
    let cursor = 0;
    let listEl;

    async function fetchOccurrences() {
        error = null;

        try {
            const [rangeStart, rangeEnd] =
                mode === "week"
                    ? [weekCells[0].date, agendaEnd]
                    : [
                          monthCells[0].date,
                          endOfDay(monthCells[monthCells.length - 1].date),
                      ];

            occurrences = (
                await ListCalendarOccurrences(rangeStart, rangeEnd)
            ).sort((a, b) => new Date(a.start) - new Date(b.start));
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
        cursor = 0;
        loading = true;
        fetchOccurrences();
    }

    async function moveCursor(delta) {
        cursor = Math.min(Math.max(cursor + delta, 0), selectable.length - 1);
        await tick();
        listEl
            ?.querySelector(".calendar-widget-item.cursor")
            ?.scrollIntoView({ block: "nearest" });
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

            case "j":
                event.preventDefault();
                moveCursor(1);
                break;

            case "k":
                event.preventDefault();
                moveCursor(-1);
                break;

            case "Enter":
                event.preventDefault();
                if (selectable[cursor]) toggleCompletion(selectable[cursor]);
                break;
        }
    }

    function dayLabel(date) {
        if (isSameDay(date, today)) return "Today";
        if (isSameDay(date, tomorrow)) return "Tomorrow";
        return date.toLocaleDateString("en-US", {
            weekday: "short",
            day: "numeric",
        });
    }

    function pipState(occ) {
        if (occ.done) return "done";
        return occ.important ? "important" : "normal";
    }

    fetchOccurrences();

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
    $: occurrencesForDay = (date) =>
        occurrences.filter((o) => isSameDay(new Date(o.start), date));

    // Up to MAX_PIPS state names ("normal" | "important" | "done"), one per
    // event that day, plus whether there were more than fit.
    $: dayPips = (date) => {
        const dayOccs = occurrencesForDay(date);
        return {
            pips: dayOccs.slice(0, MAX_PIPS).map(pipState),
            more: dayOccs.length > MAX_PIPS,
        };
    };

    // Week agenda: pending events from earlier this week, then each day
    // from today on that has anything scheduled.
    $: agendaGroups = (() => {
        const groups = [];

        const overdue = occurrences.filter(
            (o) => !o.done && new Date(o.start) < today,
        );
        if (overdue.length > 0) {
            groups.push({ label: "Overdue", overdue: true, items: overdue });
        }

        for (
            let day = today;
            day <= agendaEnd;
            day = new Date(day.getFullYear(), day.getMonth(), day.getDate() + 1)
        ) {
            const items = occurrencesForDay(day);
            if (items.length > 0) {
                groups.push({ label: dayLabel(day), items });
            }
        }

        return groups;
    })();

    $: nextUp = occurrences
        .filter((o) => !o.done && new Date(o.end) >= new Date())
        .slice(0, NEXT_UP_COUNT);

    $: selectable =
        mode === "week" ? agendaGroups.flatMap((g) => g.items) : nextUp;

    // Keep the cursor on a real row as the list shrinks.
    $: if (cursor > selectable.length - 1) {
        cursor = Math.max(0, selectable.length - 1);
    }

    $: modeLabel =
        mode === "week"
            ? "This Week"
            : today.toLocaleDateString("en-US", {
                  month: "long",
                  year: "numeric",
              });
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
        {:else if mode === "week"}
            <div class="calendar-widget-week">
                {#each weekCells as cell (cell.date.toISOString())}
                    {@const day = dayPips(cell.date)}
                    <div
                        class="calendar-widget-week-cell"
                        class:today={isSameDay(cell.date, today)}
                        class:past={cell.date < today}
                    >
                        <span class="calendar-widget-weekday"
                            >{WEEKDAY_LABELS[cell.col]}</span
                        >
                        <span class="calendar-widget-day-number"
                            >{cell.date.getDate()}</span
                        >
                        <span class="calendar-widget-pips">
                            {#each day.pips as state}
                                <span class="calendar-widget-pip {state}"></span>
                            {/each}
                            {#if day.more}
                                <span class="calendar-widget-pip-more">+</span>
                            {/if}
                        </span>
                    </div>
                {/each}
            </div>

            <div class="calendar-widget-agenda" bind:this={listEl}>
                {#if agendaGroups.length === 0}
                    <p class="calendar-widget-empty">Nothing else this week</p>
                {:else}
                    {#each agendaGroups as group (group.label)}
                        <section class="calendar-widget-agenda-group">
                            <h3
                                class="calendar-widget-agenda-day"
                                class:overdue={group.overdue}
                            >
                                {group.label}
                            </h3>
                            <ul class="calendar-widget-list">
                                {#each group.items as occ (occurrenceKey(occ))}
                                    {@const showDay = group.overdue}
                                    <li
                                        class="calendar-widget-item"
                                        class:done={occ.done}
                                        class:important={occ.important}
                                        class:cursor={focused &&
                                            selectable[cursor] === occ}
                                    >
                                        <button
                                            type="button"
                                            class="calendar-detail-check"
                                            aria-label={occ.done
                                                ? "Mark not done"
                                                : "Mark done"}
                                            on:click={() => toggleCompletion(occ)}
                                        >
                                            {occ.done ? "☑" : "☐"}
                                        </button>
                                        <span class="calendar-widget-time"
                                            >{showDay
                                                ? dayLabel(new Date(occ.start))
                                                : formatOccurrenceTime(occ)}</span
                                        >
                                        <span class="calendar-widget-event-title"
                                            >{occ.title}</span
                                        >
                                        {#if occ.recurring}
                                            <span
                                                class="calendar-widget-repeat"
                                                title="Repeats">↻</span
                                            >
                                        {/if}
                                    </li>
                                {/each}
                            </ul>
                        </section>
                    {/each}
                {/if}
            </div>
        {:else}
            <div class="calendar-widget-month-grid">
                {#each WEEKDAY_LABELS as label}
                    <span class="calendar-widget-month-weekday">{label[0]}</span
                    >
                {/each}

                {#each monthCells as cell (cell.date.toISOString())}
                    {@const day = dayPips(cell.date)}
                    <div
                        class="calendar-widget-month-cell"
                        class:dimmed={!cell.inCurrentMonth}
                        class:today={isSameDay(cell.date, today)}
                    >
                        <span class="calendar-widget-day-number"
                            >{cell.date.getDate()}</span
                        >
                        <span class="calendar-widget-pips">
                            {#each day.pips as state}
                                <span class="calendar-widget-pip {state}"></span>
                            {/each}
                        </span>
                    </div>
                {/each}
            </div>

            <div class="calendar-widget-agenda" bind:this={listEl}>
                <h3 class="calendar-widget-agenda-day">Next up</h3>
                {#if nextUp.length === 0}
                    <p class="calendar-widget-empty">Nothing pending</p>
                {:else}
                    <ul class="calendar-widget-list">
                        {#each nextUp as occ (occurrenceKey(occ))}
                            <li
                                class="calendar-widget-item"
                                class:important={occ.important}
                                class:cursor={focused &&
                                    selectable[cursor] === occ}
                            >
                                <button
                                    type="button"
                                    class="calendar-detail-check"
                                    aria-label="Mark done"
                                    on:click={() => toggleCompletion(occ)}
                                >
                                    ☐
                                </button>
                                <span class="calendar-widget-time"
                                    >{dayLabel(new Date(occ.start))}</span
                                >
                                <span class="calendar-widget-event-title"
                                    >{occ.title}</span
                                >
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        {/if}

        {#if error}
            <p class="calendar-widget-error">{error}</p>
        {/if}
    </div>
</div>

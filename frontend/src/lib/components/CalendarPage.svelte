<script>
    /*

Calendar page: a month grid on the left, the selected day's events on the
right.

Keyboard (same tabs <-> grid scheme as the Workouts page):
- h/j/k/l move the day cursor (k on the top row hands back to the tabs),
  [ ] change month, enter moves into the day's event list
- in the event list: j/k move, enter toggles done, e edits, x deletes,
  q or esc go back to the grid. x on a repeating event asks whether to
  delete just that occurrence (o) or the whole series (s)
- a adds an event on the selected day from either place; while the form
  (EventForm.svelte) is open, every key goes to it

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import {
        buildMonthGrid,
        endOfDay,
        isSameDay,
        moveDayCursor,
        WEEKDAY_LABELS,
    } from "../calendarGrid.js";
    import { formatOccurrenceTime, occurrenceKey } from "../calendarDisplay.js";
    import { emptyForm, eventToForm } from "../eventForm.js";
    import EventForm from "./EventForm.svelte";
    import {
        ListCalendarOccurrences,
        ListEvents,
        ToggleEventCompletion,
        AddEvent,
        UpdateEvent,
        DeleteEvent,
        SkipEventOccurrence,
    } from "../../../wailsjs/go/main/App.js";

    const today = new Date();

    let viewedYear = today.getFullYear();
    let viewedMonth = today.getMonth();
    let cells = buildMonthGrid(viewedYear, viewedMonth);
    let cursor = Math.max(
        0,
        cells.findIndex((c) => isSameDay(c.date, today)),
    );
    let eventCursor = 0;
    let occurrences = [];
    let loading = true;
    let error = null;
    let listMode = false;
    // { form, editingId } while the event form is open, else null.
    let formState = null;
    let formRef;
    // The occurrence awaiting an "this one or the whole series?" answer.
    let pendingDelete = null;

    async function fetchOccurrences() {
        loading = true;

        try {
            const rangeStart = cells[0].date;
            const rangeEnd = endOfDay(cells[cells.length - 1].date);
            occurrences = await ListCalendarOccurrences(rangeStart, rangeEnd);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    }

    async function showMonth(year, month) {
        viewedYear = year;
        viewedMonth = month;
        cells = buildMonthGrid(viewedYear, viewedMonth);

        await fetchOccurrences();
    }

    async function changeMonth(delta) {
        const first = new Date(viewedYear, viewedMonth + delta, 1);
        await showMonth(first.getFullYear(), first.getMonth());

        cursor = Math.max(
            0,
            cells.findIndex((c) => c.inCurrentMonth),
        );
    }

    // Moves the day cursor to `date`, switching months if it isn't in the
    // grid (e.g. an event was saved onto a date in another month).
    async function goToDate(date) {
        if (!cells.some((c) => isSameDay(c.date, date))) {
            await showMonth(date.getFullYear(), date.getMonth());
        }
        cursor = Math.max(
            0,
            cells.findIndex((c) => isSameDay(c.date, date)),
        );
    }

    async function toggleCompletion(occ) {
        try {
            await ToggleEventCompletion(occ.eventId, occ.originalStart);
            await fetchOccurrences();
        } catch (e) {
            error = String(e);
        }
    }

    // ---- event form ----
    function openCreateForm() {
        error = null;
        // "a" also works from the tab bar, where App would otherwise keep
        // h/l/j for tab switching instead of passing them to the form.
        mode.set("grid");
        formState = { form: emptyForm(cells[cursor].date), editingId: null };
    }

    async function openEditForm(occ) {
        error = null;

        try {
            const event = (await ListEvents()).find((e) => e.id === occ.eventId);
            if (!event) {
                error = "That event no longer exists";
                await fetchOccurrences();
                return;
            }
            formState = { form: eventToForm(event), editingId: event.id };
        } catch (e) {
            error = String(e);
        }
    }

    function closeForm() {
        formState = null;
        document.activeElement?.blur();
    }

    // Called by EventForm with a validated calendar.EventInput. Errors
    // propagate back to the form, which shows them and stays open.
    async function saveForm(input) {
        if (formState.editingId) {
            await UpdateEvent(formState.editingId, input);
        } else {
            await AddEvent(input);
        }

        closeForm();
        await goToDate(input.start);
        await fetchOccurrences();
    }

    // ---- delete ----
    async function deleteOccurrence(occ, wholeSeries) {
        pendingDelete = null;

        try {
            if (wholeSeries) {
                await DeleteEvent(occ.eventId);
            } else {
                await SkipEventOccurrence(occ.eventId, occ.originalStart);
            }
            await fetchOccurrences();
        } catch (e) {
            error = String(e);
        }
    }

    function requestDelete(occ) {
        if (occ.recurring) {
            pendingDelete = occ;
        } else {
            deleteOccurrence(occ, true);
        }
    }

    function handleKey(event) {
        if (formState) {
            formRef?.handleKey(event);
            return;
        }

        if (pendingDelete) {
            event.preventDefault();
            if (event.key === "o") {
                deleteOccurrence(pendingDelete, false);
            } else if (event.key === "s") {
                deleteOccurrence(pendingDelete, true);
            } else if (event.key === "Escape" || event.key === "q") {
                pendingDelete = null;
            }
            return;
        }

        if (event.key === "a") {
            event.preventDefault();
            openCreateForm();
            return;
        }

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
                    if (selectedDayOccurrences.length > 0) {
                        listMode = true;
                        eventCursor = 0;
                    }
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
            return;
        }

        const selected = selectedDayOccurrences[eventCursor];

        switch (event.key) {
            case "q":
            case "Escape":
                event.preventDefault();
                listMode = false;
                break;

            case "Enter":
                event.preventDefault();
                if (selected) toggleCompletion(selected);
                break;
            case "k":
                event.preventDefault();
                eventCursor = Math.max(eventCursor - 1, 0);
                break;
            case "j":
                event.preventDefault();
                eventCursor = Math.min(
                    eventCursor + 1,
                    selectedDayOccurrences.length - 1,
                );
                break;
            case "e":
                event.preventDefault();
                if (selected) openEditForm(selected);
                break;
            case "x":
                event.preventDefault();
                if (selected) requestDelete(selected);
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
    // Keep the list cursor on a real item as the list shrinks (deletes,
    // edits moving an event to another day), and leave list mode once
    // there's nothing left to select.
    $: if (eventCursor > selectedDayOccurrences.length - 1) {
        eventCursor = Math.max(0, selectedDayOccurrences.length - 1);
    }
    $: if (listMode && selectedDayOccurrences.length === 0) {
        listMode = false;
    }
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
            <kbd>enter</kbd> events <kbd>a</kbd> add
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
                                        {occ.done ? "[✓]" : "[ ]"}{occ.title}
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
            {#if formState}
                {#key formState}
                    <EventForm
                        bind:this={formRef}
                        initial={formState.form}
                        editing={formState.editingId !== null}
                        onSave={saveForm}
                        onCancel={closeForm}
                    />
                {/key}
            {:else}
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
                                {#if occ.recurring}
                                    <span
                                        class="calendar-detail-repeat"
                                        title="Repeats">↻</span
                                    >
                                {/if}
                            </li>
                        {/each}
                    </ul>
                {/if}

                {#if pendingDelete}
                    <div class="calendar-delete-prompt">
                        Delete “{pendingDelete.title}”?
                        <span>
                            <kbd>o</kbd> this occurrence
                            <kbd>s</kbd> whole series
                            <kbd>esc</kbd> cancel
                        </span>
                    </div>
                {:else}
                    <div class="tasks-hint calendar-detail-hint">
                        <kbd>a</kbd> add
                        {#if listMode}
                            <kbd>e</kbd> edit <kbd>x</kbd> delete
                            <kbd>enter</kbd> done
                        {/if}
                    </div>
                {/if}
            {/if}
        </div>

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    </div>
</div>

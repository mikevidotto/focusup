<script>
    /*

Habits page for habit building.

                             m   t   w   t  *f*  s   s
Wind-down at 8pm            [x] [ ] [x] [x] [ ] [ ] [ ]
Wake up at 5:30             [x] [x] [x] [x] [x] [ ] [ ]
Go for a morning walk       [x] [x] [ ] [x] [x] [ ] [ ]

Each habit is a single record that keeps the list of days ("YYYY-MM-DD")
it was completed on. A row's boxes are the seven dates of the viewed week
(Mon..Sun), checked when that date is in the habit's completions, so past
weeks come for free by moving the viewed week with [ / ].

Keyboard:
- list mode (keyboard mode "tabs", so h/l still switch tabs):
  j/k move between habits, enter drills into the selected habit,
  a add, e edit name, x x delete, [ ] previous/next week
- drilled in (keyboard mode "grid", so this page owns h/l):
  h/l move across the weekday boxes, enter/space toggle that day,
  j/k jump to the same day on the next/previous habit, q/esc back to the list
- either mode: v cycles the history chart (weekly / monthly / year)

The history panel always shows the habit under the cursor.

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import {
        toDateKey,
        isDoneOn,
        countDoneInWeek,
        formatWeekRange,
    } from "../habitDisplay.js";
    import { HISTORY_VIEWS } from "../habitHistory.js";
    import HabitHistory from "./HabitHistory.svelte";
    import {
        buildWeekGrid,
        isSameDay,
        startOfDay,
        WEEKDAY_LABELS,
    } from "../calendarGrid.js";
    import {
        ListHabits,
        AddHabit,
        RenameHabit,
        ToggleHabitCompletion,
        DeleteHabit,
    } from "../../../wailsjs/go/main/App.js";

    const today = startOfDay(new Date());

    let habits = [];
    let cursor = 0;
    let dayCursor = 0;
    let weekOffset = 0;
    // null | "add" | "edit"
    let inputMode = null;
    let inputValue = "";
    let inputEl;
    let pendingDeleteId = null;
    let rowEls = [];
    let historyView = "week";
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            habits = await ListHabits();
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }

        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
    });

    // Drilled-in state lives in the shared keyboard mode rather than a local
    // flag, so anything that resets the mode (e.g. a digit tab shortcut)
    // also pops us back out to the list.
    $: drilled = $mode === "grid";

    $: weekCells = buildWeekGrid(
        new Date(
            today.getFullYear(),
            today.getMonth(),
            today.getDate() + weekOffset * 7,
        ),
    );
    $: todayCol = weekCells.findIndex((cell) => isSameDay(cell.date, today));
    $: doneToday = habits.filter((h) => isDoneOn(h, today)).length;

    $: rowEls[cursor]?.scrollIntoView({ block: "nearest" });

    function isFuture(date) {
        return date > today;
    }

    function drillIn() {
        dayCursor = todayCol >= 0 ? todayCol : 0;
        mode.set("grid");
    }

    function drillOut() {
        mode.set("tabs");
    }

    function changeWeek(delta) {
        weekOffset = Math.min(weekOffset + delta, 0);
    }

    async function toggleDay(habit, cell) {
        if (!habit || !cell || isFuture(cell.date)) {
            return;
        }

        try {
            const updated = await ToggleHabitCompletion(
                habit.id,
                toDateKey(cell.date),
            );
            habits = habits.map((h) => (h.id === updated.id ? updated : h));
        } catch (e) {
            error = String(e);
        }
    }

    async function deleteHabit(habit) {
        try {
            await DeleteHabit(habit.id);
            habits = habits.filter((h) => h.id !== habit.id);
            cursor = Math.max(0, Math.min(cursor, habits.length - 1));
        } catch (e) {
            error = String(e);
        }
    }

    async function openInput(kind) {
        inputMode = kind;
        inputValue = kind === "edit" ? habits[cursor].name : "";
        await tick();
        inputEl?.focus();
        inputEl?.select();
    }

    function closeInput() {
        inputMode = null;
        inputValue = "";
        inputEl?.blur();
    }

    async function submitInput() {
        const name = inputValue.trim();

        if (name) {
            try {
                if (inputMode === "edit") {
                    const updated = await RenameHabit(habits[cursor].id, name);
                    habits = habits.map((h) =>
                        h.id === updated.id ? updated : h,
                    );
                } else {
                    const created = await AddHabit(name);
                    habits = [...habits, created];
                    cursor = habits.length - 1;
                }
                error = null;
            } catch (e) {
                error = String(e);
            }
        }

        closeInput();
    }

    function onInputKeydown(event) {
        if (event.key === "Enter") {
            event.preventDefault();
            submitInput();
        } else if (event.key === "Escape") {
            event.preventDefault();
            closeInput();
        } else if (event.key === "Tab") {
            event.preventDefault();
        }
    }

    function moveCursor(delta) {
        cursor = Math.max(0, Math.min(cursor + delta, habits.length - 1));
    }

    function handleKey(event) {
        if (inputMode) {
            return;
        }

        // "x" arms a delete; the very next key either confirms it (x) or
        // cancels it (anything else, which is then handled normally).
        if (pendingDeleteId) {
            const id = pendingDeleteId;
            pendingDeleteId = null;

            if (event.key === "x") {
                event.preventDefault();
                const habit = habits.find((h) => h.id === id);
                if (habit) {
                    deleteHabit(habit);
                }
                return;
            }
        }

        if (event.key === "v") {
            event.preventDefault();
            const next = HISTORY_VIEWS.indexOf(historyView) + 1;
            historyView = HISTORY_VIEWS[next % HISTORY_VIEWS.length];
            return;
        }

        if (event.key === "[" || event.key === "]") {
            event.preventDefault();
            changeWeek(event.key === "[" ? -1 : 1);
            return;
        }

        if (drilled) {
            handleDrilledKey(event);
        } else {
            handleListKey(event);
        }
    }

    function handleListKey(event) {
        if (event.key === "a") {
            event.preventDefault();
            openInput("add");
            return;
        }

        if (habits.length === 0) {
            return;
        }

        switch (event.key) {
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
                drillIn();
                break;

            case "e":
                event.preventDefault();
                openInput("edit");
                break;

            case "x":
                event.preventDefault();
                pendingDeleteId = habits[cursor].id;
                break;
        }
    }

    function handleDrilledKey(event) {
        if (habits.length === 0) {
            drillOut();
            return;
        }

        switch (event.key) {
            case "h":
                event.preventDefault();
                dayCursor = Math.max(dayCursor - 1, 0);
                break;

            case "l":
                event.preventDefault();
                dayCursor = Math.min(dayCursor + 1, weekCells.length - 1);
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
            case " ":
                event.preventDefault();
                toggleDay(habits[cursor], weekCells[dayCursor]);
                break;

            case "q":
            case "Escape":
                event.preventDefault();
                drillOut();
                break;
        }
    }
</script>

<div class="page-placeholder habits-page">
    <span class="eyebrow">FOCUSUP / HABITS</span>

    <div class="tasks-heading-row">
        <h1>Habits</h1>
        <span class="todo-header-count"
            >{doneToday}/{habits.length} done today</span
        >
    </div>

    <div class="tasks-hint">
        {#if drilled}
            <kbd>h</kbd><kbd>l</kbd> move day
            <kbd>j</kbd><kbd>k</kbd> move habit
            <kbd>enter</kbd> toggle
            <kbd>[</kbd><kbd>]</kbd> week
            <kbd>v</kbd> history view
            <kbd>q</kbd> back
        {:else}
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> open
            <kbd>a</kbd> add
            <kbd>e</kbd> edit
            <kbd>x</kbd> delete
            <kbd>[</kbd><kbd>]</kbd> week
            <kbd>v</kbd> history view
        {/if}
    </div>

    <div class="habits-layout">
        <div class="habits-main">
            <div class="habits-week-label">
                {formatWeekRange(weekCells)}
                {#if weekOffset === 0}
                    <span class="habits-week-tag">this week</span>
                {:else}
                    <span class="habits-week-tag"
                        >{-weekOffset} week{weekOffset === -1 ? "" : "s"} ago</span
                    >
                {/if}
            </div>

            {#if loading}
                <p>Loading habits…</p>
            {:else if habits.length === 0}
                <div class="empty-widget">
                    <span>空</span>
                    <p>No habits yet — press a to add one</p>
                </div>
            {:else}
                <div class="habits-grid">
                    <div class="habits-row habits-header">
                        <span class="habits-name"></span>
                        {#each weekCells as cell (cell.col)}
                            <span
                                class="habits-day-label"
                                class:today={cell.col === todayCol}
                            >
                                <span>{WEEKDAY_LABELS[cell.col]}</span>
                                <span class="habits-day-number"
                                    >{cell.date.getDate()}</span
                                >
                            </span>
                        {/each}
                        <span class="habits-count"></span>
                    </div>

                    <ul class="habits-list">
                        {#each habits as habit, index (habit.id)}
                            {@const selected = inputMode !== "add" && index === cursor}
                            <li
                                class="habits-row"
                                class:cursor={selected}
                                class:drilled={selected && drilled}
                                bind:this={rowEls[index]}
                            >
                                <span class="habits-name" title={habit.name}>
                                    {#if pendingDeleteId === habit.id}
                                        <span class="habits-delete-confirm"
                                            >x again to delete</span
                                        >
                                    {:else}
                                        {habit.name}
                                    {/if}
                                </span>

                                {#each weekCells as cell (cell.col)}
                                    {@const done = isDoneOn(habit, cell.date)}
                                    <button
                                        type="button"
                                        class="habits-box"
                                        class:done
                                        class:today={cell.col === todayCol}
                                        class:future={isFuture(cell.date)}
                                        class:cursor={selected &&
                                            drilled &&
                                            cell.col === dayCursor}
                                        disabled={isFuture(cell.date)}
                                        on:click={() => toggleDay(habit, cell)}
                                    >
                                        {done ? "✓" : ""}
                                    </button>
                                {/each}

                                <span class="habits-count"
                                    >{countDoneInWeek(habit, weekCells)}/7</span
                                >
                            </li>
                        {/each}
                    </ul>
                </div>
            {/if}

            <div class="tasks-add-row">
                <input
                    class="todo-input"
                    type="text"
                    bind:this={inputEl}
                    bind:value={inputValue}
                    placeholder={inputMode === "edit"
                        ? "rename habit…"
                        : "press a to add a habit…"}
                    on:keydown={onInputKeydown}
                    on:focus={() => (inputMode = inputMode || "add")}
                    on:blur={closeInput}
                />
            </div>

            {#if error}
                <p class="todo-error">{error}</p>
            {/if}
        </div>

        {#if habits[cursor]}
            <HabitHistory
                habit={habits[cursor]}
                {today}
                view={historyView}
                onSelectView={(v) => (historyView = v)}
            />
        {/if}
    </div>
</div>

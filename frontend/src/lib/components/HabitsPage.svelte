<script>
    /* 

Habits page for habit building.

This page will be very similar to the Tasks page. 

It should have:
- A list of Habits that you are keeping track of
- A section to check off whether a habit was successfully completed for a given day of the week
- A "history" section where you can see which days you had completed the habitual activity.


For example:
                             m   t   w   t  *f*  s   s
Wind-down at 8pm            [x] [ ] [x] [x] [ ] [ ] [ ]
Wake up at 5:30             [x] [x] [x] [x] [x] [ ] [ ]
Go for a morning walk       [x] [x] [ ] [x] [x] [ ] [ ]
Brush your teeth            [x] [x] [x] [x] [x] [ ] [ ]



---------------------
keyboard navigation 
---------------------
should behave similar to any list, where we can traverse vertically, 
and then when we hit enter, we focus in on that habit,
and we should be able to move horizontally across the days of the current week for that habit, 
then use enter to toggle whether the habit has been successfully completed for that day.
'q' should be the quit key

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        splitHabits,
    } from "../habitDisplay.js";
    import {
        ListHabits,
        AddHabit,
        ToggleHabit,
        DeleteHabit,
    } from "../../../wailsjs/go/main/App.js";

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

    let habits = [];
    let cursor = 0;
    let insertMode = false;
    let newTitle = "";
    let inputEl;
    let loading = true;
    let error = null;

    async function toggleCurrent() {
        const habit = rows[cursor];

        if (!habit) {
            return;
        }

        try {
            const updated = await ToggleHabit(habit.id);
            habits = habits.map((t) => (t.id === updated.id ? updated : t));
        } catch (e) {
            error = String(e);
        }
    }

    async function deleteCurrent() {
        const habit = rows[cursor];

        if (!habit) {
            return;
        }

        try {
            await DeleteHabit(habit.id);

            const updated = habits.filter((t) => t.id !== habit.id);
            habits = updated;
            cursor = Math.max(0, Math.min(cursor, updated.length - 1));
        } catch (e) {
            error = String(e);
        }
    }

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
                    const created = await AddHabit(title);
                    habits = [...habits, created];
                } catch (e) {
                    error = String(e);
                }
            }
            exitInsertMode();
        } else if (event.key === "Escape") {
            event.preventDefault();
            exitInsertMode();
        } else if (event.key === "Tab") {
            event.preventDefault();
            cyclePendingPriority();
        }
    }

    function handleKey(event) {
        if (insertMode) {
            return;
        }

        if (event.key === "a") {
            event.preventDefault();
            enterInsertMode();
            return;
        }

        if (rows.length === 0) {
            return;
        }

        switch (event.key) {
            case "j":
                event.preventDefault();
                cursor = Math.min(cursor + 1, rows.length - 1);
                break;

            case "k":
                event.preventDefault();
                cursor = Math.max(cursor - 1, 0);
                break;

            case "Enter":
            case " ":
                event.preventDefault();
                toggleCurrent();
                break;

            case "x":
                event.preventDefault();
                deleteCurrent();
                break;
        }
    }

    $: ({ active, completed } = splitHabits(habits));
    $: rows = [...active, ...completed];
</script>

<div class="page-placeholder habits-page">
    <span class="eyebrow">FOCUSUP / HABITS</span>

    <div class="habits-heading-row">
        <h1>Habits</h1>
        <span class="todo-header-count"
            >{active.length} active • {habits.length} total</span
        >
    </div>

    <div class="habits-hint">
        <kbd>j</kbd><kbd>k</kbd> move
        <kbd>enter</kbd> toggle
        <kbd>a</kbd> add
        <kbd>x</kbd> delete
    </div>

    {#if loading}
        <p>Loading habits…</p>
    {:else}
        {#if habits.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No habits yet — press a to add one</p>
            </div>
        {:else}
            <ul class="todo-list">
                {#each active as habit (habit.id)}
                    {@const index = rows.indexOf(habit)}
                    <li
                        class="todo-item"
                        class:cursor={!insertMode && index === cursor}
                    >
                        <span class="todo-mark">☐</span>
                        <span class="todo-title">{habit.name}</span>
                    </li>
                {/each}
            </ul>
        {/if}

        <div class="habits-add-row">
            <input
                class="todo-input"
                type="text"
                bind:this={inputEl}
                bind:value={newTitle}
                placeholder="press a to add a habit…"
                on:keydown={onInputKeydown}
                on:focus={() => (insertMode = true)}
                on:blur={exitInsertMode}
            />

            {#if insertMode}
                <button
                    type="button"
                    class="priority-picker"
                    style="color: {PRIORITY_META[pendingPriority].color}"
                    on:mousedown|preventDefault={cyclePendingPriority}
                    title="Press Tab to cycle priority"
                >
                    {PRIORITY_META[pendingPriority].icon}
                    {PRIORITY_META[pendingPriority].label}
                </button>
            {/if}
        </div>

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    {/if}
</div>

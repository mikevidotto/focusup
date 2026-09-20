<script>
<<<<<<< HEAD
=======
    import { onMount, onDestroy, tick } from "svelte";
>>>>>>> parent of 9aae43e (nav modifications for switching between tab bar and tasks page)

    import { onMount, onDestroy, tick } from "svelte";

    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        PRIORITY_META,
        nextPriority,
        splitTasks,
        formatCompletedDate,
    } from "../taskDisplay.js";
    import {
        ListTasks,
        AddTask,
        ToggleTask,
        DeleteTask,
    } from "../../../wailsjs/go/main/App.js";

    let tasks = [];
    let cursor = 0;
    let insertMode = false;
    let newTitle = "";
    let pendingPriority = "medium";
    let inputEl;
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            tasks = await ListTasks();
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

    $: ({ active, completed } = splitTasks(tasks));
    $: rows = [...active, ...completed];

    async function toggleCurrent() {
        const task = rows[cursor];

        if (!task) {
            return;
        }

        try {
            const updated = await ToggleTask(task.id);
            tasks = tasks.map((t) => (t.id === updated.id ? updated : t));
        } catch (e) {
            error = String(e);
        }
    }

    async function deleteCurrent() {
        const task = rows[cursor];

        if (!task) {
            return;
        }

        try {
            await DeleteTask(task.id);

            const updated = tasks.filter((t) => t.id !== task.id);
            tasks = updated;
            cursor = Math.max(0, Math.min(cursor, updated.length - 1));
        } catch (e) {
            error = String(e);
        }
    }

    async function enterInsertMode() {
        insertMode = true;
        pendingPriority = "medium";
        await tick();
        inputEl?.focus();
    }

    function exitInsertMode() {
        insertMode = false;
        newTitle = "";
        inputEl?.blur();
    }

    function cyclePendingPriority() {
        pendingPriority = nextPriority(pendingPriority);
    }

    async function onInputKeydown(event) {
        if (event.key === "Enter") {
            event.preventDefault();

            const title = newTitle.trim();

            if (title) {
                try {
                    const created = await AddTask(title, pendingPriority);
                    tasks = [...tasks, created];
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
<<<<<<< HEAD
<<<<<<< HEAD
=======
=======

                if (cursor === 0) {
                    mode.set("tabs");
                } else {
                }
>>>>>>> parent of 9aae43e (nav modifications for switching between tab bar and tasks page)

                if (cursor === 0) {
                    mode.set("tabs");
                } else {
                }

>>>>>>> parent of 9aae43e (nav modifications for switching between tab bar and tasks page)
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
</script>

<div class="page-placeholder tasks-page">
    <span class="eyebrow">FOCUSUP / TASKS</span>

    <div class="tasks-heading-row">
        <h1>Tasks</h1>
        <span class="todo-header-count"
            >{active.length} active • {tasks.length} total</span
        >
    </div>

    <div class="tasks-hint">
        <kbd>j</kbd><kbd>k</kbd> move
        <kbd>enter</kbd> toggle
        <kbd>a</kbd> add
        <kbd>x</kbd> delete
    </div>

    {#if loading}
        <p>Loading tasks…</p>
    {:else}
        {#if tasks.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No tasks yet — press a to add one</p>
            </div>
        {:else}
            <ul class="todo-list">
                {#each active as task (task.id)}
                    {@const index = rows.indexOf(task)}
<<<<<<< HEAD
<<<<<<< HEAD
                    <li
                        class="todo-item"
                        class:cursor={!insertMode && index === cursor}
                    >
=======
                    <li class="todo-item" class:cursor={!insertMode && index === cursor}>
>>>>>>> parent of 9aae43e (nav modifications for switching between tab bar and tasks page)
=======
                    <li class="todo-item" class:cursor={!insertMode && index === cursor}>
>>>>>>> parent of 9aae43e (nav modifications for switching between tab bar and tasks page)
                        <span class="todo-mark">☐</span>
                        <span class="todo-title">{task.title}</span>
                        <span
                            class="todo-priority"
                            style="color: {PRIORITY_META[task.priority].color}"
                            title="{PRIORITY_META[task.priority]
                                .label} priority"
                        >
                            {PRIORITY_META[task.priority].icon}
                        </span>
                    </li>
                {/each}
            </ul>

            {#if completed.length > 0}
                <div class="todo-section-label">Completed</div>

                <ul class="todo-list">
                    {#each completed as task (task.id)}
                        {@const index = rows.indexOf(task)}
                        <li
                            class="todo-item done"
                            class:cursor={!insertMode && index === cursor}
                        >
                            <span class="todo-mark todo-mark-done">☑</span>
                            <span class="todo-title">{task.title}</span>
                            <span class="todo-completed-date"
                                >{formatCompletedDate(task.completedAt)}</span
                            >
                        </li>
                    {/each}
                </ul>
            {/if}
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

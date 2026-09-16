<script>
    import { onMount, tick } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import {
        ListTasks,
        AddTask,
        ToggleTask,
        DeleteTask
    } from "../../../../wailsjs/go/main/App.js";

    export let focused = false;

    let tasks = [];
    let cursor = 0;
    let insertMode = false;
    let newTitle = "";
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
    });

    async function toggleCurrent() {
        const task = tasks[cursor];

        if (!task) {
            return;
        }

        try {
            const updated = await ToggleTask(task.id);
            tasks = tasks.map(t => (t.id === updated.id ? updated : t));
        } catch (e) {
            error = String(e);
        }
    }

    async function deleteCurrent() {
        const task = tasks[cursor];

        if (!task) {
            return;
        }

        try {
            await DeleteTask(task.id);
            tasks = tasks.filter(t => t.id !== task.id);
            cursor = Math.max(0, Math.min(cursor, tasks.length - 1));
        } catch (e) {
            error = String(e);
        }
    }

    async function enterInsertMode() {
        insertMode = true;
        await tick();
        inputEl?.focus();
    }

    function exitInsertMode() {
        insertMode = false;
        newTitle = "";
        inputEl?.blur();
    }

    async function onInputKeydown(event) {
        if (event.key === "Enter") {
            event.preventDefault();

            const title = newTitle.trim();

            if (title) {
                try {
                    const created = await AddTask(title);
                    tasks = [...tasks, created];
                    cursor = tasks.length - 1;
                } catch (e) {
                    error = String(e);
                }
            }

            exitInsertMode();
        } else if (event.key === "Escape") {
            event.preventDefault();
            exitInsertMode();
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

        if (tasks.length === 0) {
            return;
        }

        switch (event.key) {
            case "j":
                event.preventDefault();
                cursor = Math.min(cursor + 1, tasks.length - 1);
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

    $: if (focused) {
        activeWidgetKeyHandler.set(handleKey);
    } else {
        activeWidgetKeyHandler.set(null);

        if (insertMode) {
            exitInsertMode();
        }
    }
</script>

<div class="todo-widget">
    {#if loading}
        <div class="empty-widget">
            <span>…</span>
            <p>Loading tasks</p>
        </div>
    {:else}
        {#if tasks.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No tasks yet{focused ? " — press a to add one" : ""}</p>
            </div>
        {:else}
            <ul class="todo-list">
                {#each tasks as task, index (task.id)}
                    <li
                        class="todo-item"
                        class:done={task.done}
                        class:cursor={focused && !insertMode && index === cursor}
                    >
                        <span class="todo-mark">{task.done ? "☑" : "☐"}</span>
                        <span class="todo-title">{task.title}</span>
                    </li>
                {/each}
            </ul>
        {/if}

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

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    {/if}
</div>

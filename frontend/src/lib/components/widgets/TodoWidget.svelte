<script>
    import { onMount } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import { PRIORITY_META, splitTasks, formatCompletedDate } from "../../taskDisplay.js";
    import { ListTasks, ToggleTask } from "../../../../wailsjs/go/main/App.js";

    export let focused = false;

    const MAX_ACTIVE_VISIBLE = 5;
    const MAX_COMPLETED_VISIBLE = 2;

    let tasks = [];
    let cursor = 0;
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

    $: ({ active, completed } = splitTasks(tasks));
    $: visibleActive = active.slice(0, MAX_ACTIVE_VISIBLE);
    $: hiddenActiveCount = active.length - visibleActive.length;
    $: visibleCompleted = completed.slice(0, MAX_COMPLETED_VISIBLE);
    $: hiddenCompletedCount = completed.length - visibleCompleted.length;
    $: rows = [...visibleActive, ...visibleCompleted];

    async function toggleCurrent() {
        const task = rows[cursor];

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

    function handleKey(event) {
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
        }
    }

    $: if (focused) {
        activeWidgetKeyHandler.set(handleKey);
    } else {
        activeWidgetKeyHandler.set(null);
    }
</script>

<div class="todo-widget">
    <header class="todo-header">
        <span class="todo-header-title">Tasks</span>
        <span class="todo-header-count">{active.length} active • {tasks.length} total</span>
    </header>

    <div class="todo-body">
        {#if loading}
            <div class="empty-widget">
                <span>…</span>
                <p>Loading tasks</p>
            </div>
        {:else if tasks.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No tasks yet — add some on the Tasks page</p>
            </div>
        {:else}
            <ul class="todo-list">
                {#each visibleActive as task (task.id)}
                    {@const index = rows.indexOf(task)}
                    <li class="todo-item" class:cursor={focused && index === cursor}>
                        <span class="todo-mark">☐</span>
                        <span class="todo-title">{task.title}</span>
                        <span
                            class="todo-priority"
                            style="color: {PRIORITY_META[task.priority].color}"
                            title="{PRIORITY_META[task.priority].label} priority"
                        >
                            {PRIORITY_META[task.priority].icon}
                        </span>
                    </li>
                {/each}

                {#if hiddenActiveCount > 0}
                    <li class="todo-more-hint">+{hiddenActiveCount} more on the Tasks page</li>
                {/if}
            </ul>

            {#if completed.length > 0}
                <div class="todo-section-label">Completed</div>

                <ul class="todo-list">
                    {#each visibleCompleted as task (task.id)}
                        {@const index = rows.indexOf(task)}
                        <li class="todo-item done" class:cursor={focused && index === cursor}>
                            <span class="todo-mark todo-mark-done">☑</span>
                            <span class="todo-title">{task.title}</span>
                            <span class="todo-completed-date">{formatCompletedDate(task.completedAt)}</span>
                        </li>
                    {/each}

                    {#if hiddenCompletedCount > 0}
                        <li class="todo-more-hint">+{hiddenCompletedCount} more completed</li>
                    {/if}
                </ul>
            {/if}
        {/if}

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    </div>
</div>

<script>
    import { onMount } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import { splitTasks, countByList, isDoneToday } from "../../taskDisplay.js";
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

    // Only next actions belong on the dashboard; the inbox is processed on
    // the Tasks page, so it just shows up as a count here.
    $: counts = countByList(tasks);
    // Only today's completions show here; older ones live in the Tasks
    // page's Done list.
    $: ({ active: open, completed: allCompleted } = splitTasks(tasks));
    $: completed = allCompleted.filter(t => isDoneToday(t));
    $: active = open.filter(t => t.list === "next");
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
        <span class="todo-header-count">{counts.next} next • {counts.inbox} in inbox</span>
    </header>

    <div class="todo-body">
        {#if loading}
            <div class="empty-widget">
                <span>…</span>
                <p>Loading tasks</p>
            </div>
        {:else if active.length === 0 && completed.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No next actions — process your inbox on the Tasks page</p>
            </div>
        {:else}
            <ul class="todo-list">
                {#each visibleActive as task (task.id)}
                    {@const index = rows.indexOf(task)}
                    <li class="todo-item" class:cursor={focused && index === cursor}>
                        <span class="todo-mark">☐</span>
                        <span class="todo-title">{task.title}</span>
                        {#each task.contexts as context}
                            <span class="gtd-context">@{context}</span>
                        {/each}
                    </li>
                {/each}

                {#if hiddenActiveCount > 0}
                    <li class="todo-more-hint">+{hiddenActiveCount} more on the Tasks page</li>
                {/if}
            </ul>

            {#if completed.length > 0}
                <div class="todo-section-label">Done today</div>

                <ul class="todo-list">
                    {#each visibleCompleted as task (task.id)}
                        {@const index = rows.indexOf(task)}
                        <li class="todo-item done" class:cursor={focused && index === cursor}>
                            <span class="todo-mark todo-mark-done">☑</span>
                            <span class="todo-title">{task.title}</span>
                        </li>
                    {/each}

                    {#if hiddenCompletedCount > 0}
                        <li class="todo-more-hint">+{hiddenCompletedCount} more done today</li>
                    {/if}
                </ul>
            {/if}
        {/if}

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    </div>
</div>

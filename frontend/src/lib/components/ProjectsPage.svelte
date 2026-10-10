<script>
    /*

Projects page: the GTD project list.

A project is any outcome that takes more than one action. Each active
project should always have at least one open next action, so projects
without one are flagged. The panel on the right lists the tasks of the
project under the cursor; tasks are linked from the Tasks page (p) or
added here directly (n).

Keyboard:
- j/k move, enter toggle done
- a add project, e rename, x x delete (its tasks are kept, unlinked)
- n add a next action to the selected project ("@tag" words become contexts)

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        splitTasks,
        formatCompletedDate,
        parseCapture,
        nextActionCount,
    } from "../taskDisplay.js";
    import {
        ListTasks,
        ListProjects,
        AddProject,
        RenameProject,
        ToggleProject,
        DeleteProject,
        AddProjectTask,
    } from "../../../wailsjs/go/main/App.js";

    let projects = [];
    let tasks = [];
    let cursor = 0;
    // null | "add" | "edit" | "task"
    let inputMode = null;
    let inputValue = "";
    let inputEl;
    let pendingDeleteId = null;
    let rowEls = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            [projects, tasks] = await Promise.all([ListProjects(), ListTasks()]);
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

    // Active projects first, finished ones after.
    $: rows = [
        ...projects.filter((p) => !p.done),
        ...projects.filter((p) => p.done),
    ];
    $: if (cursor > rows.length - 1) {
        cursor = Math.max(0, rows.length - 1);
    }
    $: selected = rows[cursor];
    $: selectedTasks = splitTasks(
        tasks.filter((t) => selected && t.projectId === selected.id),
    );
    $: stalledCount = rows.filter(
        (p) => !p.done && nextActionCount(p, tasks) === 0,
    ).length;

    $: rowEls[cursor]?.scrollIntoView({ block: "nearest" });

    async function run(action) {
        try {
            await action();
            error = null;
        } catch (e) {
            error = String(e);
        }
    }

    function replaceProject(updated) {
        projects = projects.map((p) => (p.id === updated.id ? updated : p));
    }

    function deleteProject(id) {
        run(async () => {
            await DeleteProject(id);
            projects = projects.filter((p) => p.id !== id);
            tasks = tasks.map((t) =>
                t.projectId === id ? { ...t, projectId: "" } : t,
            );
        });
    }

    async function openInput(kind) {
        inputMode = kind;
        inputValue = kind === "edit" ? selected.title : "";
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
        const kind = inputMode;
        const value = inputValue.trim();
        const project = selected;
        closeInput();

        if (!value) {
            return;
        }

        await run(async () => {
            if (kind === "edit") {
                replaceProject(await RenameProject(project.id, value));
            } else if (kind === "task") {
                const { title, contexts } = parseCapture(value);
                if (title) {
                    const created = await AddProjectTask(project.id, title, contexts);
                    tasks = [...tasks, created];
                }
            } else {
                const created = await AddProject(value);
                projects = [...projects, created];
                await tick();
                cursor = rows.findIndex((p) => p.id === created.id);
            }
        });
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
                deleteProject(id);
                return;
            }
        }

        if (event.key === "a") {
            event.preventDefault();
            openInput("add");
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
                run(async () => replaceProject(await ToggleProject(selected.id)));
                break;

            case "e":
                event.preventDefault();
                openInput("edit");
                break;

            case "n":
                event.preventDefault();
                openInput("task");
                break;

            case "x":
                event.preventDefault();
                pendingDeleteId = selected.id;
                break;
        }
    }

    const PLACEHOLDERS = {
        add: "new project — what does done look like?",
        edit: "rename project…",
        task: "next action for this project… (@tag adds a context)",
    };
</script>

<div class="page-placeholder tasks-page">
    <span class="eyebrow">FOCUSUP / PROJECTS</span>

    <div class="tasks-heading-row">
        <h1>Projects</h1>
        <span class="todo-header-count">
            {rows.filter((p) => !p.done).length} active
            {#if stalledCount > 0}
                • <span class="gtd-stalled">{stalledCount} need a next action</span>
            {/if}
        </span>
    </div>

    <div class="tasks-hint">
        <kbd>j</kbd><kbd>k</kbd> move
        <kbd>enter</kbd> done
        <kbd>a</kbd> add
        <kbd>n</kbd> next action
        <kbd>e</kbd> rename
        <kbd>x</kbd> delete
    </div>

    <div class="habits-layout">
        <div class="habits-main">
            {#if loading}
                <p>Loading projects…</p>
            {:else if rows.length === 0}
                <div class="empty-widget">
                    <span>空</span>
                    <p>No projects yet — press a to add one</p>
                </div>
            {:else}
                <ul class="todo-list">
                    {#each rows as project, index (project.id)}
                        {@const count = nextActionCount(project, tasks)}
                        <li
                            class="todo-item"
                            class:done={project.done}
                            class:cursor={!inputMode && index === cursor}
                            bind:this={rowEls[index]}
                        >
                            <span class="todo-mark" class:todo-mark-done={project.done}
                                >{project.done ? "☑" : "▸"}</span
                            >
                            <span class="todo-title">
                                {#if pendingDeleteId === project.id}
                                    <span class="habits-delete-confirm"
                                        >x again to delete</span
                                    >
                                {:else}
                                    {project.title}
                                {/if}
                            </span>
                            {#if project.done}
                                <span class="todo-completed-date"
                                    >{formatCompletedDate(project.completedAt)}</span
                                >
                            {:else if count === 0}
                                <span class="gtd-stalled">no next action</span>
                            {:else}
                                <span class="todo-completed-date"
                                    >{count} next</span
                                >
                            {/if}
                        </li>
                    {/each}
                </ul>
            {/if}

            <div class="tasks-add-row">
                <input
                    class="todo-input"
                    type="text"
                    bind:this={inputEl}
                    bind:value={inputValue}
                    placeholder={PLACEHOLDERS[inputMode] ??
                        "press a to add a project…"}
                    on:keydown={onInputKeydown}
                    on:focus={() => (inputMode = inputMode || "add")}
                    on:blur={closeInput}
                />
            </div>

            {#if error}
                <p class="todo-error">{error}</p>
            {/if}
        </div>

        {#if selected}
            <div class="habit-history gtd-project-panel">
                <div class="habit-history-title">
                    <span class="eyebrow">PROJECT</span>
                    <h2>{selected.title}</h2>
                </div>

                {#if selectedTasks.active.length === 0 && selectedTasks.completed.length === 0}
                    <p class="todo-more-hint">
                        No tasks yet — press n to add the next action
                    </p>
                {/if}

                {#if selectedTasks.active.length > 0}
                    <ul class="todo-list">
                        {#each selectedTasks.active as task (task.id)}
                            <li class="todo-item">
                                <span class="todo-mark">☐</span>
                                <span class="todo-title">{task.title}</span>
                                {#each task.contexts as context}
                                    <span class="gtd-context">@{context}</span>
                                {/each}
                                {#if task.list !== "next"}
                                    <span class="todo-completed-date"
                                        >{task.list}</span
                                    >
                                {/if}
                            </li>
                        {/each}
                    </ul>
                {/if}

                {#if selectedTasks.completed.length > 0}
                    <div class="todo-section-label">Completed</div>
                    <ul class="todo-list">
                        {#each selectedTasks.completed as task (task.id)}
                            <li class="todo-item done">
                                <span class="todo-mark todo-mark-done">☑</span>
                                <span class="todo-title">{task.title}</span>
                                <span class="todo-completed-date"
                                    >{formatCompletedDate(task.completedAt)}</span
                                >
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        {/if}
    </div>
</div>

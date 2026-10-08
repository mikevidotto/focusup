<script>
    /*

Tasks page, modeled on GTD (Getting Things Done).

Everything is captured into the Inbox, then clarified: do it now (enter),
make it a Next Action (n), park it in Someday/Maybe (s), or attach it to a
project (p, which also makes it a next action). Next Actions can be
filtered by @context (f). Projects themselves live on the Projects tab.

Keyboard:
- [ / ] switch list (Inbox / Next Actions / Someday)
- j/k move, enter toggle done
- a capture to the inbox ("@tag" words become contexts)
- e edit title and contexts, p pick project, x x delete
- n / s / i move to Next / Someday / Inbox
- f cycle the @context filter (Next Actions only)

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        LISTS,
        splitTasks,
        formatCompletedDate,
        parseCapture,
        formatForEdit,
        allContexts,
        countByList,
    } from "../taskDisplay.js";
    import {
        ListTasks,
        ListProjects,
        AddTask,
        ToggleTask,
        MoveTask,
        RenameTask,
        SetTaskContexts,
        SetTaskProject,
        DeleteTask,
    } from "../../../wailsjs/go/main/App.js";

    const MOVE_KEYS = { n: "next", s: "someday", i: "inbox" };

    let tasks = [];
    let projects = [];
    let listId = "inbox";
    let contextFilter = null;
    let cursor = 0;
    // null | "add" | "edit"
    let inputMode = null;
    let inputValue = "";
    let inputEl;
    let pendingDeleteId = null;
    // Set while the project picker is open for the task under the cursor.
    let pickerOpen = false;
    let pickerCursor = 0;
    let rowEls = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            [tasks, projects] = await Promise.all([ListTasks(), ListProjects()]);

            // Land on the inbox only when there is something to process.
            if (countByList(tasks).inbox === 0) {
                listId = "next";
            }
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

    $: counts = countByList(tasks);
    $: projectById = Object.fromEntries(projects.map((p) => [p.id, p]));
    $: activeProjects = projects.filter((p) => !p.done);
    $: pickerOptions = [null, ...activeProjects];

    $: isNext = listId === "next";
    $: ({ active: open, completed: done } = splitTasks(tasks));
    $: nextContexts = allContexts(open.filter((t) => t.list === "next"));
    $: if (contextFilter && !nextContexts.includes(contextFilter)) {
        contextFilter = null;
    }

    $: matchesFilter = (t) =>
        !isNext || !contextFilter || t.contexts.includes(contextFilter);
    $: active = open.filter((t) => t.list === listId && matchesFilter(t));
    // Done tasks are out of the GTD flow, so they only show under Next Actions.
    $: completed = isNext ? done.filter(matchesFilter) : [];
    $: rows = [...active, ...completed];
    $: if (cursor > rows.length - 1) {
        cursor = Math.max(0, rows.length - 1);
    }

    $: rowEls[cursor]?.scrollIntoView({ block: "nearest" });

    function replaceTask(updated) {
        tasks = tasks.map((t) => (t.id === updated.id ? updated : t));
    }

    async function run(action) {
        try {
            await action();
            error = null;
        } catch (e) {
            error = String(e);
        }
    }

    function switchList(delta) {
        const index = LISTS.findIndex((l) => l.id === listId);
        listId = LISTS[(index + delta + LISTS.length) % LISTS.length].id;
        cursor = 0;
    }

    function cycleContextFilter() {
        const options = [null, ...nextContexts];
        contextFilter =
            options[(options.indexOf(contextFilter) + 1) % options.length];
        cursor = 0;
    }

    function toggleCurrent() {
        const task = rows[cursor];
        if (task) {
            run(async () => replaceTask(await ToggleTask(task.id)));
        }
    }

    function moveCurrent(list) {
        const task = rows[cursor];
        if (task && task.list !== list) {
            run(async () => replaceTask(await MoveTask(task.id, list)));
        }
    }

    function deleteTask(id) {
        run(async () => {
            await DeleteTask(id);
            tasks = tasks.filter((t) => t.id !== id);
        });
    }

    function openPicker() {
        const task = rows[cursor];
        if (!task) {
            return;
        }

        pickerCursor = Math.max(
            0,
            pickerOptions.findIndex((p) => (p?.id ?? "") === task.projectId),
        );
        pickerOpen = true;
    }

    function pickProject() {
        const task = rows[cursor];
        const project = pickerOptions[pickerCursor];
        pickerOpen = false;

        if (task) {
            run(async () =>
                replaceTask(await SetTaskProject(task.id, project?.id ?? "")),
            );
        }
    }

    async function openInput(kind) {
        inputMode = kind;
        inputValue = kind === "edit" ? formatForEdit(rows[cursor]) : "";
        await tick();
        inputEl?.focus();
    }

    function closeInput() {
        inputMode = null;
        inputValue = "";
        inputEl?.blur();
    }

    async function submitInput() {
        const { title, contexts } = parseCapture(inputValue);
        const editing = inputMode === "edit" ? rows[cursor] : null;
        closeInput();

        if (editing) {
            await run(async () => {
                // A bare "@tag" edit retags the task and keeps its title.
                if (title && title !== editing.title) {
                    replaceTask(await RenameTask(editing.id, title));
                }
                replaceTask(await SetTaskContexts(editing.id, contexts));
            });
        } else if (title) {
            await run(async () => {
                tasks = [...tasks, await AddTask(title, contexts)];
            });
        }
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

    function handlePickerKey(event) {
        event.preventDefault();

        switch (event.key) {
            case "j":
                pickerCursor = Math.min(pickerCursor + 1, pickerOptions.length - 1);
                break;

            case "k":
                pickerCursor = Math.max(pickerCursor - 1, 0);
                break;

            case "-":
                pickerCursor = 0;
                pickProject();
                break;

            case "Enter":
            case " ":
                pickProject();
                break;

            case "q":
            case "Escape":
            case "p":
                pickerOpen = false;
                break;
        }
    }

    function handleKey(event) {
        if (inputMode) {
            return;
        }

        if (pickerOpen) {
            handlePickerKey(event);
            return;
        }

        // "x" arms a delete; the very next key either confirms it (x) or
        // cancels it (anything else, which is then handled normally).
        if (pendingDeleteId) {
            const id = pendingDeleteId;
            pendingDeleteId = null;

            if (event.key === "x") {
                event.preventDefault();
                deleteTask(id);
                return;
            }
        }

        switch (event.key) {
            case "a":
                event.preventDefault();
                openInput("add");
                return;

            case "[":
            case "]":
                event.preventDefault();
                switchList(event.key === "[" ? -1 : 1);
                return;

            case "f":
                if (isNext) {
                    event.preventDefault();
                    cycleContextFilter();
                }
                return;
        }

        if (rows.length === 0) {
            return;
        }

        if (MOVE_KEYS[event.key]) {
            event.preventDefault();
            moveCurrent(MOVE_KEYS[event.key]);
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

            case "e":
                event.preventDefault();
                openInput("edit");
                break;

            case "p":
                event.preventDefault();
                openPicker();
                break;

            case "x":
                event.preventDefault();
                pendingDeleteId = rows[cursor].id;
                break;
        }
    }
</script>

<div class="page-placeholder tasks-page">
    <span class="eyebrow">FOCUSUP / TASKS</span>

    <div class="tasks-heading-row">
        <h1>Tasks</h1>
        <span class="todo-header-count"
            >{counts.inbox} in inbox • {counts.next} next</span
        >
    </div>

    <div class="tasks-hint">
        {#if pickerOpen}
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> pick
            <kbd>-</kbd> no project
            <kbd>q</kbd> cancel
        {:else}
            <kbd>[</kbd><kbd>]</kbd> list
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> done
            <kbd>a</kbd> capture
            <kbd>n</kbd><kbd>s</kbd><kbd>i</kbd> next / someday / inbox
            <kbd>p</kbd> project
            <kbd>e</kbd> edit
            <kbd>x</kbd> delete
            {#if isNext}<kbd>f</kbd> context{/if}
        {/if}
    </div>

    <div class="gtd-lists">
        {#each LISTS as list (list.id)}
            <button
                type="button"
                class:active={list.id === listId}
                on:click={() => {
                    listId = list.id;
                    cursor = 0;
                }}
            >
                {list.label}
                <span class="gtd-list-count">{counts[list.id]}</span>
            </button>
        {/each}
    </div>

    {#if isNext && nextContexts.length > 0}
        <div class="gtd-context-filter">
            {#each [null, ...nextContexts] as context}
                <button
                    type="button"
                    class="gtd-context"
                    class:active={context === contextFilter}
                    on:click={() => {
                        contextFilter = context;
                        cursor = 0;
                    }}
                >
                    {context ? "@" + context : "all"}
                </button>
            {/each}
        </div>
    {/if}

    {#if loading}
        <p>Loading tasks…</p>
    {:else}
        {#if rows.length === 0}
            <div class="empty-widget">
                {#if listId === "inbox"}
                    <span>無</span>
                    <p>Inbox zero — press a to capture something</p>
                {:else if listId === "someday"}
                    <span>空</span>
                    <p>Nothing parked — press s on a task to move it here</p>
                {:else}
                    <span>空</span>
                    <p>No next actions — process your inbox with n</p>
                {/if}
            </div>
        {:else}
            <ul class="todo-list">
                {#each active as task, index (task.id)}
                    <li
                        class="todo-item"
                        class:cursor={!inputMode && index === cursor}
                        bind:this={rowEls[index]}
                    >
                        <span class="todo-mark">☐</span>
                        <span class="todo-title">
                            {#if pendingDeleteId === task.id}
                                <span class="habits-delete-confirm"
                                    >x again to delete</span
                                >
                            {:else}
                                {task.title}
                            {/if}
                        </span>
                        {#each task.contexts as context}
                            <span class="gtd-context">@{context}</span>
                        {/each}
                        {#if projectById[task.projectId]}
                            <span class="gtd-project-tag"
                                >{projectById[task.projectId].title}</span
                            >
                        {/if}
                    </li>
                {/each}
            </ul>

            {#if completed.length > 0}
                <div class="todo-section-label">Completed</div>

                <ul class="todo-list">
                    {#each completed as task, i (task.id)}
                        {@const index = active.length + i}
                        <li
                            class="todo-item done"
                            class:cursor={!inputMode && index === cursor}
                            bind:this={rowEls[index]}
                        >
                            <span class="todo-mark todo-mark-done">☑</span>
                            <span class="todo-title">
                                {#if pendingDeleteId === task.id}
                                    <span class="habits-delete-confirm"
                                        >x again to delete</span
                                    >
                                {:else}
                                    {task.title}
                                {/if}
                            </span>
                            <span class="todo-completed-date"
                                >{formatCompletedDate(task.completedAt)}</span
                            >
                        </li>
                    {/each}
                </ul>
            {/if}
        {/if}

        {#if pickerOpen}
            <div class="gtd-picker">
                <div class="todo-section-label">
                    Project for “{rows[cursor]?.title}”
                </div>
                <ul class="todo-list">
                    {#each pickerOptions as project, index}
                        <li
                            class="todo-item"
                            class:cursor={index === pickerCursor}
                        >
                            {project ? project.title : "— no project —"}
                        </li>
                    {/each}
                </ul>
                {#if activeProjects.length === 0}
                    <p class="todo-more-hint">
                        No active projects — add some on the Projects tab
                    </p>
                {/if}
            </div>
        {/if}

        <div class="tasks-add-row">
            <input
                class="todo-input"
                type="text"
                bind:this={inputEl}
                bind:value={inputValue}
                placeholder={inputMode === "edit"
                    ? "edit task… (@tag adds a context)"
                    : "press a to capture to the inbox… (@tag adds a context)"}
                on:keydown={onInputKeydown}
                on:focus={() => (inputMode = inputMode || "add")}
                on:blur={closeInput}
            />
        </div>

        {#if error}
            <p class="todo-error">{error}</p>
        {/if}
    {/if}
</div>

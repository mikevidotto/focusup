<script>
    /*

Tasks page, modeled on GTD (Getting Things Done).

Everything is captured into the Inbox, then clarified: do it now (enter),
make it a Next Action (n), park it in Someday/Maybe (s), or attach it to a
project (p, which also makes it a next action). Next Actions can be
filtered by @context (f). Projects themselves live on the Projects tab.

Tasks finished today stay in their list (struck through, so a misclick is
one enter away from undone); older ones only show in the Done list.

The page has three views, all with the detail pane on the right:
- lists: the GTD lists below
- clarify: the inbox one item at a time (InboxProcess.svelte)
- review: the weekly review (WeeklyReview.svelte)

Keyboard (lists view):
- [ / ] switch list (Inbox / Next Actions / Someday / Done)
- j/k move, enter toggle done
- a capture to the inbox ("@tag" words become contexts)
- e edit title and contexts, p pick project, x x delete
- n / s / i move to Next / Someday / Inbox
- f cycle the @context filter (Next Actions only)
- c clarify the inbox, w weekly review

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        LISTS,
        splitTasks,
        parseCapture,
        formatForEdit,
        allContexts,
        countByList,
        isDoneToday,
        doneThisWeek,
        groupDoneByDay,
        isReviewDue,
    } from "../taskDisplay.js";
    import TaskDetail from "./TaskDetail.svelte";
    import InboxProcess from "./InboxProcess.svelte";
    import WeeklyReview from "./WeeklyReview.svelte";
    import {
        ListTasks,
        ListProjects,
        GetSettings,
        MarkWeeklyReviewed,
        AddTask,
        AddProjectTask,
        ToggleTask,
        ToggleProject,
        MoveTask,
        RenameTask,
        SetTaskContexts,
        SetTaskProject,
        DeleteTask,
    } from "../../../wailsjs/go/main/App.js";

    const MOVE_KEYS = { n: "next", s: "someday", i: "inbox" };

    const LIST_TIPS = {
        inbox: "Everything lands here first. Press c to clarify it one item at a time.",
        next: "Concrete actions you can do now. Press f to narrow by @context.",
        someday: "Ideas parked for later. The weekly review (w) revisits them.",
        done: "Tasks completed before today. Enter puts one back in its list.",
    };

    let tasks = [];
    let projects = [];
    let lastReviewAt = null;
    // "lists" | "clarify" | "review"
    let view = "lists";
    // Where clarify returns to: the lists or back into the review.
    let clarifyReturn = "lists";
    let reviewStep = 0;
    let reviewTask = null;
    let reviewProject = null;
    let childRef;

    let listId = "inbox";
    let contextFilter = null;
    let cursor = 0;
    // null | "add" | "edit" | "project-task"
    let inputMode = null;
    // The task being edited, or the project getting a next action.
    let inputTarget = null;
    let inputValue = "";
    let inputEl;
    let pendingDeleteId = null;
    // The task whose project is being picked, while the picker is open.
    let pickerTask = null;
    let pickerCursor = 0;
    let rowEls = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            let settings;
            [tasks, projects, settings] = await Promise.all([
                ListTasks(),
                ListProjects(),
                GetSettings(),
            ]);
            lastReviewAt = settings.lastReviewAt ?? null;

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
    $: weekDone = doneThisWeek(tasks);
    $: reviewDue = isReviewDue(lastReviewAt);
    $: projectById = Object.fromEntries(projects.map((p) => [p.id, p]));
    $: activeProjects = projects.filter((p) => !p.done);
    $: pickerOptions = [null, ...activeProjects];

    $: isNext = listId === "next";
    $: isDone = listId === "done";
    $: ({ active: open, completed: done } = splitTasks(tasks));
    $: inbox = open.filter((t) => t.list === "inbox");
    $: nextContexts = allContexts(open.filter((t) => t.list === "next"));
    $: if (contextFilter && !nextContexts.includes(contextFilter)) {
        contextFilter = null;
    }

    $: matchesFilter = (t) =>
        !isNext || !contextFilter || t.contexts.includes(contextFilter);
    $: active = isDone
        ? []
        : open.filter((t) => t.list === listId && matchesFilter(t));
    $: doneToday = isDone
        ? []
        : done.filter(
              (t) => t.list === listId && isDoneToday(t) && matchesFilter(t),
          );
    $: doneGroups = isDone ? groupDoneByDay(tasks) : [];
    $: rows = isDone
        ? doneGroups.flatMap((g) => g.tasks)
        : [...active, ...doneToday];
    $: if (cursor > rows.length - 1) {
        cursor = Math.max(0, rows.length - 1);
    }

    $: if (view === "lists") rowEls[cursor]?.scrollIntoView({ block: "nearest" });

    $: detailTask =
        view === "review" ? reviewTask : view === "lists" ? rows[cursor] : null;
    $: detailTip =
        view === "clarify"
            ? "Is it actionable? If it takes under two minutes, do it now. Otherwise decide where it lives — don't leave it undecided."
            : view === "review"
              ? reviewProject
                  ? `“${reviewProject.title}” has no next action. Press a to add one, or d if the project is finished.`
                  : "The weekly review keeps the system trustworthy: empty the inbox, unstick projects, prune Someday."
              : LIST_TIPS[listId];

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

    // ---- task actions (shared by every view) ----
    function toggleTask(task) {
        run(async () => replaceTask(await ToggleTask(task.id)));
    }

    function moveTask(task, list) {
        if (task.list !== list) {
            run(async () => replaceTask(await MoveTask(task.id, list)));
        }
    }

    function deleteTask(task) {
        run(async () => {
            await DeleteTask(task.id);
            tasks = tasks.filter((t) => t.id !== task.id);
        });
    }

    function toggleProject(project) {
        run(async () => {
            const updated = await ToggleProject(project.id);
            projects = projects.map((p) => (p.id === updated.id ? updated : p));
        });
    }

    // ---- views ----
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

    function startClarify(returnTo) {
        clarifyReturn = returnTo;
        view = "clarify";
    }

    function exitClarify() {
        view = clarifyReturn;
    }

    function startReview() {
        reviewStep = 0;
        view = "review";
    }

    function finishReview() {
        run(async () => {
            const settings = await MarkWeeklyReviewed();
            lastReviewAt = settings.lastReviewAt ?? null;
            view = "lists";
        });
    }

    // ---- project picker ----
    function openPicker(task) {
        pickerCursor = Math.max(
            0,
            pickerOptions.findIndex((p) => (p?.id ?? "") === task.projectId),
        );
        pickerTask = task;
    }

    function pickProject() {
        const task = pickerTask;
        const project = pickerOptions[pickerCursor];
        pickerTask = null;

        run(async () =>
            replaceTask(await SetTaskProject(task.id, project?.id ?? "")),
        );
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
                pickerTask = null;
                break;
        }
    }

    // ---- text input (capture, edit, project next action) ----
    async function openInput(kind, target = null) {
        inputMode = kind;
        inputTarget = target;
        inputValue = kind === "edit" ? formatForEdit(target) : "";
        await tick();
        inputEl?.focus();
    }

    function closeInput() {
        inputMode = null;
        inputTarget = null;
        inputValue = "";
        inputEl?.blur();
    }

    async function submitInput() {
        const { title, contexts } = parseCapture(inputValue);
        const kind = inputMode;
        const target = inputTarget;
        closeInput();

        if (kind === "edit") {
            await run(async () => {
                // A bare "@tag" edit retags the task and keeps its title.
                if (title && title !== target.title) {
                    replaceTask(await RenameTask(target.id, title));
                }
                replaceTask(await SetTaskContexts(target.id, contexts));
            });
        } else if (kind === "project-task" && title) {
            await run(async () => {
                tasks = [...tasks, await AddProjectTask(target.id, title, contexts)];
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

    $: inputPlaceholder =
        inputMode === "edit"
            ? "edit task… (@tag adds a context)"
            : inputMode === "project-task"
              ? `next action for “${inputTarget?.title}”… (@tag adds a context)`
              : "press a to capture to the inbox… (@tag adds a context)";

    // ---- keyboard ----
    function handleKey(event) {
        if (inputMode) {
            return;
        }

        if (pickerTask) {
            handlePickerKey(event);
            return;
        }

        if (view !== "lists") {
            childRef?.handleKey(event);
            return;
        }

        // "x" arms a delete; the very next key either confirms it (x) or
        // cancels it (anything else, which is then handled normally).
        if (pendingDeleteId) {
            const id = pendingDeleteId;
            pendingDeleteId = null;

            if (event.key === "x") {
                event.preventDefault();
                const task = tasks.find((t) => t.id === id);
                if (task) deleteTask(task);
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

            case "c":
                event.preventDefault();
                startClarify("lists");
                return;

            case "w":
                event.preventDefault();
                startReview();
                return;
        }

        const task = rows[cursor];
        if (!task) {
            return;
        }

        if (MOVE_KEYS[event.key]) {
            event.preventDefault();
            if (!task.done) moveTask(task, MOVE_KEYS[event.key]);
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
                toggleTask(task);
                break;

            case "e":
                event.preventDefault();
                openInput("edit", task);
                break;

            case "p":
                event.preventDefault();
                openPicker(task);
                break;

            case "x":
                event.preventDefault();
                pendingDeleteId = task.id;
                break;
        }
    }
</script>

<div class="page-placeholder tasks-page">
    <span class="eyebrow">FOCUSUP / TASKS</span>

    <div class="tasks-heading-row">
        <h1>Tasks</h1>
        <span class="todo-header-count">
            {counts.inbox} in inbox • {counts.next} next • {weekDone} done this week
            {#if reviewDue}
                • <span class="gtd-review-due">review due (w)</span>
            {/if}
        </span>
    </div>

    <div class="tasks-hint">
        {#if pickerTask}
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> pick
            <kbd>-</kbd> no project
            <kbd>q</kbd> cancel
        {:else if view === "lists"}
            <kbd>[</kbd><kbd>]</kbd> list
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> done
            <kbd>a</kbd> capture
            <kbd>c</kbd> clarify inbox
            <kbd>w</kbd> weekly review
            {#if isNext}<kbd>f</kbd> context{/if}
        {:else if view === "clarify"}
            Clarifying the inbox — one decision per item
        {:else}
            Weekly review
        {/if}
    </div>

    <div class="habits-layout">
        <div class="habits-main">
            {#if loading}
                <p>Loading tasks…</p>
            {:else if view === "clarify"}
                <InboxProcess
                    bind:this={childRef}
                    {inbox}
                    onMove={moveTask}
                    onToggle={toggleTask}
                    onPick={openPicker}
                    onEdit={(task) => openInput("edit", task)}
                    onDelete={deleteTask}
                    onExit={exitClarify}
                />
            {:else if view === "review"}
                <WeeklyReview
                    bind:this={childRef}
                    bind:step={reviewStep}
                    bind:selectedTask={reviewTask}
                    bind:selectedProject={reviewProject}
                    {tasks}
                    {projects}
                    onProcessInbox={() => startClarify("review")}
                    onAddProjectTask={(project) => openInput("project-task", project)}
                    onToggleProject={toggleProject}
                    onMove={moveTask}
                    onDelete={deleteTask}
                    onFinish={finishReview}
                    onExit={() => (view = "lists")}
                />
            {:else}
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

                {#if rows.length === 0}
                    <div class="empty-widget">
                        {#if listId === "inbox"}
                            <span>無</span>
                            <p>Inbox zero — press a to capture something</p>
                        {:else if listId === "someday"}
                            <span>空</span>
                            <p>Nothing parked — press s on a task to move it here</p>
                        {:else if listId === "done"}
                            <span>空</span>
                            <p>Nothing finished before today yet</p>
                        {:else}
                            <span>空</span>
                            <p>No next actions — press c to clarify your inbox</p>
                        {/if}
                    </div>
                {:else if isDone}
                    {#each doneGroups as group (group.label)}
                        <div class="todo-section-label gtd-done-day">
                            {group.label}
                            <span class="gtd-list-count">{group.tasks.length}</span>
                        </div>
                        <ul class="todo-list">
                            {#each group.tasks as task (task.id)}
                                {@const index = rows.indexOf(task)}
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
                                    {#if projectById[task.projectId]}
                                        <span class="gtd-project-tag"
                                            >{projectById[task.projectId].title}</span
                                        >
                                    {/if}
                                </li>
                            {/each}
                        </ul>
                    {/each}
                {:else}
                    <ul class="todo-list">
                        {#each rows as task, index (task.id)}
                            {#if index === active.length}
                                <li class="todo-section-label gtd-done-day">
                                    Done today
                                </li>
                            {/if}
                            <li
                                class="todo-item"
                                class:done={task.done}
                                class:cursor={!inputMode && index === cursor}
                                bind:this={rowEls[index]}
                            >
                                <span
                                    class="todo-mark"
                                    class:todo-mark-done={task.done}
                                    >{task.done ? "☑" : "☐"}</span
                                >
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
                {/if}
            {/if}

            <div class="tasks-add-row">
                <input
                    class="todo-input"
                    type="text"
                    bind:this={inputEl}
                    bind:value={inputValue}
                    placeholder={inputPlaceholder}
                    on:keydown={onInputKeydown}
                    on:focus={() => (inputMode = inputMode || "add")}
                    on:blur={closeInput}
                />
            </div>

            {#if error}
                <p class="todo-error">{error}</p>
            {/if}
        </div>

        {#if pickerTask}
            <div class="habit-history gtd-picker">
                <div class="habit-history-title">
                    <span class="eyebrow">PROJECT FOR</span>
                    <h2>{pickerTask.title}</h2>
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
        {:else}
            <TaskDetail
                task={detailTask}
                project={detailTask ? projectById[detailTask.projectId] : null}
                tip={detailTip}
                showKeys={view === "lists"}
            />
        {/if}
    </div>
</div>

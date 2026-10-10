<script>
    /*

GTD weekly review for the Tasks page, as three steps:
1. Inbox: get it to zero (c jumps into clarify mode)
2. Stalled projects: every active project needs a next action
   (a adds one, d marks the project done)
3. Someday/Maybe: promote (n), keep (j moves on) or drop (x x)

Keyboard (forwarded by TasksPage through handleKey): [ / ] change step,
j/k move, enter on the last step finishes the review, q / esc leave.
The parent owns the data and the API calls; `step` is bound so the review
picks up where it was after a detour through clarify mode.

*/
    import { formatAge, stalledProjects } from "../taskDisplay.js";

    export let tasks = [];
    export let projects = [];
    export let step = 0;
    // Bound by the parent to show the selection in the detail pane.
    export let selectedTask = null;
    export let selectedProject = null;

    export let onProcessInbox;
    export let onAddProjectTask;
    export let onToggleProject;
    export let onMove;
    export let onDelete;
    export let onFinish;
    export let onExit;

    const STEPS = ["Inbox", "Stalled projects", "Someday/Maybe"];

    let cursor = 0;
    let pendingDeleteId = null;

    $: inboxCount = tasks.filter((t) => !t.done && t.list === "inbox").length;
    $: stalled = stalledProjects(projects, tasks);
    $: someday = tasks.filter((t) => !t.done && t.list === "someday");
    $: items = step === 1 ? stalled : step === 2 ? someday : [];
    $: if (cursor > items.length - 1) {
        cursor = Math.max(0, items.length - 1);
    }
    $: selectedTask = step === 2 ? (someday[cursor] ?? null) : null;
    $: selectedProject = step === 1 ? (stalled[cursor] ?? null) : null;

    function changeStep(delta) {
        step = Math.min(Math.max(step + delta, 0), STEPS.length - 1);
        cursor = 0;
    }

    export function handleKey(event) {
        if (pendingDeleteId) {
            const id = pendingDeleteId;
            pendingDeleteId = null;

            if (event.key === "x") {
                event.preventDefault();
                const task = someday.find((t) => t.id === id);
                if (task) onDelete(task);
                return;
            }
        }

        switch (event.key) {
            case "q":
            case "Escape":
                event.preventDefault();
                onExit();
                return;

            case "[":
            case "]":
                event.preventDefault();
                changeStep(event.key === "[" ? -1 : 1);
                return;

            case "Enter":
                event.preventDefault();
                if (step === STEPS.length - 1) {
                    onFinish();
                } else {
                    changeStep(1);
                }
                return;

            case "j":
                event.preventDefault();
                cursor = Math.min(cursor + 1, items.length - 1);
                return;

            case "k":
                event.preventDefault();
                cursor = Math.max(cursor - 1, 0);
                return;
        }

        if (step === 0 && event.key === "c" && inboxCount > 0) {
            event.preventDefault();
            onProcessInbox();
        } else if (step === 1 && selectedProject) {
            if (event.key === "a") {
                event.preventDefault();
                onAddProjectTask(selectedProject);
            } else if (event.key === "d") {
                event.preventDefault();
                onToggleProject(selectedProject);
            }
        } else if (step === 2 && selectedTask) {
            if (event.key === "n") {
                event.preventDefault();
                onMove(selectedTask, "next");
            } else if (event.key === "x") {
                event.preventDefault();
                pendingDeleteId = selectedTask.id;
            }
        }
    }
</script>

<div class="gtd-review">
    <ol class="gtd-review-steps">
        {#each STEPS as label, index}
            <li class:active={index === step} class:past={index < step}>
                <span class="gtd-review-step-number">{index + 1}</span>
                {label}
            </li>
        {/each}
    </ol>

    {#if step === 0}
        <div class="gtd-review-body">
            {#if inboxCount === 0}
                <div class="empty-widget">
                    <span>無</span>
                    <p>Inbox zero — press enter or ] for the next step</p>
                </div>
            {:else}
                <p class="gtd-review-prompt">
                    {inboxCount} item{inboxCount === 1 ? "" : "s"} in the inbox.
                    Clarify each one so nothing waits undecided.
                </p>
                <div class="tasks-hint"><kbd>c</kbd> clarify the inbox</div>
            {/if}
        </div>
    {:else if step === 1}
        <div class="gtd-review-body">
            {#if stalled.length === 0}
                <div class="empty-widget">
                    <span>動</span>
                    <p>Every active project has a next action</p>
                </div>
            {:else}
                <p class="gtd-review-prompt">
                    These projects have no next action. What's the very next
                    physical step?
                </p>
                <ul class="todo-list">
                    {#each stalled as project, index (project.id)}
                        <li class="todo-item" class:cursor={index === cursor}>
                            <span class="todo-mark">▸</span>
                            <span class="todo-title">{project.title}</span>
                            <span class="todo-completed-date"
                                >{formatAge(project.createdAt)}</span
                            >
                        </li>
                    {/each}
                </ul>
                <div class="tasks-hint">
                    <kbd>a</kbd> add next action <kbd>d</kbd> project done
                </div>
            {/if}
        </div>
    {:else}
        <div class="gtd-review-body">
            {#if someday.length === 0}
                <div class="empty-widget">
                    <span>空</span>
                    <p>Nothing parked in Someday/Maybe</p>
                </div>
            {:else}
                <p class="gtd-review-prompt">
                    Anything here ready to start, or no longer wanted?
                </p>
                <ul class="todo-list">
                    {#each someday as task, index (task.id)}
                        <li class="todo-item" class:cursor={index === cursor}>
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
                            <span class="todo-completed-date"
                                >{formatAge(task.createdAt)}</span
                            >
                        </li>
                    {/each}
                </ul>
                <div class="tasks-hint">
                    <kbd>n</kbd> promote to next <kbd>j</kbd> keep
                    <kbd>x</kbd><kbd>x</kbd> drop
                </div>
            {/if}
        </div>
    {/if}

    <div class="tasks-hint gtd-review-nav">
        <kbd>[</kbd><kbd>]</kbd> step
        {#if step === STEPS.length - 1}
            <kbd>enter</kbd> finish review
        {:else}
            <kbd>enter</kbd> next step
        {/if}
        <kbd>q</kbd> leave
    </div>
</div>

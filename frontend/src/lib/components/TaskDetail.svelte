<script>
    // Right-hand pane of the Tasks page: everything about the task under the
    // cursor, plus the keys that apply to it. With no task it shows `tip`.
    import { LISTS, formatAge, formatCompletedDate } from "../taskDisplay.js";

    export let task = null;
    export let project = null;
    export let tip = "";
    // Off outside the lists view, where these keys mean something else.
    export let showKeys = true;

    $: listLabel = task
        ? task.done
            ? "Done"
            : (LISTS.find((l) => l.id === task.list)?.label ?? task.list)
        : "";
</script>

<div class="habit-history task-detail">
    {#if task}
        <div class="habit-history-title">
            <span class="eyebrow">{listLabel.toUpperCase()}</span>
            <h2 class="task-detail-title" class:done={task.done}>{task.title}</h2>
        </div>

        <dl class="task-detail-fields">
            <dt>Project</dt>
            <dd>
                {#if project}
                    <span class="gtd-project-tag">{project.title}</span>
                {:else}
                    <span class="task-detail-none">none</span>
                {/if}
            </dd>

            <dt>Contexts</dt>
            <dd class="task-detail-contexts">
                {#each task.contexts as context}
                    <span class="gtd-context">@{context}</span>
                {:else}
                    <span class="task-detail-none">none</span>
                {/each}
            </dd>

            <dt>Captured</dt>
            <dd>{formatAge(task.createdAt)}</dd>

            {#if task.done}
                <dt>Completed</dt>
                <dd>{formatCompletedDate(task.completedAt)}</dd>
            {/if}
        </dl>

        {#if showKeys}
            <div class="tasks-hint task-detail-keys">
                <kbd>enter</kbd>
                {task.done ? "not done" : "done"}
                <kbd>e</kbd> edit
                {#if !task.done}
                    {#if task.list !== "next"}<kbd>n</kbd> next{/if}
                    {#if task.list !== "someday"}<kbd>s</kbd> someday{/if}
                    {#if task.list !== "inbox"}<kbd>i</kbd> inbox{/if}
                {/if}
                <kbd>p</kbd> project
                <kbd>x</kbd><kbd>x</kbd> delete
            </div>
        {/if}
    {:else}
        <p class="task-detail-tip">{tip}</p>
    {/if}
</div>

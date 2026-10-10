<script>
    /*

Clarify mode for the Tasks page: the inbox one item at a time. Every
decision key acts on the current item, which then leaves the inbox, so the
next one slides into place. The parent owns the data and the API calls.

Keyboard (forwarded by TasksPage through handleKey):
- n next, s someday, p project (also makes it a next action)
- enter done (the two-minute rule), e edit, x x delete
- j skip, k back, q / esc leave

*/
    import { formatAge } from "../taskDisplay.js";

    export let inbox = [];
    export let onMove;
    export let onToggle;
    export let onPick;
    export let onEdit;
    export let onDelete;
    export let onExit;

    // Skipped items stay in the inbox before `position`, so it doubles as
    // the skip count; acted-on items leave, letting the next one move up.
    let position = 0;
    let pendingDelete = false;

    $: current = inbox[position] ?? null;

    export function handleKey(event) {
        if (pendingDelete) {
            pendingDelete = false;

            if (event.key === "x") {
                event.preventDefault();
                if (current) onDelete(current);
                return;
            }
        }

        switch (event.key) {
            case "q":
            case "Escape":
                event.preventDefault();
                onExit();
                return;
        }

        if (event.key === "k") {
            event.preventDefault();
            position = Math.max(0, Math.min(position, inbox.length) - 1);
            return;
        }

        if (!current) {
            return;
        }

        switch (event.key) {
            case "n":
                event.preventDefault();
                onMove(current, "next");
                break;

            case "s":
                event.preventDefault();
                onMove(current, "someday");
                break;

            case "p":
                event.preventDefault();
                onPick(current);
                break;

            case "Enter":
            case " ":
                event.preventDefault();
                onToggle(current);
                break;

            case "e":
                event.preventDefault();
                onEdit(current);
                break;

            case "x":
                event.preventDefault();
                pendingDelete = true;
                break;

            case "j":
                event.preventDefault();
                position += 1;
                break;

        }
    }
</script>

<div class="gtd-process">
    {#if current}
        <div class="gtd-process-progress">
            <span class="eyebrow">CLARIFY</span>
            <span>{position + 1} of {inbox.length}</span>
        </div>

        <div class="gtd-process-card">
            <h2 class="gtd-process-title">
                {#if pendingDelete}
                    <span class="habits-delete-confirm">x again to delete</span>
                {:else}
                    {current.title}
                {/if}
            </h2>

            <div class="gtd-process-meta">
                <span>captured {formatAge(current.createdAt)}</span>
                {#each current.contexts as context}
                    <span class="gtd-context">@{context}</span>
                {/each}
            </div>
        </div>

        <ul class="gtd-process-choices">
            <li><kbd>enter</kbd> under two minutes? do it now</li>
            <li><kbd>n</kbd> a single next action</li>
            <li><kbd>p</kbd> part of a project</li>
            <li><kbd>s</kbd> not now, someday/maybe</li>
            <li><kbd>x</kbd><kbd>x</kbd> not needed, delete</li>
            <li class="gtd-process-secondary">
                <kbd>e</kbd> edit <kbd>j</kbd> skip <kbd>k</kbd> back
                <kbd>q</kbd> stop
            </li>
        </ul>
    {:else}
        <div class="empty-widget">
            <span>無</span>
            {#if inbox.length === 0}
                <p>Inbox zero — press q to go back</p>
            {:else}
                <p>
                    {inbox.length} skipped and left in the inbox — press k to revisit
                    or q to go back
                </p>
            {/if}
        </div>
    {/if}
</div>

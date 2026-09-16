<script>
    import { onMount, onDestroy } from "svelte";

    import { EventsOn } from "../../../wailsjs/runtime/runtime.js";

    const REMINDER_DUE_EVENT = "calendar:reminder-due";
    const DISMISS_AFTER_MS = 8000;

    let popups = [];
    let nextId = 0;
    let unsubscribe;

    function formatWhen(iso) {
        return new Date(iso).toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
    }

    function dismiss(id) {
        popups = popups.filter(p => p.id !== id);
    }

    function show(reminder) {
        const id = nextId++;
        popups = [...popups, { id, reminder }];
        setTimeout(() => dismiss(id), DISMISS_AFTER_MS);
    }

    onMount(() => {
        unsubscribe = EventsOn(REMINDER_DUE_EVENT, show);
    });

    onDestroy(() => {
        unsubscribe?.();
    });
</script>

{#if popups.length > 0}
    <div class="reminder-popup-stack">
        {#each popups as popup (popup.id)}
            <div class="reminder-popup" role="alert">
                <div class="reminder-popup-body">
                    <span class="reminder-popup-time">{formatWhen(popup.reminder.occurrenceStart)}</span>
                    <span class="reminder-popup-title">{popup.reminder.eventTitle}</span>
                </div>

                <button
                    type="button"
                    class="reminder-popup-dismiss"
                    aria-label="Dismiss reminder"
                    on:click={() => dismiss(popup.id)}
                >
                    ×
                </button>
            </div>
        {/each}
    </div>
{/if}

<script>
    import { onMount } from "svelte";

    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import { formatOccurrenceWhen, occurrenceKey } from "../../calendarDisplay.js";
    import { ListCalendarOccurrences } from "../../../../wailsjs/go/main/App.js";

    export let focused = false;

    const MAX_VISIBLE = 5;
    const LOOKAHEAD_DAYS = 30;

    let occurrences = [];
    let cursor = 0;
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            const now = new Date();
            const rangeEnd = new Date(now.getTime() + LOOKAHEAD_DAYS * 24 * 60 * 60 * 1000);
            occurrences = await ListCalendarOccurrences(now, rangeEnd);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    });

    $: visible = occurrences.slice(0, MAX_VISIBLE);
    $: hiddenCount = occurrences.length - visible.length;

    function handleKey(event) {
        if (visible.length === 0) {
            return;
        }

        switch (event.key) {
            case "j":
                event.preventDefault();
                cursor = Math.min(cursor + 1, visible.length - 1);
                break;

            case "k":
                event.preventDefault();
                cursor = Math.max(cursor - 1, 0);
                break;
        }
    }

    $: if (focused) {
        activeWidgetKeyHandler.set(handleKey);
    } else {
        activeWidgetKeyHandler.set(null);
    }
</script>

<div class="calendar-widget">
    <header class="calendar-widget-header">
        <span class="calendar-widget-title">Calendar</span>
        <span class="calendar-widget-count">{occurrences.length} upcoming</span>
    </header>

    <div class="calendar-widget-body">
        {#if loading}
            <div class="empty-widget">
                <span>…</span>
                <p>Loading events</p>
            </div>
        {:else if occurrences.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No upcoming events in the next {LOOKAHEAD_DAYS} days</p>
            </div>
        {:else}
            <ul class="calendar-widget-list">
                {#each visible as occ, index (occurrenceKey(occ))}
                    <li class="calendar-widget-item" class:cursor={focused && index === cursor}>
                        <span class="calendar-widget-when">{formatOccurrenceWhen(occ)}</span>
                        <span class="calendar-widget-event-title">{occ.title}</span>
                    </li>
                {/each}

                {#if hiddenCount > 0}
                    <li class="calendar-widget-more-hint">+{hiddenCount} more on the Calendar page</li>
                {/if}
            </ul>
        {/if}

        {#if error}
            <p class="calendar-widget-error">{error}</p>
        {/if}
    </div>
</div>

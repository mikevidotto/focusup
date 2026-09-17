<script>
    import { onMount } from "svelte";

    import { startOfDay, endOfDay } from "../calendarGrid.js";
    import { occurrenceKey } from "../calendarDisplay.js";
    import { ListCalendarOccurrences, ToggleEventCompletion } from "../../../wailsjs/go/main/App.js";

    let occurrences = [];

    async function fetchDueSoon() {
        const now = new Date();
        const tomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);

        try {
            occurrences = await ListCalendarOccurrences(startOfDay(now), endOfDay(tomorrow));
        } catch (e) {
            // This banner is a nice-to-have prompt, not a critical surface —
            // a failed fetch just means it stays hidden rather than erroring
            // out loud on the dashboard.
            occurrences = [];
        }
    }

    async function toggleCompletion(occ) {
        await ToggleEventCompletion(occ.eventId, occ.originalStart);
        await fetchDueSoon();
    }

    onMount(fetchDueSoon);

    $: pending = occurrences.filter(o => !o.done);
</script>

{#if pending.length > 0}
    <div class="due-soon-banner" role="alert">
        <span class="due-soon-label">Due soon</span>

        <ul class="due-soon-list">
            {#each pending as occ (occurrenceKey(occ))}
                <li class="due-soon-item">
                    <button
                        type="button"
                        class="calendar-detail-check"
                        aria-label="Mark done"
                        on:click={() => toggleCompletion(occ)}
                    >
                        ☐
                    </button>
                    <span class="due-soon-title">{occ.title}</span>
                </li>
            {/each}
        </ul>
    </div>
{/if}

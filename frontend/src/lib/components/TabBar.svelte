<script>
    import { onMount, onDestroy } from "svelte";

    import { quoteForToday } from "../quotes.js";

    export let tabs = [];
    export let activeTab;
    export let onSelect;

    const quote = quoteForToday();

    let now = new Date();
    let interval;

    onMount(() => {
        interval = setInterval(() => {
            now = new Date();
        }, 15000);
    });

    onDestroy(() => {
        clearInterval(interval);
    });

    $: dateStr = now.toLocaleDateString("en-US", {
        weekday: "short",
        month: "short",
        day: "numeric",
        year: "numeric"
    });

    $: timeStr = now.toLocaleTimeString("en-US", {
        hour: "numeric",
        minute: "2-digit"
    });
</script>

<nav class="tab-bar">
    {#each tabs as tab}
        <button
            class:active={activeTab === tab.id}
            on:click={() => onSelect(tab.id)}
        >
            <span class="tab-key">{tab.key}</span>
            <span>{tab.label}</span>
        </button>
    {/each}

    <div class="tab-bar-meta">
        <div class="tab-bar-clock">
            <span class="tab-bar-date">{dateStr}</span>
            <span class="tab-bar-time">{timeStr}</span>
        </div>

        <div class="tab-bar-divider"></div>

        <span class="tab-bar-quote">{quote}</span>
    </div>
</nav>

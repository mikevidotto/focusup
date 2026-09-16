<script>
    export let tabs = [];
    export let activeTab;
    export let onSelect;
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
</nav>

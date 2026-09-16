<script>
    export let id;
    export let title;
    export let shortcut;
    export let selected = false;
    export let showHeader = true;
</script>

<section
    class="widget-slot"
    class:selected
    data-widget={id}
>
    {#if showHeader}
        <header class="widget-header">
            <div class="widget-title">
                {#if shortcut}
                    <span class="widget-shortcut">{shortcut}</span>
                {/if}

                <span>{title}</span>
            </div>

            <span class="widget-open">↗</span>
        </header>
    {/if}

    <div class="widget-content">
        <slot />
    </div>
</section>

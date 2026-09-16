<script>
    import WidgetSlot from "./WidgetSlot.svelte";
    import { widgets } from "../widgets.js";
    import { mode, selectedWidgetId } from "../stores/keyboard.js";
</script>

<div class="dashboard">
    <div class="dashboard-heading">
        <div>
            <span class="eyebrow">FOCUSUP / DASHBOARD</span>
            <h1>Good morning.</h1>
        </div>

        <div class="dashboard-hint">
            <kbd>h</kbd><kbd>j</kbd><kbd>k</kbd><kbd>l</kbd>
            select

            <kbd>enter</kbd>
            open

            <kbd>q</kbd>
            back
        </div>
    </div>

    <div class="widget-grid">
        {#each widgets as widget (widget.id)}
            <div class={`widget-wrapper ${widget.size}`}>
                <WidgetSlot
                    id={widget.id}
                    title={widget.title}
                    shortcut={widget.shortcut}
                    selected={$selectedWidgetId === widget.id}
                    showHeader={!widget.customHeader}
                >
                    {#if widget.component}
                        <svelte:component
                            this={widget.component}
                            focused={$mode === "widget" && $selectedWidgetId === widget.id}
                        />
                    {:else}
                        <div class="empty-widget">
                            <span>空</span>
                            <p>Widget content</p>
                        </div>
                    {/if}
                </WidgetSlot>
            </div>
        {/each}
    </div>
</div>

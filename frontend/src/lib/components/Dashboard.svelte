<script>
    import WidgetSlot from "./WidgetSlot.svelte";

    const placeholders = [
        {
            id: "widget-1",
            title: "Widget Slot",
            shortcut: "1",
            size: "large"
        },
        {
            id: "widget-2",
            title: "Widget Slot",
            shortcut: "2",
            size: "medium"
        },
        {
            id: "widget-3",
            title: "Widget Slot",
            shortcut: "3",
            size: "medium"
        }
    ];
</script>

<div class="dashboard">
    <div class="dashboard-heading">
        <div>
            <span class="eyebrow">FOCUSUP / DASHBOARD</span>
            <h1>Good morning.</h1>
        </div>

        <div class="dashboard-hint">
            <kbd>j</kbd><kbd>k</kbd>
            select

            <kbd>enter</kbd>
            open
        </div>
    </div>

    <div class="widget-grid">
        {#each placeholders as widget}
            <div class={`widget-wrapper ${widget.size}`}>
                <WidgetSlot
                    id={widget.id}
                    title={widget.title}
                    shortcut={widget.shortcut}
                >
                    <div class="empty-widget">
                        <span>空</span>
                        <p>Widget content</p>
                    </div>
                </WidgetSlot>
            </div>
        {/each}
    </div>
</div>

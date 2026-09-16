<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";

    import MenuBar from "./lib/components/MenuBar.svelte";
    import TabBar from "./lib/components/TabBar.svelte";
    import Dashboard from "./lib/components/Dashboard.svelte";

    import { tabs } from "./lib/navigation.js";
    import { widgets } from "./lib/widgets.js";
    import { firstWidget, moveSelection } from "./lib/keyboardGrid.js";
    import {
        mode,
        selectedWidgetId,
        activeWidgetKeyHandler
    } from "./lib/stores/keyboard.js";

    const DIRECTIONS = { h: "left", l: "right", j: "down", k: "up" };

    let activeTab = "dashboard";
    let version = "0.1.0";

    function selectTab(id) {
        activeTab = id;
        mode.set("tabs");
        selectedWidgetId.set(null);
    }

    function moveTab(direction) {
        const index = tabs.findIndex(tab => tab.id === activeTab);

        let next = index + direction;

        if (next < 0) {
            next = tabs.length - 1;
        }

        if (next >= tabs.length) {
            next = 0;
        }

        activeTab = tabs[next].id;
    }

    function handleDashboardKey(event) {
        const direction = DIRECTIONS[event.key];

        if (direction) {
            event.preventDefault();

            const nextId = moveSelection(widgets, get(selectedWidgetId), direction);

            if (nextId === null) {
                mode.set("tabs");
                selectedWidgetId.set(null);
            } else {
                selectedWidgetId.set(nextId);
            }

            return;
        }

        if (event.key === "Enter") {
            event.preventDefault();
            mode.set("widget");
        }
    }

    function handleTabsKey(event) {
        const numericTab = tabs.find(tab => tab.key === event.key);

        if (numericTab) {
            selectTab(numericTab.id);
            return;
        }

        switch (event.key) {
            case "h":
                moveTab(-1);
                break;

            case "l":
                moveTab(1);
                break;

            case "j": {
                if (activeTab !== "dashboard") {
                    break;
                }

                const first = firstWidget(widgets);

                if (first) {
                    mode.set("dashboard");
                    selectedWidgetId.set(first.id);
                }

                break;
            }

            case "?":
                console.log("Open keyboard help");
                break;

            case "/":
                event.preventDefault();
                console.log("Open command palette");
                break;
        }
    }

    function handleKeyboard(event) {
        const target = event.target;

        if (
            target instanceof HTMLInputElement ||
            target instanceof HTMLTextAreaElement ||
            target instanceof HTMLSelectElement
        ) {
            return;
        }

        const currentMode = get(mode);

        if (currentMode === "widget") {
            if (event.key === "q") {
                event.preventDefault();
                mode.set("dashboard");
                return;
            }

            get(activeWidgetKeyHandler)?.(event);
            return;
        }

        if (currentMode === "dashboard") {
            handleDashboardKey(event);
            return;
        }

        handleTabsKey(event);
    }

    onMount(() => {
        window.addEventListener("keydown", handleKeyboard);

        return () => {
            window.removeEventListener("keydown", handleKeyboard);
        };
    });
</script>

<div class="app-shell">
    <MenuBar {version} />

    <TabBar
        {tabs}
        {activeTab}
        onSelect={selectTab}
    />

    <main>
        {#if activeTab === "dashboard"}
            <Dashboard />
        {:else}
            <div class="page-placeholder">
                <span class="eyebrow">
                    FOCUSUP / {activeTab.toUpperCase()}
                </span>

                <h1>
                    {tabs.find(tab => tab.id === activeTab)?.label}
                </h1>

                <p>This section is ready to be built.</p>
            </div>
        {/if}
    </main>

    <footer>
        <div>
            <span>FOCUSUP</span>
            <span>v{version}</span>
        </div>

        <div class="footer-keys">
            <span><kbd>h</kbd>/<kbd>l</kbd> tabs</span>
            <span><kbd>j</kbd>/<kbd>k</kbd> navigate</span>
            <span><kbd>enter</kbd> open</span>
            <span><kbd>q</kbd> back</span>
            <span><kbd>/</kbd> command</span>
        </div>

        <span class="footer-message">より良い自分へ</span>
    </footer>
</div>

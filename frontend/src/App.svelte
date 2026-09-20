<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";

    import MenuBar from "./lib/components/MenuBar.svelte";
    import TabBar from "./lib/components/TabBar.svelte";
    import Dashboard from "./lib/components/Dashboard.svelte";
    import TasksPage from "./lib/components/TasksPage.svelte";
    import CalendarPage from "./lib/components/CalendarPage.svelte";
    import NotesPage from "./lib/components/NotesPage.svelte";
    import ReminderPopup from "./lib/components/ReminderPopup.svelte";

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

    // The single place responsible for changing which tab is active. Always
    // resets the dashboard's grid/widget-focus state; only releases the
    // shared activeWidgetKeyHandler when the tab is actually changing, so a
    // same-tab digit press can't wipe out a page's just-registered handler
    // (e.g. pressing "2" again while already on the Tasks page).
    function goToTab(id) {
        const changingTab = id !== activeTab;

        activeTab = id;
        mode.set("tabs");
        selectedWidgetId.set(null);

        if (changingTab) {
            activeWidgetKeyHandler.set(null);
        }
    }

    function selectTab(id) {
        goToTab(id);
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

        goToTab(tabs[next].id);
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

    // Reached only when currentMode === "tabs" (i.e. not locked into a
    // grid or a dashboard widget). Handles the always-available tab-bar
    // keys; anything else (j/k/enter/a/x, etc.) falls through to whichever
    // page has registered activeWidgetKeyHandler (e.g. the Tasks page), or
    // is a no-op if nothing has.
    //
    // h/l always cycle tabs here. Pages with their own 2D grid (currently
    // just Calendar) opt into the same tabs<->grid scheme the dashboard
    // uses: "j" hands off to mode "grid" (their own handler then owns
    // h/j/k/l — see handleKeyboard below), and until that happens, "k" is a
    // no-op rather than leaking into the page's handler (the dashboard
    // doesn't have this leak since nothing is registered pre-entry; a full
    // page's handler is always registered, so it needs an explicit gate).
    function handleTabsAndPageKey(event) {
        switch (event.key) {
            case "h":
                moveTab(-1);
                return;

            case "l":
                moveTab(1);
                return;

            case "j": {
                if (activeTab === "dashboard") {
                    const first = firstWidget(widgets);

                    if (first) {
                        mode.set("grid");
                        selectedWidgetId.set(first.id);
                    }

                    return;
                }

                if (activeTab === "calendar") {
                    mode.set("grid");
                    return;
                }

                break;
            }

            case "k":
                if (activeTab === "calendar") {
                    return;
                }
                break;

            case "?":
                console.log("Open keyboard help");
                return;

            case "/":
                event.preventDefault();
                console.log("Open command palette");
                return;
        }

        get(activeWidgetKeyHandler)?.(event);
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

        // Digit tab-shortcuts always work, regardless of mode (dashboard
        // grid nav, a locked widget, or a page like Tasks owning the keys).
        const numericTab = tabs.find(tab => tab.key === event.key);

        if (numericTab) {
            selectTab(numericTab.id);
            return;
        }

        const currentMode = get(mode);

        if (currentMode === "widget") {
            if (event.key === "q") {
                event.preventDefault();
                mode.set("grid");
                return;
            }

            get(activeWidgetKeyHandler)?.(event);
            return;
        }

        if (currentMode === "grid") {
            if (activeTab === "dashboard") {
                handleDashboardKey(event);
            } else {
                get(activeWidgetKeyHandler)?.(event);
            }
            return;
        }

        handleTabsAndPageKey(event);
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
        {:else if activeTab === "tasks"}
            <TasksPage />
        {:else if activeTab === "calendar"}
            <CalendarPage />
        {:else if activeTab === "notes"}
            <NotesPage />
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

    <ReminderPopup />
</div>

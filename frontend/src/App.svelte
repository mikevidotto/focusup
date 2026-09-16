<script>
    import { onMount } from "svelte";

    import MenuBar from "./lib/components/MenuBar.svelte";
    import TabBar from "./lib/components/TabBar.svelte";
    import Dashboard from "./lib/components/Dashboard.svelte";

    import { tabs } from "./lib/navigation.js";

    let activeTab = "dashboard";
    let version = "0.1.0";

    function selectTab(id) {
        activeTab = id;
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

    function handleKeyboard(event) {
        const target = event.target;

        if (
            target instanceof HTMLInputElement ||
            target instanceof HTMLTextAreaElement ||
            target instanceof HTMLSelectElement
        ) {
            return;
        }

        const numericTab = tabs.find(tab => tab.key === event.key);

        if (numericTab) {
            activeTab = numericTab.id;
            return;
        }

        switch (event.key) {
            case "h":
                moveTab(-1);
                break;

            case "l":
                moveTab(1);
                break;

            case "?":
                console.log("Open keyboard help");
                break;

            case "/":
                event.preventDefault();
                console.log("Open command palette");
                break;
        }
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
            <span><kbd>/</kbd> command</span>
        </div>

        <span class="footer-message">より良い自分へ</span>
    </footer>
</div>

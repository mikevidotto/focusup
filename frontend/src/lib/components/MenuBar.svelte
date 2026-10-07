<script>
    import {
        WindowMinimise,
        WindowToggleMaximise,
        Quit
    } from "../../../wailsjs/runtime/runtime.js";
    import { theme, toggleTheme } from "../stores/theme.js";

    export let version = "0.1.0";
</script>

<header class="menu-bar">
    <div class="brand">
        <span class="brand-mark">集中</span>

        <div>
            <strong>FocusUp</strong>
            <span class="version">v{version}</span>
        </div>
    </div>

    <div class="window-controls">
        <button
            aria-label="Switch to {$theme === 'dark' ? 'light' : 'dark'} theme"
            title="Toggle theme (t)"
            on:click={toggleTheme}
        >
            {$theme === "dark" ? "☀" : "☾"}
        </button>

        <button
            aria-label="Minimize"
            on:click={WindowMinimise}
        >
            ─
        </button>

        <button
            aria-label="Maximize"
            on:click={WindowToggleMaximise}
        >
            □
        </button>

        <button
            class="close"
            aria-label="Close"
            on:click={Quit}
        >
            ×
        </button>
    </div>
</header>

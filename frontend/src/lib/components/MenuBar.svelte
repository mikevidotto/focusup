<script>
    import {
        WindowMinimise,
        WindowToggleMaximise,
        Quit
    } from "../../../wailsjs/runtime/runtime.js";

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

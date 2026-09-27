<script>
    // Import the runtime from the auto-generated wailsjs directory
    import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
    import { onMount, onDestroy } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard";

    let cursor = 0;

    function openCtf101() {
        BrowserOpenURL("https://ctf101.org");
    }

    function openLiveOverFlow() {
        BrowserOpenURL("https://liveoverflow.com");
    }

    function openHelloInterview() {
        BrowserOpenURL("https://hellointerview.com");
    }

    async function handleKey(event) {
        switch (event.Key) {
        }
    }

    onMount(async () => {
        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
    });
</script>

<div class="page-placeholder notes-page" style="margin-top:0">
    <div class="page-placeholder">
        <span class="eyebrow"> FOCUSUP / NOTES </span>

        <h1>Notes</h1>
        <br />
        <br />
        <br />
        <br />

        <p>useful links</p>
        <ul>
            <li>
                <button on:click={openCtf101}>Ctf101</button>
            </li>
            <li>
                <button on:click={openLiveOverFlow}>LiveOverFlow</button>
            </li>
            <li>
                <button on:click={openHelloInterview}>hellointerview</button>
            </li>
        </ul>
    </div>
</div>

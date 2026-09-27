<script>
    // Import the runtime from the auto-generated wailsjs directory
    import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";

    let toggle = false;

    function openCtf101() {
        BrowserOpenURL("https://ctf101.org");
        toggle = !toggle;
    }

    function openLiveOverFlow() {
        BrowserOpenURL("https://liveoverflow.com");
        toggle = !toggle;
    }
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

        <button on:click={openCtf101}>Ctf101</button>
        <button on:click={openLiveOverFlow}>LiveOverFlow</button>

        <a>{toggle}</a>
    </div>
</div>

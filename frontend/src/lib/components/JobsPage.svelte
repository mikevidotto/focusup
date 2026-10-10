<script>
    /*

Jobs page: tracks the ai-job-search workspace's progress.

  To apply                          │ Applied & tracked
  > high  IBM (Confluent)  Cloud…   │   applied  Fellow  New Grad   Oct 8
    med   Harbor           Junior…  │   drafted  Solace  Junior…    Oct 7
    ─ CSIS  IT Analyst  applied Oct 7

The queue (job_scraper/apply_queue.md) and tracker (job_search_tracker.csv)
are read straight from the repo, which stays the source of truth. Scraping
and drafting happen outside FocusUp; this page only records when a job is
applied to or skipped.

Keyboard (list mode, so h/l still switch tabs):
- j/k move, w switch between the two lists, r reload from disk
- to apply: s skip with a reason
- applied & tracked: m m mark applied today

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler } from "../stores/keyboard.js";
    import {
        ListJobQueue,
        ListJobApplications,
        SetJobQueueOutcome,
        MarkJobApplied,
        GetJobsInfo,
    } from "../../../wailsjs/go/main/App.js";
    import {
        appliedDate,
        queueOutcomeDate,
        isThisWeek,
        formatShortDate,
    } from "../jobDisplay.js";

    let queue = [];
    let apps = [];
    let dir = "";
    let pane = "queue";
    let queueCursor = 0;
    let appCursor = 0;
    let loading = true;
    let error = null;
    let flash = "";
    let flashTimer;

    let skipping = false;
    let skipReason = "";
    let skipEl;
    let pendingApplied = false;

    $: openQueue = queue.filter((item) => !item.done);
    $: doneQueue = queue.filter((item) => item.done);
    $: orderedQueue = [...openQueue, ...doneQueue];
    $: orderedApps = [...apps].sort((a, b) => b.row - a.row);
    $: appliedCount = apps.filter((app) => app.status === "applied").length;
    $: appliedThisWeek = apps.filter((app) =>
        isThisWeek(appliedDate(app)),
    ).length;

    $: selectedQueue = orderedQueue[queueCursor];
    $: selectedApp = orderedApps[appCursor];

    async function reload() {
        try {
            [queue, apps] = await Promise.all([
                ListJobQueue(),
                ListJobApplications(),
            ]);
            queueCursor = clamp(queueCursor, queue.length);
            appCursor = clamp(appCursor, apps.length);
            error = null;
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    }

    onMount(async () => {
        try {
            dir = (await GetJobsInfo()).dir;
        } catch (e) {
            error = String(e);
        }

        await reload();
        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
        clearTimeout(flashTimer);
    });

    function showFlash(message) {
        flash = message;
        clearTimeout(flashTimer);
        flashTimer = setTimeout(() => (flash = ""), 2500);
    }

    async function attempt(action, message) {
        try {
            await action();
            if (message) {
                showFlash(message);
            }
        } catch (e) {
            error = String(e);
        }
    }

    function markApplied(app) {
        if (!app || app.status === "applied") {
            return;
        }
        attempt(async () => {
            await MarkJobApplied(app.row);
            await reload();
        }, `${app.company} marked applied`);
    }

    async function startSkip() {
        if (!selectedQueue || selectedQueue.done) {
            return;
        }
        skipping = true;
        skipReason = "";
        await tick();
        skipEl?.focus();
    }

    function closeSkip() {
        skipping = false;
        skipReason = "";
        skipEl?.blur();
    }

    async function submitSkip() {
        const reason = skipReason.trim();
        const item = selectedQueue;
        closeSkip();

        if (!reason || !item) {
            return;
        }

        await attempt(async () => {
            await SetJobQueueOutcome(item.number, `skipped - ${reason}`);
            await reload();
        }, `${item.company} skipped`);
    }

    function onSkipKeydown(event) {
        if (event.key === "Enter") {
            event.preventDefault();
            submitSkip();
        } else if (event.key === "Escape") {
            event.preventDefault();
            closeSkip();
        }
    }

    function moveCursor(delta) {
        if (pane === "queue") {
            queueCursor = clamp(queueCursor + delta, orderedQueue.length);
        } else {
            appCursor = clamp(appCursor + delta, orderedApps.length);
        }
    }

    function clamp(value, length) {
        return Math.max(0, Math.min(value, length - 1));
    }

    // What the date column shows for a tracked application.
    function appDateLabel(app) {
        const applied = appliedDate(app);
        if (applied) {
            return `applied ${formatShortDate(applied)}`;
        }
        return app.date ? `${app.status} ${formatShortDate(app.date)}` : "";
    }

    function handleKey(event) {
        const key = event.key;

        // "m" arms mark-applied; the very next key either confirms it (m)
        // or cancels it (anything else, which is then handled normally).
        if (pendingApplied) {
            pendingApplied = false;

            if (key === "m") {
                event.preventDefault();
                markApplied(selectedApp);
                return;
            }
        }

        switch (key) {
            case "j":
                event.preventDefault();
                moveCursor(1);
                return;
            case "k":
                event.preventDefault();
                moveCursor(-1);
                return;
            case "w":
                event.preventDefault();
                pane = pane === "queue" ? "apps" : "queue";
                return;
            case "r":
                event.preventDefault();
                attempt(reload, "Reloaded");
                return;
        }

        if (pane === "queue" && key === "s") {
            event.preventDefault();
            startSkip();
            return;
        }

        if (
            pane === "apps" &&
            key === "m" &&
            selectedApp &&
            selectedApp.status !== "applied"
        ) {
            event.preventDefault();
            pendingApplied = true;
        }
    }
</script>

<div class="page-placeholder jobs-page">
    <span class="eyebrow">FOCUSUP / JOBS</span>

    <div class="tasks-heading-row">
        <h1>Jobs</h1>
        <span class="todo-header-count">
            {openQueue.length} to apply • {appliedThisWeek} applied this week •
            {appliedCount} applied total
        </span>
    </div>

    <div class="tasks-hint">
        <kbd>j</kbd><kbd>k</kbd> move
        <kbd>w</kbd> switch list
        {#if pane === "queue"}
            <kbd>s</kbd> skip
        {:else}
            <kbd>m</kbd> mark applied
        {/if}
        <kbd>r</kbd> reload
    </div>

    {#if loading}
        <p>Loading job search…</p>
    {:else}
        <div class="jobs-layout">
            <section class="jobs-section" class:active={pane === "queue"}>
                <div class="todo-section-label">
                    To apply <span class="jobs-count">{openQueue.length}</span>
                </div>

                {#if orderedQueue.length === 0}
                    <div class="empty-widget">
                        <span>空</span>
                        <p>No apply queue in {dir}</p>
                    </div>
                {:else}
                    <ul class="jobs-list">
                        {#each orderedQueue as item, index (item.number)}
                            {@const selected =
                                pane === "queue" && index === queueCursor}
                            <li
                                class="jobs-row"
                                class:cursor={selected}
                                class:done={item.done}
                                on:click={() => {
                                    pane = "queue";
                                    queueCursor = index;
                                }}
                            >
                                <span class="jobs-fit"
                                    >{item.fit.toLowerCase()}</span
                                >
                                <span class="jobs-company">{item.company}</span>
                                <span class="jobs-role" title={item.notes}
                                    >{item.role}</span
                                >
                                <span class="jobs-meta" title={item.outcome}>
                                    {#if item.done}
                                        {item.outcome.startsWith("applied")
                                            ? `applied ${formatShortDate(queueOutcomeDate(item))}`
                                            : item.outcome}
                                    {:else if selected}
                                        <button
                                            type="button"
                                            class="jobs-action"
                                            on:click|stopPropagation={startSkip}
                                            >skip</button
                                        >
                                    {/if}
                                </span>
                            </li>
                        {/each}
                    </ul>
                {/if}

                {#if skipping}
                    <div class="tasks-add-row">
                        <input
                            class="todo-input"
                            type="text"
                            bind:this={skipEl}
                            bind:value={skipReason}
                            placeholder={`reason for skipping ${selectedQueue?.company ?? ""}…`}
                            on:keydown={onSkipKeydown}
                            on:blur={closeSkip}
                        />
                    </div>
                {/if}
            </section>

            <section class="jobs-section" class:active={pane === "apps"}>
                <div class="todo-section-label">
                    Applied & tracked
                    <span class="jobs-count">{apps.length}</span>
                </div>

                {#if orderedApps.length === 0}
                    <div class="empty-widget">
                        <span>空</span>
                        <p>No applications tracked yet</p>
                    </div>
                {:else}
                    <ul class="jobs-list">
                        {#each orderedApps as app, index (app.row)}
                            {@const selected =
                                pane === "apps" && index === appCursor}
                            <li
                                class="jobs-row"
                                class:cursor={selected}
                                on:click={() => {
                                    pane = "apps";
                                    appCursor = index;
                                }}
                            >
                                <span class={`jobs-status ${app.status}`}
                                    >{app.status}</span
                                >
                                <span class="jobs-company">{app.company}</span>
                                <span class="jobs-role" title={app.notes}
                                    >{app.role}</span
                                >
                                <span class="jobs-meta">
                                    {#if pendingApplied && selected}
                                        m again to mark applied
                                    {:else if selected && app.status !== "applied"}
                                        <button
                                            type="button"
                                            class="jobs-action"
                                            on:click|stopPropagation={() =>
                                                markApplied(app)}
                                            >mark applied</button
                                        >
                                    {:else}
                                        <span class="jobs-date"
                                            >{appDateLabel(app)}</span
                                        >
                                        {#if app.deadline && app.status !== "applied"}
                                            · due {formatShortDate(
                                                app.deadline,
                                            )}
                                        {/if}
                                        {#if app.fitRating}
                                            · fit {app.fitRating}
                                        {/if}
                                    {/if}
                                </span>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </section>
        </div>
    {/if}

    {#if flash}
        <p class="jobs-flash">{flash}</p>
    {/if}

    {#if error}
        <p class="todo-error">{error}</p>
    {/if}
</div>

<script>
    // Dashboard summary of the job search: headline counts plus a chart of
    // applications sent, paged between per day (last 14 days) and per week
    // (last 8 weeks).
    //
    // Keyboard while focused: h/l switch page.
    import { onMount } from "svelte";
    import { activeWidgetKeyHandler } from "../../stores/keyboard.js";
    import {
        ListJobQueue,
        ListJobApplications,
    } from "../../../../wailsjs/go/main/App.js";
    import {
        appliedDate,
        isThisWeek,
        dailyApplied,
        weeklyApplied,
    } from "../../jobDisplay.js";
    import MiniColumnChart from "./MiniColumnChart.svelte";

    export let focused = false;

    const PAGES = ["daily", "weekly"];

    const today = new Date();

    let page = "daily";

    let queue = [];
    let apps = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            [queue, apps] = await Promise.all([
                ListJobQueue(),
                ListJobApplications(),
            ]);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    });

    $: toApply = queue.filter((item) => !item.done).length;
    $: applied = apps.filter((app) => app.status === "applied").length;
    $: drafted = apps.filter((app) => app.status === "drafted").length;
    $: appliedThisWeek = apps.filter((app) =>
        isThisWeek(appliedDate(app)),
    ).length;
    $: days = dailyApplied(apps, today, 14);
    $: weeks = weeklyApplied(apps, today, 8);

    $: if (focused) {
        activeWidgetKeyHandler.set(handleKey);
    } else {
        activeWidgetKeyHandler.set(null);
    }

    function cyclePage(delta) {
        const next = PAGES.indexOf(page) + delta;
        page = PAGES[(next + PAGES.length) % PAGES.length];
    }

    function handleKey(event) {
        if (event.key === "h" || event.key === "l") {
            event.preventDefault();
            cyclePage(event.key === "h" ? -1 : 1);
        }
    }
</script>

<div class="todo-widget" class:focused>
    <header class="todo-header">
        <span class="todo-header-title">Jobs</span>
        <span class="todo-header-count">{toApply} to apply</span>
    </header>

    <div class="dash-body">
        {#if loading}
            <p>Loading…</p>
        {:else if error}
            <p class="todo-error">{error}</p>
        {:else}
            <div class="dash-stats">
                <div class="dash-stat">
                    <span class="dash-stat-label">Applied</span>
                    <span class="dash-stat-value">{applied}</span>
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">This week</span>
                    <span class="dash-stat-value">{appliedThisWeek}</span>
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">Drafted</span>
                    <span class="dash-stat-value">{drafted}</span>
                </div>
            </div>

            <div class="dash-subtitle jobs-widget-subtitle">
                <span>Applications sent per {page === "daily" ? "day" : "week"}</span>
                <span class="jobs-widget-pages" title="h / l to switch">
                    {#each PAGES as p}
                        <span class="jobs-widget-dot" class:active={p === page}
                        ></span>
                    {/each}
                </span>
            </div>
            {#if page === "daily"}
                <MiniColumnChart
                    buckets={days}
                    yMax={2}
                    ariaLabel="Applications sent per day, last 14 days"
                />
            {:else}
                <MiniColumnChart
                    buckets={weeks}
                    yMax={4}
                    ariaLabel="Applications sent per week, last 8 weeks"
                />
            {/if}
        {/if}
    </div>
</div>

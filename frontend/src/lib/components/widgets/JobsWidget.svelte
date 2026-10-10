<script>
    // Dashboard summary of the job search: headline counts plus
    // applications sent per week over the last 8 weeks.
    import { onMount } from "svelte";
    import {
        ListJobQueue,
        ListJobApplications,
    } from "../../../../wailsjs/go/main/App.js";
    import { appliedDate, isThisWeek, weeklyApplied } from "../../jobDisplay.js";
    import MiniColumnChart from "./MiniColumnChart.svelte";

    export let focused = false;

    const today = new Date();

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
    $: weeks = weeklyApplied(apps, today, 8);
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

            <div class="dash-subtitle">Applications sent per week</div>
            <MiniColumnChart
                buckets={weeks}
                yMax={4}
                ariaLabel="Applications sent per week, last 8 weeks"
            />
        {/if}
    </div>
</div>

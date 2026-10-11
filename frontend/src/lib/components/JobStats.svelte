<script>
    // Stats panel for the Jobs page: headline numbers for the search plus a
    // chart of applications sent per day (last 30 days), week (last 12
    // weeks) or month (last 12 months).
    import {
        JOB_STATS_VIEWS,
        appliedInLastDays,
        applyStreak,
        dailyApplied,
        jobFunnel,
        monthlyApplied,
        weeklyApplied,
    } from "../jobDisplay.js";
    import MiniColumnChart from "./widgets/MiniColumnChart.svelte";

    export let apps = [];
    export let queue = [];
    export let today;
    export let view = "day";
    export let onSelectView = () => {};

    const VIEW_LABELS = { day: "daily", week: "weekly", month: "monthly" };

    $: buckets =
        view === "month"
            ? monthlyApplied(apps, today, 12)
            : view === "week"
              ? weeklyApplied(apps, today, 12)
              : dailyApplied(apps, today, 30);
    $: rangeText =
        view === "month"
            ? "last 12 months"
            : view === "week"
              ? "last 12 weeks"
              : "last 30 days";
    $: yMax = view === "month" ? 10 : view === "week" ? 4 : 2;

    $: funnel = jobFunnel(apps, queue);
    $: streak = applyStreak(apps, today);
    $: todayCount = appliedInLastDays(apps, today, 1);
    $: weekCount = weeklyApplied(apps, today, 1)[0].value;
    $: perDay = appliedInLastDays(apps, today, 30) / 30;

    function plural(n, word) {
        return `${n} ${word}${n === 1 ? "" : "s"}`;
    }
</script>

<section class="habit-history job-stats">
    <div class="habit-history-head">
        <div class="habit-history-title">
            <span class="eyebrow">APPLICATIONS SENT · {rangeText}</span>
        </div>

        <div class="habit-history-views">
            {#each JOB_STATS_VIEWS as v}
                <button
                    type="button"
                    class:active={view === v}
                    on:click={() => onSelectView(v)}>{VIEW_LABELS[v]}</button
                >
            {/each}
        </div>
    </div>

    <div class="habit-history-stats job-stats-tiles">
        <div class="habit-stat">
            <span class="habit-stat-label">Today</span>
            <span class="habit-stat-value">{todayCount}</span>
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">This week</span>
            <span class="habit-stat-value">{weekCount}</span>
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Per day (30d)</span>
            <span class="habit-stat-value">{perDay.toFixed(1)}</span>
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Streak</span>
            <span class="habit-stat-value"
                >{plural(streak.current, "day")}
                <small>best {streak.best}</small></span
            >
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Sent of tracked</span>
            <span class="habit-stat-value"
                >{funnel.appliedRate === null
                    ? "—"
                    : `${Math.round(funnel.appliedRate * 100)}%`}
                <small>{funnel.applied}/{funnel.applied + funnel.drafted}</small
                ></span
            >
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Skipped</span>
            <span class="habit-stat-value">{funnel.skipped}</span>
        </div>
    </div>

    <MiniColumnChart
        {buckets}
        {yMax}
        height={160}
        ariaLabel={`Applications sent per ${view}, ${rangeText}`}
    />
</section>

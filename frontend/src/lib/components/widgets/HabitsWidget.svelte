<script>
    // Dashboard summary of habits: today's progress, 30-day completion and
    // the best running streak, a 14-day strip per habit, and how many habits
    // were done each day over the same two weeks.
    import { onMount } from "svelte";
    import { ListHabits } from "../../../../wailsjs/go/main/App.js";
    import { startOfDay } from "../../calendarGrid.js";
    import { isDoneOn } from "../../habitDisplay.js";
    import {
        streaks,
        recentDays,
        dailyTotals,
        recentCompletionRate,
    } from "../../habitHistory.js";
    import MiniColumnChart from "./MiniColumnChart.svelte";

    export let focused = false;

    const DAYS = 14;
    const today = startOfDay(new Date());

    let habits = [];
    let loading = true;
    let error = null;
    let hoveredCell = null;

    onMount(async () => {
        try {
            habits = await ListHabits();
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    });

    $: doneToday = habits.filter((h) => isDoneOn(h, today)).length;
    $: rows = habits.map((habit) => ({
        habit,
        days: recentDays(habit, today, DAYS),
        streak: streaks(habit, today).current,
    }));
    $: bestStreak = rows.reduce((best, r) => Math.max(best, r.streak), 0);
    $: rate = recentCompletionRate(habits, today, 30);
    $: totals = dailyTotals(habits, today, DAYS);

    function cellFill(day) {
        if (day.done) return "var(--accent)";
        if (!day.tracked) return "var(--border-soft)";
        return "var(--border)";
    }

    function dayLabel(date) {
        return date.toLocaleDateString("en-US", {
            weekday: "short",
            month: "short",
            day: "numeric",
        });
    }
</script>

<div class="todo-widget" class:focused>
    <header class="todo-header">
        <span class="todo-header-title">Habits</span>
        <span class="todo-header-count"
            >{doneToday}/{habits.length} done today</span
        >
    </header>

    <div class="dash-body">
        {#if loading}
            <p>Loading…</p>
        {:else if error}
            <p class="todo-error">{error}</p>
        {:else if habits.length === 0}
            <div class="empty-widget">
                <span>空</span>
                <p>No habits yet — add one on the Habits tab</p>
            </div>
        {:else}
            <div class="dash-stats">
                <div class="dash-stat">
                    <span class="dash-stat-label">Done today</span>
                    <span class="dash-stat-value"
                        >{doneToday}<small>/{habits.length}</small></span
                    >
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">Last 30 days</span>
                    <span class="dash-stat-value"
                        >{rate === null ? "—" : `${Math.round(rate * 100)}%`}</span
                    >
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">Best running streak</span>
                    <span class="dash-stat-value"
                        >{bestStreak}<small> day{bestStreak === 1 ? "" : "s"}</small></span
                    >
                </div>
            </div>

            <div class="dash-split">
                <div class="dash-pane">
                    <div class="dash-subtitle">
                        {#if hoveredCell}
                            {hoveredCell.name} · {dayLabel(hoveredCell.day.date)} ·
                            {hoveredCell.day.done
                                ? "done"
                                : hoveredCell.day.tracked
                                  ? "not done"
                                  : "not tracked yet"}
                        {:else}
                            Last {DAYS} days
                        {/if}
                    </div>
                    <ul class="dash-strips">
                        {#each rows as row (row.habit.id)}
                            <li class="dash-strip">
                                <span class="dash-strip-name" title={row.habit.name}
                                    >{row.habit.name}</span
                                >
                                <span class="dash-strip-cells">
                                    {#each row.days as day (day.key)}
                                        <span
                                            class="dash-cell"
                                            class:hovered={hoveredCell?.day === day}
                                            style="background: {cellFill(day)}"
                                            role="presentation"
                                            on:pointerenter={() =>
                                                (hoveredCell = { name: row.habit.name, day })}
                                            on:pointerleave={() => (hoveredCell = null)}
                                        ></span>
                                    {/each}
                                </span>
                                <span class="dash-strip-streak">{row.streak}d</span>
                            </li>
                        {/each}
                    </ul>
                </div>

                <div class="dash-pane">
                    <div class="dash-subtitle">Habits done per day</div>
                    <MiniColumnChart
                        buckets={totals}
                        yMax={habits.length}
                                ariaLabel="Habits done per day, last {DAYS} days"
                    />
                </div>
            </div>
        {/if}
    </div>
</div>

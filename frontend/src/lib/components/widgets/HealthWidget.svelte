<script>
    // Dashboard summary of the 5/3/1 program: the next workout and its top
    // set, progress through the current cycle, a bar per lift's training
    // max, and the cycle's 16 workouts as a weeks × lifts grid.
    import { onMount } from "svelte";
    import { ListWorkoutCycles } from "../../../../wailsjs/go/main/App.js";
    import { startOfDay } from "../../calendarGrid.js";
    import {
        LIFTS,
        LIFT_META,
        WEEKS,
        WORKOUTS_PER_CYCLE,
        workoutAt,
        topSet,
        schedule,
        dueLabel,
        formatDay,
        bestEstimates,
    } from "../../workoutProgram.js";

    export let focused = false;

    const today = startOfDay(new Date());
    const BAR_H = 10;

    let cycles = [];
    let loading = true;
    let error = null;
    let hovered = null;
    let barsWidth = 0;

    onMount(async () => {
        try {
            cycles = await ListWorkoutCycles();
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }
    });

    $: cycle = [...cycles].sort((a, b) => b.number - a.number)[0];
    $: logged = cycle?.logs?.length ?? 0;
    $: totalLogged = cycles.reduce((sum, c) => sum + (c.logs?.length ?? 0), 0);
    $: plan = cycle ? schedule(cycle, today) : [];
    $: nextIndex = plan.findIndex((s) => s.status === "next");
    $: next = nextIndex >= 0 ? { ...workoutAt(nextIndex), ...plan[nextIndex] } : null;
    $: nextSet = next ? topSet(cycle.trainingMax[next.lift], next.week) : null;
    $: best = cycle ? bestEstimates(cycle) : {};
    $: maxTM = cycle ? Math.max(...LIFTS.map((l) => cycle.trainingMax[l] ?? 0)) : 0;

    // Label column, then the bar, then room for the value at its tip.
    const LABEL_W = 64;
    const VALUE_W = 56;
    $: barSpace = Math.max(barsWidth - LABEL_W - VALUE_W, 0);

    // A horizontal bar: square at the baseline, 4px rounded data-end.
    function barPath(w) {
        const r = Math.min(4, w / 2, BAR_H / 2);
        return `M0,0 H${w - r} Q${w},0 ${w},${r} V${BAR_H - r} Q${w},${BAR_H} ${w - r},${BAR_H} H0 Z`;
    }

    function cellState(index) {
        return plan[index]?.status ?? "upcoming";
    }

    function describe(index) {
        const { week, lift } = workoutAt(index);
        const step = plan[index];
        const name = `${LIFT_META[lift].label} · week ${week} (${WEEKS[week].label})`;
        if (!step) return name;
        if (step.status === "done") return `${name} · done ${formatDay(step.date)}`;
        return `${name} · ${step.status === "next" ? "next" : "planned"} ${formatDay(step.due)}`;
    }
</script>

<div class="todo-widget" class:focused>
    <header class="todo-header">
        <span class="todo-header-title">Health</span>
        {#if cycle}
            <span class="todo-header-count"
                >Cycle {cycle.number}{next ? ` · week ${next.week}` : ""}</span
            >
        {/if}
    </header>

    <div class="dash-body">
        {#if loading}
            <p>Loading…</p>
        {:else if error}
            <p class="todo-error">{error}</p>
        {:else if !cycle}
            <div class="empty-widget">
                <span>空</span>
                <p>Set up 5/3/1 on the Health tab</p>
            </div>
        {:else}
            <div class="dash-stats">
                <div class="dash-stat">
                    <span class="dash-stat-label">
                        Next{next ? `, ${dueLabel(next.due, today)}` : ""}
                    </span>
                    <span class="dash-stat-value">
                        {#if next}
                            {LIFT_META[next.lift].label}
                            <small>{nextSet.weight} × {nextSet.reps}{nextSet.amrap ? "+" : ""}</small>
                        {:else}
                            Cycle done
                        {/if}
                    </span>
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">This cycle</span>
                    <span class="dash-stat-value"
                        >{logged}<small>/{WORKOUTS_PER_CYCLE}</small></span
                    >
                </div>
                <div class="dash-stat">
                    <span class="dash-stat-label">Workouts logged</span>
                    <span class="dash-stat-value">{totalLogged}</span>
                </div>
            </div>

            <div class="dash-split">
                <div class="dash-pane">
                    <div class="dash-subtitle">Training max (lb)</div>
                    <ul class="dash-bars" bind:clientWidth={barsWidth}>
                        {#each LIFTS as lift}
                            {@const tm = cycle.trainingMax[lift] ?? 0}
                            {@const w = maxTM ? (tm / maxTM) * barSpace : 0}
                            <li
                                class="dash-bar-row"
                                title={best[lift]
                                    ? `Best estimated 1RM this cycle: ${best[lift]} lb`
                                    : ""}
                            >
                                <span class="dash-bar-label">{LIFT_META[lift].label}</span>
                                <svg width={barSpace + VALUE_W} height={BAR_H}>
                                    {#if w > 0}
                                        <path class="habit-chart-bar" d={barPath(w)} />
                                    {/if}
                                    <text
                                        class="habit-chart-cap"
                                        x={w + 6}
                                        y={BAR_H / 2}
                                        dominant-baseline="middle">{tm}</text
                                    >
                                </svg>
                            </li>
                        {/each}
                    </ul>
                </div>

                <div class="dash-pane">
                    <div class="dash-subtitle">
                        {hovered === null ? "Cycle progress" : describe(hovered)}
                    </div>
                    <div class="dash-cycle">
                        <span></span>
                        {#each LIFTS as lift}
                            <span class="dash-cycle-head">{LIFT_META[lift].label.slice(0, 2)}</span>
                        {/each}
                        {#each [1, 2, 3, 4] as week}
                            <span class="dash-cycle-week">{WEEKS[week].label}</span>
                            {#each LIFTS as _, l}
                                {@const index = (week - 1) * LIFTS.length + l}
                                <span
                                    class="dash-cycle-cell {cellState(index)}"
                                    class:hovered={hovered === index}
                                    role="presentation"
                                    on:pointerenter={() => (hovered = index)}
                                    on:pointerleave={() => (hovered = null)}
                                ></span>
                            {/each}
                        {/each}
                    </div>
                    <div class="habit-grid-legend">
                        <span class="habit-grid-swatch" style="background: var(--accent)"></span>
                        done
                        <span class="habit-grid-swatch dash-swatch-next"></span>
                        next
                        <span class="habit-grid-swatch" style="background: var(--border)"></span>
                        planned
                    </div>
                </div>
            </div>
        {/if}
    </div>
</div>

<script>
    // Projections for the 5/3/1 program: where the training maxes go if
    // every cycle is run on schedule. Shown in place of the week grid on
    // the Workouts page ("p" toggles it); the page forwards keys here
    // through handleKey, which returns true when it used the key.
    //
    // "Implied 1RM" is TM ÷ 0.9, the max the program assumes. "Pace" starts
    // from the best AMRAP e1RM actually logged and adds the same per-cycle
    // increment, so it shows how the projection looks from real lifting.
    import { tick } from "svelte";
    import {
        LIFTS,
        LIFT_META,
        TOTAL_LIFTS,
        projectCycles,
        projectionAt,
        projectionSeries,
        cycleToReach,
        nextMilestone,
        parseDateKey,
        isValidDateKey,
        formatLongDay,
        fromNow,
    } from "../workoutProgram.js";

    export let cycles;
    export let today;

    const HORIZONS = [6, 12, 18, 24];
    const HEIGHT = 220;
    const MARGIN = { top: 16, right: 44, bottom: 24, left: 40 };
    const GOAL_TARGETS = [...LIFTS, "total"];

    let horizon = 12;
    let chartLift = "squat";
    let width = 0;
    let hovered = null;

    let lookupDate = "";
    let goalLift = "squat";
    let goalWeight = "";
    let dateInput;
    let goalInput;

    $: latest = cycles.at(-1);
    $: rows = projectCycles(latest, today, horizon);
    $: series = projectionSeries(cycles, rows, chartLift);
    $: milestones = GOAL_TARGETS.map((lift) => ({ lift, next: nextMilestone(latest, today, lift) }));

    $: dateResult = isValidDateKey(lookupDate) ? projectionAt(latest, today, parseDateKey(lookupDate)) : null;
    $: goalResult = Number(goalWeight) > 0 ? cycleToReach(latest, today, goalLift, Number(goalWeight)) : null;

    $: clearHover(chartLift, horizon);

    function clearHover() {
        hovered = null;
    }

    function liftLabel(lift) {
        return lift === "total" ? "Total" : LIFT_META[lift].label;
    }

    function shortMonth(date) {
        return date.toLocaleDateString("en-US", { month: "short", year: "2-digit" }).replace(" ", " '");
    }

    export function handleKey(event) {
        switch (event.key) {
            case "h":
            case "l": {
                const i = LIFTS.indexOf(chartLift) + (event.key === "h" ? -1 : 1);
                if (i >= 0 && i < LIFTS.length) {
                    chartLift = LIFTS[i];
                }
                event.preventDefault();
                return true;
            }

            case "+":
            case "=":
            case "-": {
                const i = HORIZONS.indexOf(horizon) + (event.key === "-" ? -1 : 1);
                if (i >= 0 && i < HORIZONS.length) {
                    horizon = HORIZONS[i];
                }
                event.preventDefault();
                return true;
            }

            case "d":
                event.preventDefault();
                focus(dateInput);
                return true;

            case "g":
                event.preventDefault();
                focus(goalInput);
                return true;
        }
        return false;
    }

    async function focus(input) {
        await tick();
        input?.focus();
        input?.select();
    }

    function onFormKeydown(event) {
        if (event.key === "Escape") {
            event.preventDefault();
            document.activeElement?.blur();
        }
    }

    // ---- chart geometry ----
    $: plotW = Math.max(width - MARGIN.left - MARGIN.right, 0);
    $: plotH = HEIGHT - MARGIN.top - MARGIN.bottom;
    $: allPoints = [...series.program, ...(series.pace ?? []), ...series.actual];
    $: xMin = Math.min(...series.program.map((p) => p.date.getTime()));
    $: xMax = Math.max(...series.program.map((p) => p.date.getTime()));
    $: yStep = niceStep(Math.max(...allPoints.map((p) => p.value)) - Math.min(...allPoints.map((p) => p.value)));
    $: yMin = Math.floor((Math.min(...allPoints.map((p) => p.value)) - yStep / 2) / yStep) * yStep;
    $: yMax = Math.ceil((Math.max(...allPoints.map((p) => p.value)) + yStep / 2) / yStep) * yStep;
    $: yTicks = Array.from({ length: Math.round((yMax - yMin) / yStep) + 1 }, (_, i) => yMin + i * yStep);

    $: xScale = (date) =>
        MARGIN.left + (xMax > xMin ? ((date.getTime() - xMin) / (xMax - xMin)) * plotW : plotW / 2);
    $: yScale = (v) => MARGIN.top + plotH - ((v - yMin) / (yMax - yMin)) * plotH;
    $: linePath = (points) =>
        points.map((p, i) => `${i ? "L" : "M"}${xScale(p.date)},${yScale(p.value)}`).join(" ");

    // Thin out the month labels so they sit at least ~56px apart.
    $: labelEvery = Math.max(1, Math.ceil(56 / Math.max(plotW / Math.max(series.program.length - 1, 1), 1)));
    $: todayX = today.getTime() >= xMin && today.getTime() <= xMax ? xScale(today) : null;

    $: hoveredPoint = hovered === null ? null : series.program[hovered];
    $: hoveredPace = hoveredPoint && series.pace?.find((p) => p.number === hoveredPoint.number);
    $: hoveredActual = hoveredPoint && series.actual.find((p) => p.number === hoveredPoint.number);

    function niceStep(range) {
        return [10, 25, 50, 100].find((s) => range / s <= 5) ?? 200;
    }

    function onPointerMove(event) {
        const x = event.clientX - event.currentTarget.ownerSVGElement.getBoundingClientRect().left;
        let best = 0;
        series.program.forEach((p, i) => {
            if (Math.abs(xScale(p.date) - x) < Math.abs(xScale(series.program[best].date) - x)) {
                best = i;
            }
        });
        hovered = best;
    }

    function tooltipAlign(i, n) {
        if (i < 2) return "start";
        if (i > n - 3) return "end";
        return "center";
    }
</script>

<div class="projections">
    <div class="projections-milestones">
        {#each milestones as { lift, next }}
            <div class="workouts-max projections-milestone">
                <span class="workouts-max-label">{liftLabel(lift)} · next milestone</span>
                {#if next?.cycle}
                    <span class="workouts-max-value">{next.target} lb</span>
                    <span class="workouts-max-note"
                        >{formatLongDay(next.cycle.start)} · cycle {next.cycle.number} ·
                        {fromNow(next.cycle.start, today)}</span
                    >
                {:else}
                    <span class="workouts-max-value">—</span>
                    <span class="workouts-max-note">past every milestone</span>
                {/if}
            </div>
        {/each}
    </div>

    <section class="projections-chart-card">
        <div class="habit-history-head">
            <div class="habit-history-title">
                <span class="eyebrow">ESTIMATED 1RM · NEXT {horizon} CYCLES</span>
                <h2>{LIFT_META[chartLift].label}</h2>
            </div>

            <div class="habit-history-views" role="group" aria-label="Lift">
                {#each LIFTS as lift}
                    <button
                        type="button"
                        class:active={lift === chartLift}
                        on:click={() => (chartLift = lift)}>{LIFT_META[lift].label}</button
                    >
                {/each}
            </div>
        </div>

        <div class="projections-legend">
            <span><i class="projections-swatch program"></i>Program (TM ÷ 0.9)</span>
            {#if series.pace}
                <span><i class="projections-swatch pace"></i>Your pace (from AMRAP e1RM)</span>
                <span><i class="projections-swatch dot"></i>Logged e1RM</span>
            {:else}
                <span class="projections-legend-note">Log an AMRAP set to see your pace</span>
            {/if}
        </div>

        <div class="habit-history-chart" bind:clientWidth={width}>
            {#if width > 0}
                <svg
                    {width}
                    height={HEIGHT}
                    role="img"
                    aria-label="Projected {LIFT_META[chartLift].label} 1RM over the next {horizon} cycles"
                >
                    {#each yTicks as t}
                        <line
                            class="habit-chart-grid"
                            x1={MARGIN.left}
                            x2={width - MARGIN.right}
                            y1={yScale(t)}
                            y2={yScale(t)}
                        />
                        <text
                            class="habit-chart-tick"
                            x={MARGIN.left - 6}
                            y={yScale(t)}
                            text-anchor="end"
                            dominant-baseline="middle">{t}</text
                        >
                    {/each}

                    {#each series.program as p, i}
                        {#if i % labelEvery === 0}
                            <text
                                class="habit-chart-label"
                                x={xScale(p.date)}
                                y={HEIGHT - 6}
                                text-anchor="middle">{shortMonth(p.date)}</text
                            >
                        {/if}
                    {/each}

                    {#if todayX !== null}
                        <line
                            class="projections-today"
                            x1={todayX}
                            x2={todayX}
                            y1={MARGIN.top}
                            y2={MARGIN.top + plotH}
                        />
                        <text class="habit-chart-label" x={todayX + 4} y={MARGIN.top + 8}>today</text>
                    {/if}

                    {#if hoveredPoint}
                        <line
                            class="projections-crosshair"
                            x1={xScale(hoveredPoint.date)}
                            x2={xScale(hoveredPoint.date)}
                            y1={MARGIN.top}
                            y2={MARGIN.top + plotH}
                        />
                    {/if}

                    <path class="projections-line program" d={linePath(series.program)} />
                    {#if series.pace}
                        <path class="projections-line pace" d={linePath(series.pace)} />
                    {/if}

                    {#each series.actual as p}
                        <circle class="projections-dot" cx={xScale(p.date)} cy={yScale(p.value)} r="4" />
                    {/each}

                    {#if hoveredPoint}
                        <circle
                            class="projections-marker program"
                            cx={xScale(hoveredPoint.date)}
                            cy={yScale(hoveredPoint.value)}
                            r="4"
                        />
                        {#if hoveredPace}
                            <circle
                                class="projections-marker pace"
                                cx={xScale(hoveredPace.date)}
                                cy={yScale(hoveredPace.value)}
                                r="4"
                            />
                        {/if}
                    {/if}

                    <!-- Direct labels at the end of each line. -->
                    {#each [series.program.at(-1), series.pace?.at(-1)].filter(Boolean) as end}
                        <text
                            class="habit-chart-cap"
                            x={xScale(end.date) + 6}
                            y={yScale(end.value)}
                            dominant-baseline="middle">{end.value}</text
                        >
                    {/each}

                    <rect
                        class="habit-chart-hit"
                        x={MARGIN.left}
                        y={MARGIN.top}
                        width={plotW}
                        height={plotH}
                        role="presentation"
                        on:pointermove={onPointerMove}
                        on:pointerleave={() => (hovered = null)}
                    />
                </svg>

                {#if hoveredPoint}
                    <div
                        class="habit-chart-tooltip {tooltipAlign(hovered, series.program.length)}"
                        style="left: {xScale(hoveredPoint.date)}px; top: {MARGIN.top - 4}px"
                    >
                        <strong>Cycle {hoveredPoint.number}</strong>
                        <span>{formatLongDay(hoveredPoint.date)}</span>
                        <span>Program {hoveredPoint.value} lb</span>
                        {#if hoveredPace}<span>Pace {hoveredPace.value} lb</span>{/if}
                        {#if hoveredActual}<span>Logged e1RM {hoveredActual.value} lb</span>{/if}
                    </div>
                {/if}
            {/if}
        </div>
    </section>

    <div class="projections-lookups">
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions (esc leaves the form) -->
        <form
            class="projections-lookup"
            on:submit|preventDefault={() => document.activeElement?.blur()}
            on:keydown={onFormKeydown}
        >
            <label class="workouts-field">
                <span>By date</span>
                <input
                    class="workouts-input"
                    bind:this={dateInput}
                    bind:value={lookupDate}
                    placeholder="YYYY-MM-DD"
                />
            </label>
            <p class="projections-result">
                {#if dateResult}
                    Cycle {dateResult.number}, from {formatLongDay(dateResult.start)}:
                    {#each LIFTS as lift, i}
                        {LIFT_META[lift].label}
                        <strong>{dateResult.implied[lift]}</strong>{i < LIFTS.length - 1 ? " · " : ""}
                    {/each}
                    · total <strong>{dateResult.total}</strong>
                {:else if lookupDate}
                    Enter a date as YYYY-MM-DD
                {:else}
                    What you'd be lifting by a date (implied 1RM)
                {/if}
            </p>
        </form>

        <!-- svelte-ignore a11y_no_noninteractive_element_interactions (esc leaves the form) -->
        <form
            class="projections-lookup"
            on:submit|preventDefault={() => document.activeElement?.blur()}
            on:keydown={onFormKeydown}
        >
            <div class="workouts-field">
                <span>Goal</span>
                <select class="workouts-input" bind:value={goalLift} aria-label="Goal lift">
                    {#each GOAL_TARGETS as lift}
                        <option value={lift}>{liftLabel(lift)}</option>
                    {/each}
                </select>
                <input
                    class="workouts-input"
                    type="number"
                    min="0"
                    step="5"
                    placeholder="lb"
                    aria-label="Goal weight"
                    bind:this={goalInput}
                    bind:value={goalWeight}
                />
            </div>
            <p class="projections-result">
                {#if goalResult?.k === 0}
                    {liftLabel(goalLift)} {goalWeight} lb: you're already there this cycle
                {:else if goalResult}
                    {liftLabel(goalLift)} {goalWeight} lb in cycle {goalResult.number}, starting
                    <strong>{formatLongDay(goalResult.start)}</strong>
                    ({fromNow(goalResult.start, today)})
                {:else if Number(goalWeight) > 0}
                    More than 40 years out on this program
                {:else}
                    When you'd reach a 1RM{goalLift === "total" ? " total" : ""}
                {/if}
            </p>
        </form>
    </div>

    <table class="projections-table">
        <thead>
            <tr>
                <th>Cycle</th>
                <th>Starts</th>
                {#each LIFTS as lift}
                    <th>{LIFT_META[lift].label}</th>
                {/each}
                <th>Total</th>
            </tr>
        </thead>
        <tbody>
            {#each rows as row (row.number)}
                <tr
                    class:current={row.k === 0}
                    class:highlight={row.number === dateResult?.number ||
                        row.number === goalResult?.number}
                >
                    <td class="projections-cycle">
                        {row.number}{#if row.k === 0}<small>now</small>{/if}
                    </td>
                    <td>{formatLongDay(row.start)}</td>
                    {#each LIFTS as lift}
                        <td>
                            <span class="projections-value">{row.implied[lift]}</span>
                            <small>TM {row.trainingMax[lift]} · 1×{row.topSingle[lift]}</small>
                        </td>
                    {/each}
                    <td><span class="projections-value">{row.total}</span></td>
                </tr>
            {/each}
        </tbody>
    </table>
    <p class="workouts-form-hint">
        1RMs are implied by the training max (TM ÷ 0.9). "1×" is the week-3 top single
        (95% TM). Total is {TOTAL_LIFTS.map((l) => LIFT_META[l].label.toLowerCase()).join(" + ")}.
    </p>
</div>

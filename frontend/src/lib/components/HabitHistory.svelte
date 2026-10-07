<script>
    // History chart for one habit. "week" and "month" plot one column per
    // period (how many days the habit was done in it); "year" is a
    // contribution-style grid of the last 53 weeks, one cell per day.
    import {
        weeklyBuckets,
        monthlyBuckets,
        yearGrid,
        streaks,
        completionRate,
        HISTORY_VIEWS,
        HISTORY_VIEW_LABELS,
    } from "../habitHistory.js";

    export let habit;
    export let today;
    export let view = "week";
    export let onSelectView = () => {};

    const HEIGHT = 180;
    const MARGIN = { top: 20, right: 8, bottom: 24, left: 28 };
    const WEEKDAY_ROWS = ["Mon", "", "Wed", "", "Fri", "", ""];

    let width = 0;
    let hovered = null;

    $: buckets =
        view === "month"
            ? monthlyBuckets(habit, today, 12)
            : weeklyBuckets(habit, today, 12);
    $: grid = view === "year" ? yearGrid(habit, today, 53) : [];
    $: ({ current: currentStreak, best: bestStreak } = streaks(habit, today));
    $: rate = view === "year" ? yearRate(grid) : completionRate(buckets);
    $: rangeText =
        view === "week"
            ? "last 12 weeks"
            : view === "month"
              ? "last 12 months"
              : "last 53 weeks";
    // Drop any hover when switching view or habit. A function call keeps
    // `hovered` out of this statement's dependencies.
    $: clearHover(view, habit.id);

    function clearHover() {
        hovered = null;
    }

    function yearRate(columns) {
        const days = columns
            .flatMap((c) => c.days)
            .filter((d) => d.tracked && !d.future);
        return days.length ? days.filter((d) => d.done).length / days.length : null;
    }

    function plural(n, word) {
        return `${n} ${word}${n === 1 ? "" : "s"}`;
    }

    // ---- column chart geometry ----
    $: plotW = Math.max(width - MARGIN.left - MARGIN.right, 0);
    $: plotH = HEIGHT - MARGIN.top - MARGIN.bottom;
    $: yMax = view === "month" ? 31 : 7;
    $: ticks = view === "month" ? [0, 10, 20, 30] : [0, 7];
    $: band = buckets.length ? plotW / buckets.length : 0;
    $: barW = Math.min(24, band * 0.6);
    // Thin out x labels when the bands are too narrow to fit them all.
    $: labelEvery = Math.max(1, Math.ceil(44 / Math.max(band, 1)));

    $: yScale = (v) => MARGIN.top + plotH - (v / yMax) * plotH;
    $: barX = (i) => MARGIN.left + band * i + (band - barW) / 2;

    // A column with a 4px rounded data-end and a square baseline.
    function columnPath(x, w, top, base) {
        const r = Math.min(4, w / 2, base - top);
        return `M${x},${base} V${top + r} Q${x},${top} ${x + r},${top} H${x + w - r} Q${x + w},${top} ${x + w},${top + r} V${base} Z`;
    }

    function tooltipAlign(i, n) {
        if (i < 2) return "start";
        if (i > n - 3) return "end";
        return "center";
    }

    // ---- year grid geometry ----
    const GRID_LABEL_W = 28;
    const GRID_TOP = 16;
    $: cell = Math.max(6, Math.min(12, Math.floor((width - GRID_LABEL_W) / 53) - 2));
    $: step = cell + 2;
    $: gridHeight = GRID_TOP + step * 7;
    $: monthMarks = grid
        .map((col, i) => {
            const first = col.days.find((d) => d.date.getDate() === 1);
            return first && i > 0
                ? { i, label: first.date.toLocaleDateString("en-US", { month: "short" }) }
                : null;
        })
        .filter(Boolean);

    function cellFill(day) {
        if (day.done) return "var(--accent)";
        if (!day.tracked) return "var(--border-soft)";
        return "var(--border)";
    }

    function longDate(date) {
        return date.toLocaleDateString("en-US", {
            weekday: "short",
            month: "short",
            day: "numeric",
            year: "numeric",
        });
    }
</script>

<section class="habit-history">
    <div class="habit-history-head">
        <div class="habit-history-title">
            <span class="eyebrow">HISTORY</span>
            <h2 title={habit.name}>{habit.name}</h2>
        </div>

        <div class="habit-history-views" role="group" aria-label="History view">
            {#each HISTORY_VIEWS as v}
                <button
                    type="button"
                    class:active={v === view}
                    on:click={() => onSelectView(v)}>{HISTORY_VIEW_LABELS[v]}</button
                >
            {/each}
        </div>
    </div>

    <div class="habit-history-stats">
        <div class="habit-stat">
            <span class="habit-stat-label">Current streak</span>
            <span class="habit-stat-value">{plural(currentStreak, "day")}</span>
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Best streak</span>
            <span class="habit-stat-value">{plural(bestStreak, "day")}</span>
        </div>
        <div class="habit-stat">
            <span class="habit-stat-label">Completion, {rangeText}</span>
            <span class="habit-stat-value"
                >{rate === null ? "—" : `${Math.round(rate * 100)}%`}</span
            >
        </div>
    </div>

    <div class="habit-history-subtitle">
        {#if view === "week"}
            Days completed per week
        {:else if view === "month"}
            Days completed per month
        {:else}
            Each square is a day; columns are weeks
        {/if}
    </div>

    <div class="habit-history-chart" bind:clientWidth={width}>
        {#if width > 0 && view !== "year"}
            <svg {width} height={HEIGHT} role="img" aria-label="Days completed per {view}">
                {#each ticks as t}
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

                {#each buckets as b, i (b.key)}
                    {@const x = barX(i)}
                    {@const base = yScale(0)}
                    {#if b.tracked && b.possible > 0}
                        <!-- Ghost track: the days that could have been done. -->
                        <path
                            class="habit-chart-track"
                            d={columnPath(x, barW, yScale(b.possible), base)}
                        />
                    {/if}
                    {#if b.count > 0}
                        <path
                            class="habit-chart-bar"
                            class:hovered={hovered === i}
                            d={columnPath(x, barW, yScale(b.count), base)}
                        />
                    {/if}
                    {#if b.current || hovered === i}
                        <text
                            class="habit-chart-cap"
                            x={x + barW / 2}
                            y={yScale(Math.max(b.count, b.possible)) - 6}
                            text-anchor="middle">{b.count}</text
                        >
                    {/if}
                    {#if i % labelEvery === (buckets.length - 1) % labelEvery}
                        <text
                            class="habit-chart-label"
                            class:untracked={!b.tracked}
                            x={x + barW / 2}
                            y={HEIGHT - 6}
                            text-anchor="middle">{b.label}</text
                        >
                    {/if}
                    <rect
                        class="habit-chart-hit"
                        x={MARGIN.left + band * i}
                        y={MARGIN.top}
                        width={band}
                        height={plotH}
                        role="presentation"
                        on:pointerenter={() => (hovered = i)}
                        on:pointerleave={() => (hovered = null)}
                    />
                {/each}
            </svg>

            {#if hovered !== null && buckets[hovered]}
                {@const b = buckets[hovered]}
                <div
                    class="habit-chart-tooltip {tooltipAlign(hovered, buckets.length)}"
                    style="left: {barX(hovered) + barW / 2}px; top: {yScale(
                        Math.max(b.count, b.possible),
                    ) - 8}px"
                >
                    {#if b.tracked}
                        <strong>{b.count} / {b.possible} days</strong>
                    {:else}
                        <strong>Not tracked yet</strong>
                    {/if}
                    <span>{b.rangeLabel}{b.current ? " (so far)" : ""}</span>
                </div>
            {/if}
        {:else if width > 0}
            <svg {width} height={gridHeight} role="img" aria-label="Daily completions, last 53 weeks">
                {#each monthMarks as m}
                    <text class="habit-chart-label" x={GRID_LABEL_W + m.i * step} y={10}
                        >{m.label}</text
                    >
                {/each}
                {#each WEEKDAY_ROWS as label, d}
                    {#if label}
                        <text
                            class="habit-chart-tick"
                            x={GRID_LABEL_W - 6}
                            y={GRID_TOP + d * step + cell / 2}
                            text-anchor="end"
                            dominant-baseline="middle">{label}</text
                        >
                    {/if}
                {/each}
                {#each grid as col, w (col.key)}
                    {#each col.days as day, d (day.key)}
                        {#if !day.future}
                            <rect
                                class="habit-grid-cell"
                                class:hovered={hovered === day.key}
                                x={GRID_LABEL_W + w * step}
                                y={GRID_TOP + d * step}
                                width={cell}
                                height={cell}
                                rx="2"
                                fill={cellFill(day)}
                                role="presentation"
                                on:pointerenter={() => (hovered = day.key)}
                                on:pointerleave={() => (hovered = null)}
                            />
                        {/if}
                    {/each}
                {/each}
            </svg>

            {#if hovered !== null}
                {@const w = grid.findIndex((c) => c.days.some((d) => d.key === hovered))}
                {@const d = w >= 0 ? grid[w].days.findIndex((x) => x.key === hovered) : -1}
                {#if w >= 0}
                    {@const day = grid[w].days[d]}
                    <div
                        class="habit-chart-tooltip {tooltipAlign(w, grid.length)}"
                        style="left: {GRID_LABEL_W + w * step + cell / 2}px; top: {GRID_TOP +
                            d * step -
                            6}px"
                    >
                        <strong
                            >{day.done ? "Done" : day.tracked ? "Not done" : "Not tracked yet"}</strong
                        >
                        <span>{longDate(day.date)}</span>
                    </div>
                {/if}
            {/if}

            <div class="habit-grid-legend">
                <span class="habit-grid-swatch" style="background: var(--border)"></span>
                not done
                <span class="habit-grid-swatch" style="background: var(--accent)"></span>
                done
            </div>
        {/if}
    </div>
</section>

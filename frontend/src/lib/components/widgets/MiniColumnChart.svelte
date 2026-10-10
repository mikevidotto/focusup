<script>
    // A compact single-series column chart for dashboard widgets. Each
    // bucket is { key, label, value, track?, title, sub, current? }:
    // `track` draws a ghost column behind the value (e.g. what was
    // possible), `title`/`sub` fill the hover tooltip, and the current
    // bucket's value is labelled on its cap. Leave `height` unset to fill
    // the space the parent gives the chart.
    export let buckets = [];
    export let yMax = 1;
    export let height = null;
    export let ariaLabel = "";

    const MARGIN = { top: 18, right: 4, bottom: 18, left: 22 };

    let width = 0;
    let boxHeight = 0;
    let hovered = null;

    $: h = height ?? Math.max(boxHeight, 80);

    $: max = Math.max(yMax, ...buckets.map((b) => Math.max(b.value, b.track ?? 0)), 1);
    $: ticks = niceTicks(max);
    $: plotW = Math.max(width - MARGIN.left - MARGIN.right, 0);
    $: plotH = h - MARGIN.top - MARGIN.bottom;
    $: band = buckets.length ? plotW / buckets.length : 0;
    $: barW = Math.min(24, band * 0.6);
    $: labelEvery = Math.max(1, Math.ceil(40 / Math.max(band, 1)));
    $: yScale = (v) => MARGIN.top + plotH - (v / ticks.at(-1)) * plotH;
    $: barX = (i) => MARGIN.left + band * i + (band - barW) / 2;

    // Evenly spaced clean ticks from 0 up to a round top value: 0, the
    // midpoint when it's a whole number, and the top.
    function niceTicks(top) {
        const step = top <= 4 ? 1 : top <= 10 ? 2 : top <= 25 ? 5 : 10;
        const end = Math.ceil(top / step) * step;
        const half = end / 2;
        return Number.isInteger(half) ? [0, half, end] : [0, end];
    }

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
</script>

<div
    class="habit-history-chart mini-chart"
    class:fill={height === null}
    bind:clientWidth={width}
    bind:clientHeight={boxHeight}
>
    {#if width > 0}
        <svg {width} height={h} role="img" aria-label={ariaLabel}>
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
                {#if b.track}
                    <path
                        class="habit-chart-track"
                        d={columnPath(x, barW, yScale(b.track), base)}
                    />
                {/if}
                {#if b.value > 0}
                    <path
                        class="habit-chart-bar"
                        class:hovered={hovered === i}
                        d={columnPath(x, barW, yScale(b.value), base)}
                    />
                {/if}
                {#if (b.current && hovered === null) || hovered === i}
                    <text
                        class="habit-chart-cap"
                        x={x + barW / 2}
                        y={yScale(Math.max(b.value, b.track ?? 0)) - 5}
                        text-anchor="middle">{b.value}</text
                    >
                {/if}
                {#if i % labelEvery === (buckets.length - 1) % labelEvery}
                    <text
                        class="habit-chart-label"
                        x={x + barW / 2}
                        y={h - 4}
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
                    Math.max(b.value, b.track ?? 0),
                ) - 18}px"
            >
                <strong>{b.title}</strong>
                <span>{b.sub}</span>
            </div>
        {/if}
    {/if}
</div>

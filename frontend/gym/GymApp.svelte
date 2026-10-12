<script>
    /*

Gym page: the 5/3/1 program on a phone. Shows the next workout with every
set and weight, logs it (with the reps from the AMRAP set), and lists what's
left this week. Talks to cmd/gymserver's /api, the same data the desktop
app uses once its workout server is set; the set math is the desktop's own
lib/workoutProgram.js.

*/
    import { onMount } from "svelte";
    import { startOfDay } from "../src/lib/calendarGrid.js";
    import {
        LIFT_META,
        WEEKS,
        WORKOUTS_PER_CYCLE,
        setsFor,
        topSet,
        workoutAt,
        isDeload,
        schedule,
        dueLabel,
        formatDay,
        parseDateKey,
        isValidDateKey,
        estimated1RM,
        workoutEstimate,
        toDateKey,
    } from "../src/lib/workoutProgram.js";

    let today = startOfDay(new Date());
    let cycles = [];
    let loading = true;
    let saving = false;
    let error = null;
    let notice = null;

    let date = toDateKey(today);
    let reps = "";

    onMount(() => {
        load();
        // Coming back to the tab (or the home-screen app) re-reads, in case
        // the desktop changed something in the meantime.
        const onVisible = () => document.visibilityState === "visible" && load();
        document.addEventListener("visibilitychange", onVisible);
        return () => document.removeEventListener("visibilitychange", onVisible);
    });

    async function api(method, path, body) {
        const res = await fetch(path, {
            method,
            headers: { "Content-Type": "application/json" },
            body: body ? JSON.stringify(body) : undefined,
        });
        if (!res.ok) {
            const data = await res.json().catch(() => null);
            throw new Error(data?.error ?? res.statusText);
        }
        return res.status === 204 ? null : res.json();
    }

    async function load() {
        try {
            today = startOfDay(new Date());
            cycles = await api("GET", "/api/cycles");
            error = null;
        } catch (e) {
            error = `Couldn't load the program: ${e.message}`;
        } finally {
            loading = false;
        }
    }

    $: cycle = cycles.at(-1);
    $: plan = cycle ? schedule(cycle, today) : [];
    $: index = cycle ? cycle.logs.length : 0;
    $: hasNext = cycle && index < WORKOUTS_PER_CYCLE;
    $: next = hasNext ? { ...workoutAt(index), ...plan[index] } : null;
    $: tm = next ? cycle.trainingMax[next.lift] : 0;
    $: sets = next ? setsFor(tm, next.week) : [];
    $: deload = hasNext && isDeload(index);
    $: liveEstimate = next && !deload ? estimated1RM(topSet(tm, next.week).weight, Number(reps)) : null;

    // The rest of the week the next workout is in.
    $: weekRest = next
        ? plan
              .map((step, i) => ({ ...step, ...workoutAt(i), index: i }))
              .filter((w) => w.week === next.week && w.index > index)
        : [];

    $: lastLog = lastLogged(cycles);

    function lastLogged(all) {
        for (let c = all.length - 1; c >= 0; c--) {
            const logs = all[c].logs ?? [];
            if (logs.length) {
                return { cycle: all[c], log: logs.at(-1) };
            }
        }
        return null;
    }

    // Warm-ups and work sets get a row each; the 5×10 volume sets share one.
    $: rows = sets.filter((set, i) => set.kind !== "volume" || sets[i - 1]?.kind !== "volume");

    async function logWorkout() {
        if (!isValidDateKey(date)) {
            error = "Pick a date";
            return;
        }
        const amrap = deload ? 0 : Number(reps || 0);
        if (!Number.isInteger(amrap) || amrap < 0) {
            error = "Reps must be a whole number";
            return;
        }

        const done = next;
        const estimate = liveEstimate;
        saving = true;
        try {
            cycles = await api("POST", "/api/complete", {
                cycleID: cycle.id,
                index,
                date,
                amrapReps: amrap,
            });
            notice = `Logged ${LIFT_META[done.lift].label}${estimate ? ` · e1RM ${estimate} lb` : ""}`;
            reps = "";
            date = toDateKey(today);
            error = null;
        } catch (e) {
            error = e.message;
        } finally {
            saving = false;
        }
    }

    async function undo() {
        if (!lastLog) {
            return;
        }
        const { label } = LIFT_META[workoutAt(lastLog.log.index).lift];
        if (!confirm(`Undo ${label} on ${formatDay(parseDateKey(lastLog.log.date))}?`)) {
            return;
        }
        try {
            cycles = await api("POST", "/api/undo");
            notice = `Removed ${label}`;
            error = null;
        } catch (e) {
            error = e.message;
        }
    }

    function kindLabel(set) {
        if (set.kind === "warmup") return "Warm-up";
        if (set.kind === "volume") return "5 × 10";
        return set.amrap ? "AMRAP" : "Work";
    }

    function repsLabel(set) {
        if (set.kind === "volume") return "5 sets × 10";
        return `× ${set.reps}${set.amrap ? "+" : ""}`;
    }
</script>

<main>
    <header>
        <span class="eyebrow">FOCUSUP / GYM</span>
        {#if cycle}
            <span class="meta">Cycle {cycle.number} · {index}/{WORKOUTS_PER_CYCLE}</span>
        {/if}
    </header>

    {#if loading}
        <p class="muted">Loading…</p>
    {:else if !cycle}
        <p class="muted">{error ?? "No program yet. Set up 5/3/1 on the Health tab in FocusUp."}</p>
    {:else if next}
        <section class="card">
            <span class="eyebrow">NEXT UP · {dueLabel(next.due, today).toUpperCase()}</span>
            <h1>{LIFT_META[next.lift].label}</h1>
            <p class="meta">
                Week {next.week} · {WEEKS[next.week].label} · TM {tm} lb
            </p>
            {#if next.lateDays}
                <p class="warning">
                    {next.lateDays} day{next.lateDays === 1 ? "" : "s"} behind plan
                </p>
            {/if}

            <table>
                <tbody>
                    {#each rows as set}
                        <tr class={set.kind} class:amrap={set.amrap}>
                            <td class="kind">{kindLabel(set)}</td>
                            <td class="weight">{set.weight}<small> lb</small></td>
                            <td class="reps">{repsLabel(set)}</td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </section>

        <form class="card" on:submit|preventDefault={logWorkout}>
            {#if !deload}
                <label>
                    <span>Reps on the last set</span>
                    <input
                        type="number"
                        inputmode="numeric"
                        min="0"
                        step="1"
                        placeholder="0"
                        bind:value={reps}
                    />
                </label>
                {#if liveEstimate}
                    <p class="meta">e1RM {liveEstimate} lb</p>
                {/if}
            {/if}
            <label>
                <span>Date</span>
                <input type="date" bind:value={date} />
            </label>
            <button type="submit" disabled={saving}>{saving ? "Saving…" : "Log workout"}</button>
        </form>

        {#if weekRest.length}
            <section class="card">
                <span class="eyebrow">REST OF THE WEEK</span>
                <ul>
                    {#each weekRest as w}
                        {@const top = topSet(cycle.trainingMax[w.lift], w.week)}
                        <li>
                            <span>{LIFT_META[w.lift].label}</span>
                            <span class="muted">{formatDay(w.due)}</span>
                            <span class="weight">{top.weight} × {top.reps}{top.amrap ? "+" : ""}</span>
                        </li>
                    {/each}
                </ul>
            </section>
        {/if}
    {/if}

    {#if notice}
        <p class="notice">{notice}</p>
    {/if}
    {#if error && cycle}
        <p class="warning">{error}</p>
    {/if}

    {#if lastLog}
        <footer>
            <span class="muted">
                Last: {LIFT_META[workoutAt(lastLog.log.index).lift].label},
                {formatDay(parseDateKey(lastLog.log.date))}{workoutEstimate(lastLog.cycle, lastLog.log)
                    ? ` · ${lastLog.log.amrapReps} reps`
                    : ""}
            </span>
            <button class="link" type="button" on:click={undo}>Undo</button>
        </footer>
    {/if}
</main>

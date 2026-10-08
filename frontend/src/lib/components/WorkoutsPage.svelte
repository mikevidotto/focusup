<script>
    /*

Workouts page: runs a 5/3/1 program.

The backend stores cycles (start date, training max per lift, logged
workouts) and generates the next cycle when the 16th workout of the
current one is logged. Everything per set, and the projected schedule,
comes from lib/workoutProgram.js.

Keyboard (same tabs <-> grid scheme as the Calendar page):
- from the tab bar, j enters the workout grid (keyboard mode "grid")
- j/k move down/up between weeks, h/l move left/right between the lifts
  of that week; k on week 1, q or esc hand the keyboard back to the tabs
- enter logs the next workout (or edits a done one); enter in the form
  saves, esc cancels
- e edit a done workout, c edit the cycle, u undo the last log
- [ ] previous/next cycle, x x reset the whole program
- p toggles the projections view (WorkoutProjections.svelte), which takes
  h/l, +/-, d and g while it's showing; k, q or esc still go back to the tabs

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import { startOfDay } from "../calendarGrid.js";
    import WorkoutProjections from "./WorkoutProjections.svelte";
    import {
        LIFTS,
        LIFT_META,
        WEEKS,
        WORKOUTS_PER_CYCLE,
        setsFor,
        topSet,
        workoutAt,
        isDeload,
        estimated1RM,
        workoutEstimate,
        bestEstimates,
        trainingMaxFrom1RM,
        schedule,
        dueLabel,
        formatDay,
        parseDateKey,
        isValidDateKey,
        toDateKey,
    } from "../workoutProgram.js";
    import {
        ListWorkoutCycles,
        SetupWorkouts,
        CompleteWorkout,
        UpdateWorkoutCycle,
        UpdateWorkoutLog,
        UndoLastWorkout,
        ResetWorkouts,
    } from "../../../wailsjs/go/main/App.js";

    const today = startOfDay(new Date());
    const todayKey = toDateKey(today);

    let cycles = [];
    let viewIndex = 0;
    let cursor = 0;
    let loading = true;
    let error = null;
    let pendingReset = false;

    // "program" (the week grid) | "projections"
    let view = "program";
    let projectionsEl;

    // null | "log" | "edit" | "cycle"
    let formMode = null;
    let formEl;
    let formDate = todayKey;
    let formReps = "";
    let cycleStart = "";
    let cycleTM = {};

    let setupStart = todayKey;
    let setupMax = { squat: "", bench: "", deadlift: "", press: "" };
    let setupEl;

    onMount(async () => {
        try {
            setCycles(await ListWorkoutCycles());
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }

        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
    });

    // Replaces the cycle list and jumps to the latest cycle's next workout.
    function setCycles(next) {
        cycles = next ?? [];
        viewIndex = Math.max(cycles.length - 1, 0);
        const latest = cycles[viewIndex];
        cursor = latest ? Math.min(latest.logs.length, WORKOUTS_PER_CYCLE - 1) : 0;
    }

    function replaceCycle(updated) {
        cycles = cycles.map((c) => (c.id === updated.id ? updated : c));
    }

    // The grid owns the keyboard only after "j" from the tab bar, like the
    // Calendar page; until then the cursor isn't drawn.
    $: navigating = $mode === "grid";

    $: cycle = cycles[viewIndex];
    $: isLatest = viewIndex === cycles.length - 1;
    $: plan = cycle ? schedule(cycle, today) : [];
    $: best = cycle ? bestEstimates(cycle) : {};
    $: doneCount = cycle?.logs.length ?? 0;
    $: selected = workoutAt(cursor);
    $: selectedStatus = plan[cursor];
    $: selectedLog = cycle?.logs[cursor];
    $: selectedSets = cycle ? setsFor(cycle.trainingMax[selected.lift], selected.week) : [];
    $: canLog = isLatest && cursor === doneCount && doneCount < WORKOUTS_PER_CYCLE;
    $: liveEstimate =
        cycle && !isDeload(cursor)
            ? estimated1RM(topSet(cycle.trainingMax[selected.lift], selected.week).weight, Number(formReps))
            : null;

    // ---- setup ----
    $: setupTMs = Object.fromEntries(
        LIFTS.map((lift) => [lift, Number(setupMax[lift]) > 0 ? trainingMaxFrom1RM(Number(setupMax[lift])) : null]),
    );

    async function submitSetup() {
        if (!isValidDateKey(setupStart)) {
            error = "Start date must be YYYY-MM-DD";
            return;
        }
        if (LIFTS.some((lift) => !(Number(setupMax[lift]) > 0))) {
            error = "Enter a 1RM for every lift";
            return;
        }

        try {
            const oneRepMax = Object.fromEntries(LIFTS.map((lift) => [lift, Number(setupMax[lift])]));
            const created = await SetupWorkouts(setupStart, oneRepMax);
            setCycles([created]);
            error = null;
            document.activeElement?.blur();
        } catch (e) {
            error = String(e);
        }
    }

    // ---- workout / cycle forms ----
    async function openForm(mode) {
        formMode = mode;
        error = null;

        if (mode === "log") {
            formDate = todayKey;
            formReps = "";
        } else if (mode === "edit") {
            formDate = selectedLog.date;
            formReps = isDeload(cursor) ? "" : String(selectedLog.amrapReps);
        } else if (mode === "cycle") {
            cycleStart = cycle.startDate;
            cycleTM = { ...cycle.trainingMax };
        }

        await tick();
        // Focus reps when there are reps to enter, otherwise the first field.
        const target =
            formEl?.querySelector("input[data-autofocus]") ?? formEl?.querySelector("input");
        target?.focus();
        target?.select();
    }

    function closeForm() {
        formMode = null;
        document.activeElement?.blur();
    }

    function onFormKeydown(event) {
        if (event.key === "Escape") {
            event.preventDefault();
            closeForm();
        }
    }

    async function submitForm() {
        if (formMode === "cycle") {
            return submitCycle();
        }

        if (!isValidDateKey(formDate)) {
            error = "Date must be YYYY-MM-DD";
            return;
        }
        const reps = isDeload(cursor) ? 0 : Number(formReps || 0);
        if (!Number.isInteger(reps) || reps < 0) {
            error = "Reps must be a whole number";
            return;
        }

        try {
            if (formMode === "log") {
                setCycles(await CompleteWorkout(cycle.id, cursor, formDate, reps));
            } else {
                replaceCycle(await UpdateWorkoutLog(cycle.id, cursor, formDate, reps));
            }
            error = null;
            closeForm();
        } catch (e) {
            error = String(e);
        }
    }

    async function submitCycle() {
        if (!isValidDateKey(cycleStart)) {
            error = "Start date must be YYYY-MM-DD";
            return;
        }
        const tm = Object.fromEntries(LIFTS.map((lift) => [lift, Number(cycleTM[lift])]));
        if (LIFTS.some((lift) => !(tm[lift] > 0))) {
            error = "Every training max must be greater than 0";
            return;
        }

        try {
            replaceCycle(await UpdateWorkoutCycle(cycle.id, cycleStart, tm));
            error = null;
            closeForm();
        } catch (e) {
            error = String(e);
        }
    }

    async function undo() {
        try {
            setCycles(await UndoLastWorkout());
            error = null;
        } catch (e) {
            error = String(e);
        }
    }

    async function reset() {
        try {
            await ResetWorkouts();
            setCycles([]);
            error = null;
        } catch (e) {
            error = String(e);
        }
    }

    function handleKey(event) {
        if (formMode) {
            return;
        }

        // Before setup, enter jumps into the setup form. It isn't focused
        // on mount because focused inputs swallow the tab shortcuts.
        if (!cycle) {
            if (event.key === "Enter" && !loading) {
                event.preventDefault();
                setupEl?.querySelector("input")?.focus();
            } else if (navigating && ["k", "q", "Escape"].includes(event.key)) {
                event.preventDefault();
                mode.set("tabs");
            }
            return;
        }

        if (!navigating) {
            return;
        }

        // "x" arms a reset; the next key confirms (x) or cancels it.
        if (pendingReset) {
            pendingReset = false;
            if (event.key === "x") {
                event.preventDefault();
                reset();
                return;
            }
        }

        if (event.key === "p") {
            event.preventDefault();
            view = view === "program" ? "projections" : "program";
            return;
        }

        if (view === "projections") {
            if (projectionsEl?.handleKey(event)) {
                return;
            }
            if (["k", "q", "Escape"].includes(event.key)) {
                event.preventDefault();
                mode.set("tabs");
            }
            return;
        }

        switch (event.key) {
            case "j":
                event.preventDefault();
                if (cursor + 4 < WORKOUTS_PER_CYCLE) {
                    cursor += 4;
                }
                break;

            case "k":
                event.preventDefault();
                if (cursor - 4 >= 0) {
                    cursor -= 4;
                } else {
                    mode.set("tabs");
                }
                break;

            case "h":
                event.preventDefault();
                if (cursor % 4 > 0) {
                    cursor -= 1;
                }
                break;

            case "l":
                event.preventDefault();
                if (cursor % 4 < 3) {
                    cursor += 1;
                }
                break;

            case "q":
            case "Escape":
                event.preventDefault();
                mode.set("tabs");
                break;

            case "[":
            case "]": {
                event.preventDefault();
                const next = viewIndex + (event.key === "[" ? -1 : 1);
                if (next >= 0 && next < cycles.length) {
                    viewIndex = next;
                    const c = cycles[next];
                    cursor = Math.min(c.logs.length, WORKOUTS_PER_CYCLE - 1);
                }
                break;
            }

            case "Enter":
                event.preventDefault();
                if (canLog) {
                    openForm("log");
                } else if (selectedLog) {
                    openForm("edit");
                }
                break;

            case "e":
                if (selectedLog) {
                    event.preventDefault();
                    openForm("edit");
                }
                break;

            case "c":
                event.preventDefault();
                openForm("cycle");
                break;

            case "u":
                event.preventDefault();
                undo();
                break;

            case "x":
                event.preventDefault();
                pendingReset = true;
                break;
        }
    }

    function setLabel(set) {
        return `${set.weight} × ${set.reps}${set.amrap ? "+" : ""}`;
    }
</script>

<div class="page-placeholder workouts-page">
    <span class="eyebrow">FOCUSUP / HEALTH</span>

    <div class="tasks-heading-row">
        <h1>Workouts</h1>
        {#if cycle}
            <span class="todo-header-count"
                >Cycle {cycle.number} • {doneCount}/{WORKOUTS_PER_CYCLE} done</span
            >
        {/if}
    </div>

    {#if loading}
        <p>Loading program…</p>
    {:else if !cycle}
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions (esc cancels the form) -->
        <form
            class="workouts-setup"
            bind:this={setupEl}
            on:submit|preventDefault={submitSetup}
            on:keydown={(e) => e.key === "Escape" && document.activeElement?.blur()}
        >
            <p class="workouts-setup-intro">
                Enter your current one-rep max for each lift. Training maxes are
                90% of it, and every weight in the program is worked out from
                those. Press <kbd>enter</kbd> to start typing, <kbd>esc</kbd> to leave
                the form.
            </p>

            <label class="workouts-field">
                <span>Start date</span>
                <input class="workouts-input" bind:value={setupStart} placeholder="YYYY-MM-DD" />
            </label>

            {#each LIFTS as lift}
                <label class="workouts-field">
                    <span>{LIFT_META[lift].label} 1RM</span>
                    <input
                        class="workouts-input"
                        type="number"
                        min="0"
                        step="5"
                        bind:value={setupMax[lift]}
                        placeholder="lb"
                    />
                    <span class="workouts-field-note"
                        >{setupTMs[lift] ? `TM ${setupTMs[lift]} lb` : ""}</span
                    >
                </label>
            {/each}

            <button type="submit" class="workouts-button">Start program</button>
        </form>
    {:else}
        <div class="tasks-hint">
            {#if navigating && view === "projections"}
                <kbd>h</kbd><kbd>l</kbd> lift
                <kbd>+</kbd><kbd>-</kbd> horizon
                <kbd>d</kbd> by date
                <kbd>g</kbd> goal
                <kbd>p</kbd> program
                <kbd>q</kbd> back
            {:else if navigating}
                <kbd>h</kbd><kbd>j</kbd><kbd>k</kbd><kbd>l</kbd> move
                <kbd>enter</kbd> log / edit
                <kbd>c</kbd> edit cycle
                <kbd>u</kbd> undo
                <kbd>[</kbd><kbd>]</kbd> cycle
                <kbd>p</kbd> projections
                <kbd>x</kbd> reset
                <kbd>q</kbd> back
            {:else}
                <kbd>j</kbd> navigate workouts
            {/if}
        </div>

        {#if pendingReset}
            <p class="workouts-warning">
                Press x again to delete the whole program and start over.
            </p>
        {/if}

        <div class="workouts-cycle-bar">
            <div class="workouts-cycle-title">
                Cycle {cycle.number}
                <span>started {formatDay(parseDateKey(cycle.startDate))}</span>
                {#if !isLatest}<span class="workouts-tag">past cycle</span>{/if}
            </div>

            <div class="workouts-maxes">
                {#each LIFTS as lift}
                    <div class="workouts-max">
                        <span class="workouts-max-label">{LIFT_META[lift].label}</span>
                        <span class="workouts-max-value">TM {cycle.trainingMax[lift]}</span>
                        <span class="workouts-max-note"
                            >{best[lift] ? `best e1RM ${best[lift]}` : "no e1RM yet"}</span
                        >
                    </div>
                {/each}
            </div>
        </div>

        {#if formMode === "cycle"}
            <!-- svelte-ignore a11y_no_noninteractive_element_interactions (esc cancels the form) -->
            <form
                class="workouts-form workouts-cycle-form"
                bind:this={formEl}
                on:submit|preventDefault={submitForm}
                on:keydown={onFormKeydown}
            >
                <label class="workouts-field">
                    <span>Start date</span>
                    <input class="workouts-input" bind:value={cycleStart} placeholder="YYYY-MM-DD" />
                </label>
                {#each LIFTS as lift}
                    <label class="workouts-field">
                        <span>{LIFT_META[lift].label} TM</span>
                        <input class="workouts-input" type="number" min="0" step="5" bind:value={cycleTM[lift]} />
                    </label>
                {/each}
                <button type="submit" class="workouts-button">Save</button>
                <span class="workouts-form-hint">enter save • esc cancel</span>
            </form>
        {/if}

        {#if view === "projections"}
            <WorkoutProjections bind:this={projectionsEl} {cycles} {today} />
        {:else}
            <div class="workouts-layout">
                <div class="workouts-grid">
                    <div class="workouts-grid-row workouts-grid-head">
                        <span></span>
                        {#each LIFTS as lift}
                            <span>{LIFT_META[lift].label}</span>
                        {/each}
                    </div>

                    {#each [1, 2, 3, 4] as week}
                        <div class="workouts-grid-row">
                            <span class="workouts-week-label">
                                Week {week}
                                <small>{WEEKS[week].label}</small>
                            </span>

                            {#each LIFTS as lift, l}
                                {@const index = (week - 1) * 4 + l}
                                {@const status = plan[index]}
                                {@const log = cycle.logs[index]}
                                {@const top = topSet(cycle.trainingMax[lift], week)}
                                <div
                                    class="workouts-cell"
                                    class:done={status?.status === "done"}
                                    class:next={status?.status === "next"}
                                    class:cursor={navigating && index === cursor}
                                >
                                    <span class="workouts-cell-top">{setLabel(top)}</span>
                                    {#if log}
                                        <span class="workouts-cell-date"
                                            >✓ {formatDay(status.date)}</span
                                        >
                                        {#if workoutEstimate(cycle, log)}
                                            <span class="workouts-cell-note"
                                                >{log.amrapReps} reps → {workoutEstimate(cycle, log)}</span
                                            >
                                        {/if}
                                    {:else if status}
                                        <span class="workouts-cell-date"
                                            >{formatDay(status.due)}</span
                                        >
                                        {#if status.status === "next"}
                                            <span class="workouts-cell-note">next up</span>
                                        {/if}
                                    {/if}
                                </div>
                            {/each}
                        </div>
                    {/each}
                </div>

                <section class="workouts-detail">
                    <span class="eyebrow">
                        {canLog ? "NEXT UP" : selectedLog ? "DONE" : "PLANNED"}
                    </span>
                    <h2>
                        {LIFT_META[selected.lift].label}
                        <small>Week {selected.week} · {WEEKS[selected.week].label}</small>
                    </h2>

                    <div class="workouts-detail-status">
                        {#if selectedLog}
                            Done {formatDay(selectedStatus.date)}
                            {#if workoutEstimate(cycle, selectedLog)}
                                · {selectedLog.amrapReps} reps on the last set · e1RM
                                {workoutEstimate(cycle, selectedLog)} lb
                            {/if}
                        {:else if selectedStatus}
                            Due {formatDay(selectedStatus.due)} ({dueLabel(selectedStatus.due, today)})
                        {/if}
                    </div>

                    {#if selectedStatus?.lateDays}
                        <p class="workouts-warning">
                            {selectedStatus.lateDays} day{selectedStatus.lateDays === 1 ? "" : "s"}
                            behind plan. The rest of the cycle has shifted to keep your
                            rest days.
                        </p>
                    {/if}

                    <table class="workouts-sets">
                        <tbody>
                            {#each selectedSets as set, i}
                                {#if set.kind !== "volume" || selectedSets[i - 1]?.kind !== "volume"}
                                    <tr class={set.kind} class:amrap={set.amrap}>
                                        <td class="workouts-set-kind">
                                            {set.kind === "warmup"
                                                ? "Warm-up"
                                                : set.kind === "main"
                                                  ? "Work"
                                                  : "5 × 10"}
                                        </td>
                                        <td class="workouts-set-pct">{set.pct}%</td>
                                        <td class="workouts-set-weight">{set.weight} lb</td>
                                        <td class="workouts-set-reps">
                                            {set.kind === "volume"
                                                ? "5 sets × 10"
                                                : `× ${set.reps}${set.amrap ? "+" : ""}`}
                                        </td>
                                    </tr>
                                {/if}
                            {/each}
                        </tbody>
                    </table>

                    {#if formMode === "log" || formMode === "edit"}
                        <!-- svelte-ignore a11y_no_noninteractive_element_interactions (esc cancels the form) -->
                        <form
                            class="workouts-form"
                            bind:this={formEl}
                            on:submit|preventDefault={submitForm}
                            on:keydown={onFormKeydown}
                        >
                            <label class="workouts-field">
                                <span>Date</span>
                                <input class="workouts-input" bind:value={formDate} placeholder="YYYY-MM-DD" />
                            </label>
                            {#if !isDeload(cursor)}
                                <label class="workouts-field">
                                    <span>Reps on last set</span>
                                    <input
                                        class="workouts-input"
                                        type="number"
                                        min="0"
                                        data-autofocus
                                        bind:value={formReps}
                                    />
                                    <span class="workouts-field-note"
                                        >{liveEstimate ? `e1RM ${liveEstimate} lb` : ""}</span
                                    >
                                </label>
                            {/if}
                            <button type="submit" class="workouts-button"
                                >{formMode === "log" ? "Log workout" : "Save"}</button
                            >
                            <span class="workouts-form-hint">enter save • esc cancel</span>
                        </form>
                    {:else if canLog}
                        <p class="workouts-detail-hint">Press <kbd>enter</kbd> to log this workout.</p>
                    {:else if selectedLog}
                        <p class="workouts-detail-hint">Press <kbd>e</kbd> to edit this workout.</p>
                    {/if}
                </section>
            </div>
        {/if}
    {/if}

    {#if error}
        <p class="todo-error">{error}</p>
    {/if}
</div>

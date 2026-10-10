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
- enter logs the next workout (or edits a done one). Forms work like
  every form in the app (KeyForm.svelte): j/k fields, enter edit, q/esc
  back; the log form starts out typing the reps
- e edit a done workout, c edit the cycle, u undo the last log
- [ ] previous/next cycle, x x reset the whole program
- p toggles the projections view (WorkoutProjections.svelte), which takes
  h/l, +/-, d and g while it's showing; k, q or esc still go back to the tabs

*/
    import { onMount, onDestroy } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import { startOfDay } from "../calendarGrid.js";
    import WorkoutProjections from "./WorkoutProjections.svelte";
    import KeyForm from "./KeyForm.svelte";
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
    let formRef;
    let formValues = { date: todayKey, reps: "" };
    // { start, <lift>: training max, ... }
    let cycleValues = {};

    // { start, <lift>: one-rep max, ... }
    let setupValues = { start: todayKey, squat: "", bench: "", deadlift: "", press: "" };
    let setupRef;

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
            ? estimated1RM(topSet(cycle.trainingMax[selected.lift], selected.week).weight, Number(formValues.reps))
            : null;

    // ---- setup ----
    $: setupTMs = Object.fromEntries(
        LIFTS.map((lift) => [lift, Number(setupValues[lift]) > 0 ? trainingMaxFrom1RM(Number(setupValues[lift])) : null]),
    );

    $: setupRows = [
        { key: "start", type: "text", label: "Start date", placeholder: "YYYY-MM-DD" },
        ...LIFTS.map((lift) => ({
            key: lift,
            type: "number",
            label: `${LIFT_META[lift].label} 1RM`,
            placeholder: "lb",
            note: setupTMs[lift] ? `TM ${setupTMs[lift]} lb` : "",
        })),
    ];

    // Form submit handlers throw on invalid input; KeyForm shows the
    // message under the form and keeps it open.
    async function submitSetup(values) {
        if (!isValidDateKey(values.start)) {
            throw new Error("Start date must be YYYY-MM-DD");
        }
        if (LIFTS.some((lift) => !(Number(values[lift]) > 0))) {
            throw new Error("Enter a 1RM for every lift");
        }

        const oneRepMax = Object.fromEntries(LIFTS.map((lift) => [lift, Number(values[lift])]));
        setCycles([await SetupWorkouts(values.start, oneRepMax)]);
        error = null;
    }

    // ---- workout / cycle forms ----
    $: logRows = [
        { key: "date", type: "text", label: "Date", placeholder: "YYYY-MM-DD" },
        ...(cycle && isDeload(cursor)
            ? []
            : [
                  {
                      key: "reps",
                      type: "number",
                      label: "Reps on last set",
                      note: liveEstimate ? `e1RM ${liveEstimate} lb` : "",
                  },
              ]),
    ];

    $: cycleRows = [
        { key: "start", type: "text", label: "Start date", placeholder: "YYYY-MM-DD" },
        ...LIFTS.map((lift) => ({
            key: lift,
            type: "number",
            label: `${LIFT_META[lift].label} TM`,
            suffix: "lb",
        })),
    ];

    function openForm(mode) {
        formMode = mode;
        error = null;

        if (mode === "log") {
            formValues = { date: todayKey, reps: "" };
        } else if (mode === "edit") {
            formValues = {
                date: selectedLog.date,
                reps: isDeload(cursor) ? "" : String(selectedLog.amrapReps),
            };
        } else if (mode === "cycle") {
            cycleValues = {
                start: cycle.startDate,
                ...Object.fromEntries(LIFTS.map((lift) => [lift, String(cycle.trainingMax[lift])])),
            };
        }
    }

    function closeForm() {
        formMode = null;
        document.activeElement?.blur();
    }

    async function submitForm(values) {
        if (!isValidDateKey(values.date)) {
            throw new Error("Date must be YYYY-MM-DD");
        }
        const reps = isDeload(cursor) ? 0 : Number(values.reps || 0);
        if (!Number.isInteger(reps) || reps < 0) {
            throw new Error("Reps must be a whole number");
        }

        if (formMode === "log") {
            setCycles(await CompleteWorkout(cycle.id, cursor, values.date, reps));
        } else {
            replaceCycle(await UpdateWorkoutLog(cycle.id, cursor, values.date, reps));
        }
        error = null;
        closeForm();
    }

    async function submitCycle(values) {
        if (!isValidDateKey(values.start)) {
            throw new Error("Start date must be YYYY-MM-DD");
        }
        const tm = Object.fromEntries(LIFTS.map((lift) => [lift, Number(values[lift])]));
        if (LIFTS.some((lift) => !(tm[lift] > 0))) {
            throw new Error("Every training max must be greater than 0");
        }

        replaceCycle(await UpdateWorkoutCycle(cycle.id, values.start, tm));
        error = null;
        closeForm();
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
            formRef?.handleKey(event);
            return;
        }

        // Before setup, the setup form takes the keyboard once the page has
        // it (enter or j from the tab bar); q/esc in the form hand it back.
        if (!cycle) {
            if (loading) {
                return;
            }
            if (navigating) {
                setupRef?.handleKey(event);
            } else if (event.key === "Enter") {
                event.preventDefault();
                mode.set("grid");
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
        <div class="workouts-setup">
            <p class="workouts-setup-intro">
                Enter your current one-rep max for each lift. Training maxes are
                90% of it, and every weight in the program is worked out from
                those. Press <kbd>enter</kbd> to start, <kbd>j</kbd>/<kbd>k</kbd> to
                move between fields, <kbd>enter</kbd> to type a value and
                <kbd>q</kbd> to leave the form.
            </p>

            <KeyForm
                bind:this={setupRef}
                bind:values={setupValues}
                rows={setupRows}
                submitLabel="Start program"
                active={navigating}
                onSubmit={submitSetup}
                onCancel={() => mode.set("tabs")}
            />
        </div>
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
            <div class="workouts-form workouts-cycle-form">
                <KeyForm
                    bind:this={formRef}
                    bind:values={cycleValues}
                    rows={cycleRows}
                    onSubmit={submitCycle}
                    onCancel={closeForm}
                />
            </div>
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
                        <div class="workouts-form">
                            <KeyForm
                                bind:this={formRef}
                                bind:values={formValues}
                                rows={logRows}
                                submitLabel={formMode === "log" ? "Log workout" : "Save"}
                                startEditing={isDeload(cursor) ? "date" : "reps"}
                                onSubmit={submitForm}
                                onCancel={closeForm}
                            />
                        </div>
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

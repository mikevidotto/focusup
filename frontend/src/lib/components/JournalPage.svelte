<script>
    /*

Journal page: one entry per day.

Each day shows a few prompts (picked deterministically from the pool in
journalPrompts.js) with an answer box under each, plus a free-write area.
An entry stores the prompt text it was answered against, so past days keep
their own prompts. Edits auto-save shortly after you stop typing; a day with
nothing written in it isn't stored at all.

Keyboard:
- browse mode (keyboard mode "tabs", so h/l still switch tabs):
  [ ] previous/next day, g today, enter (or j) drills into the page,
  s swap the first unanswered prompt, x x delete the day's entry
- drilled in (keyboard mode "grid"):
  j/k move between the prompts and free write, enter/i start writing,
  s swap the prompt under the cursor (only while it's unanswered),
  [ ] day, q/esc back
- while writing: esc stops writing and returns to the cursor

*/
    import { onMount, onDestroy, tick } from "svelte";
    import { activeWidgetKeyHandler, mode } from "../stores/keyboard.js";
    import { toDateKey } from "../habitDisplay.js";
    import { startOfDay } from "../calendarGrid.js";
    import {
        promptsForDate,
        swapPrompt,
        wordCount,
        currentStreak,
        entryPreview,
        formatEntryDate,
        fromDateKey,
    } from "../journalPrompts.js";
    import {
        ListJournalEntries,
        SaveJournalEntry,
        DeleteJournalEntry,
    } from "../../../wailsjs/go/main/App.js";

    const SAVE_DELAY_MS = 800;
    const DAY_MS = 86400000;

    const today = startOfDay(new Date());
    const todayKey = toDateKey(today);

    let entries = [];
    let viewedKey = todayKey;
    let draft = blankDraft(todayKey);
    // 0..prompts.length-1 are the prompts; prompts.length is the free write.
    let cursor = 0;
    let fieldEls = [];
    let dirty = false;
    let saving = false;
    let saveTimer = null;
    let pendingDelete = false;
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            entries = await ListJournalEntries();
            loadDraft(viewedKey);
        } catch (e) {
            error = String(e);
        } finally {
            loading = false;
        }

        activeWidgetKeyHandler.set(handleKey);
    });

    onDestroy(() => {
        activeWidgetKeyHandler.set(null);
        saveNow();
    });

    // Like the Habits page, "drilled in" is the shared keyboard mode, so
    // anything that resets the mode also pops us back out.
    $: drilled = $mode === "grid";

    $: viewedDate = fromDateKey(viewedKey);
    $: daysAgo = Math.round((today - viewedDate) / DAY_MS);
    $: savedEntry = entries.find((e) => e.date === viewedKey);
    $: words = wordCount(draft);
    $: streak = currentStreak(entries, today);
    $: fieldCount = draft.prompts.length + 1;

    function blankDraft(key) {
        return {
            prompts: promptsForDate(key).map((prompt) => ({
                prompt,
                answer: "",
            })),
            body: "",
        };
    }

    function loadDraft(key) {
        const entry = entries.find((e) => e.date === key);

        draft = entry
            ? {
                  prompts: entry.prompts.map((p) => ({ ...p })),
                  body: entry.body,
              }
            : blankDraft(key);
        dirty = false;
        pendingDelete = false;
        cursor = Math.min(cursor, draft.prompts.length);
    }

    function markDirty() {
        dirty = true;
        clearTimeout(saveTimer);
        saveTimer = setTimeout(saveNow, SAVE_DELAY_MS);
    }

    async function saveNow() {
        clearTimeout(saveTimer);
        saveTimer = null;

        if (!dirty) {
            return;
        }

        // Snapshot what's being saved: the user may keep typing, or move to
        // another day, while the request is in flight.
        const key = viewedKey;
        const prompts = draft.prompts.map((p) => ({ ...p }));
        const body = draft.body;

        dirty = false;
        saving = true;

        try {
            const saved = await SaveJournalEntry(key, prompts, body);
            const others = entries.filter((e) => e.date !== key);

            entries =
                wordCount(saved) === 0
                    ? others
                    : [...others, saved].sort((a, b) =>
                          b.date.localeCompare(a.date),
                      );
            error = null;
        } catch (e) {
            error = String(e);
        } finally {
            saving = false;
        }
    }

    async function goToDay(key) {
        if (key > todayKey || key === viewedKey) {
            return;
        }

        await saveNow();
        viewedKey = key;
        loadDraft(key);
    }

    function shiftDay(delta) {
        goToDay(
            toDateKey(
                new Date(
                    viewedDate.getFullYear(),
                    viewedDate.getMonth(),
                    viewedDate.getDate() + delta,
                ),
            ),
        );
    }

    async function deleteEntry() {
        clearTimeout(saveTimer);
        dirty = false;

        try {
            await DeleteJournalEntry(viewedKey);
            entries = entries.filter((e) => e.date !== viewedKey);
            loadDraft(viewedKey);
        } catch (e) {
            error = String(e);
        }
    }

    // Swaps a prompt for another from the pool, but never one that already
    // has an answer — that would silently re-label what was written.
    function swap(index) {
        const target = draft.prompts[index];

        if (!target || target.answer.trim()) {
            return;
        }

        const current = draft.prompts.map((p) => p.prompt);
        target.prompt = swapPrompt(current, index);
        draft = draft;

        // Only worth persisting once the day has something written in it;
        // a blank day just shows the swap until it's left.
        if (savedEntry) {
            markDirty();
        }
    }

    async function startWriting(index) {
        cursor = index;
        mode.set("grid");
        await tick();
        fieldEls[index]?.focus();
    }

    function onFieldFocus(index) {
        cursor = index;
        mode.set("grid");
    }

    function onFieldKeydown(event) {
        if (event.key === "Escape") {
            event.preventDefault();
            event.target.blur();
        }
    }

    // Grows a textarea to fit its content; re-measured whenever the bound
    // value changes (including switching days).
    function autosize(node, _value) {
        const resize = () => {
            node.style.height = "auto";
            node.style.height = `${node.scrollHeight}px`;
        };

        resize();

        return { update: resize };
    }

    function handleKey(event) {
        // "x" arms a delete; the very next key either confirms it (x) or
        // cancels it (anything else, which is then handled normally).
        if (pendingDelete) {
            pendingDelete = false;

            if (event.key === "x") {
                event.preventDefault();
                deleteEntry();
                return;
            }
        }

        if (event.key === "[" || event.key === "]") {
            event.preventDefault();
            shiftDay(event.key === "[" ? -1 : 1);
            return;
        }

        if (event.key === "g") {
            event.preventDefault();
            goToDay(todayKey);
            return;
        }

        if (drilled) {
            handleDrilledKey(event);
        } else {
            handleBrowseKey(event);
        }
    }

    function handleBrowseKey(event) {
        switch (event.key) {
            case "Enter":
                event.preventDefault();
                mode.set("grid");
                break;

            case "s": {
                event.preventDefault();
                const index = draft.prompts.findIndex((p) => !p.answer.trim());
                swap(index);
                break;
            }

            case "x":
                if (savedEntry) {
                    event.preventDefault();
                    pendingDelete = true;
                }
                break;
        }
    }

    function handleDrilledKey(event) {
        switch (event.key) {
            case "j":
                event.preventDefault();
                cursor = Math.min(cursor + 1, fieldCount - 1);
                break;

            case "k":
                event.preventDefault();
                cursor = Math.max(cursor - 1, 0);
                break;

            case "Enter":
            case "i":
                event.preventDefault();
                startWriting(cursor);
                break;

            case "s":
                event.preventDefault();
                swap(cursor);
                break;

            case "q":
            case "Escape":
                event.preventDefault();
                mode.set("tabs");
                break;
        }
    }
</script>

<div class="page-placeholder journal-page">
    <span class="eyebrow">FOCUSUP / JOURNAL</span>

    <div class="tasks-heading-row">
        <h1>Journal</h1>
        <span class="todo-header-count">
            {streak}-day streak · {entries.length}
            {entries.length === 1 ? "entry" : "entries"}
        </span>
    </div>

    <div class="tasks-hint">
        {#if drilled}
            <kbd>j</kbd><kbd>k</kbd> move
            <kbd>enter</kbd> write
            <kbd>esc</kbd> stop writing
            <kbd>s</kbd> swap prompt
            <kbd>[</kbd><kbd>]</kbd> day
            <kbd>q</kbd> back
        {:else}
            <kbd>enter</kbd> open
            <kbd>[</kbd><kbd>]</kbd> day
            <kbd>g</kbd> today
            <kbd>s</kbd> swap prompt
            <kbd>x</kbd> delete
        {/if}
    </div>

    <div class="journal-layout">
        <div class="journal-main">
            <div class="habits-week-label">
                {formatEntryDate(viewedDate)}
                <span class="habits-week-tag">
                    {#if daysAgo === 0}
                        today
                    {:else if daysAgo === 1}
                        yesterday
                    {:else}
                        {daysAgo} days ago
                    {/if}
                </span>
                {#if pendingDelete}
                    <span class="habits-delete-confirm">x again to delete</span>
                {/if}
            </div>

            {#if loading}
                <p>Loading journal…</p>
            {:else}
                {#each draft.prompts as item, index (index)}
                    <label
                        class="journal-field"
                        class:cursor={drilled && cursor === index}
                    >
                        <span class="journal-prompt">
                            <span class="journal-prompt-number">{index + 1}</span>
                            {item.prompt}
                        </span>
                        <textarea
                            rows="2"
                            placeholder="write a few words…"
                            bind:this={fieldEls[index]}
                            bind:value={item.answer}
                            use:autosize={item.answer}
                            on:input={markDirty}
                            on:focus={() => onFieldFocus(index)}
                            on:blur={saveNow}
                            on:keydown={onFieldKeydown}
                        ></textarea>
                    </label>
                {/each}

                <label
                    class="journal-field journal-freewrite"
                    class:cursor={drilled && cursor === draft.prompts.length}
                >
                    <span class="journal-prompt">Free write</span>
                    <textarea
                        rows="6"
                        placeholder="anything else on your mind…"
                        bind:this={fieldEls[draft.prompts.length]}
                        bind:value={draft.body}
                        use:autosize={draft.body}
                        on:input={markDirty}
                        on:focus={() => onFieldFocus(draft.prompts.length)}
                        on:blur={saveNow}
                        on:keydown={onFieldKeydown}
                    ></textarea>
                </label>

                <div class="journal-status">
                    <span>{words} {words === 1 ? "word" : "words"}</span>
                    <span>
                        {#if saving || dirty}
                            saving…
                        {:else if savedEntry}
                            saved
                        {:else}
                            not started
                        {/if}
                    </span>
                </div>
            {/if}

            {#if error}
                <p class="todo-error">{error}</p>
            {/if}
        </div>

        <aside class="journal-history">
            <span class="eyebrow">PAST ENTRIES</span>

            {#if entries.length === 0}
                <div class="empty-widget">
                    <span>空</span>
                    <p>Nothing written yet — today's a good day to start</p>
                </div>
            {:else}
                <ul class="journal-history-list">
                    {#each entries as entry (entry.date)}
                        <li>
                            <button
                                type="button"
                                class="journal-history-item"
                                class:active={entry.date === viewedKey}
                                on:click={() => goToDay(entry.date)}
                            >
                                <span class="journal-history-head">
                                    <span>
                                        {fromDateKey(
                                            entry.date,
                                        ).toLocaleDateString("en-US", {
                                            weekday: "short",
                                            month: "short",
                                            day: "numeric",
                                        })}
                                    </span>
                                    <span class="journal-history-words"
                                        >{wordCount(entry)}w</span
                                    >
                                </span>
                                <span class="journal-history-preview"
                                    >{entryPreview(entry)}</span
                                >
                            </button>
                        </li>
                    {/each}
                </ul>
            {/if}
        </aside>
    </div>
</div>

<script>
    /*

A keyboard-driven form, shared by every form in the app (the Calendar
page's event form, the Workouts forms) so they all behave the same way:

- j/k move between fields; the last one is always the submit row
- enter activates the field under the cursor:
  - text/number/textarea: type into it; enter confirms (shift+enter is a
    newline in a textarea), esc reverts what you typed. Either goes back
    to moving between fields
  - select: j/k pick an option, enter confirms, q/esc back out unchanged
  - multi (e.g. weekdays): j/k or h/l move, enter/space toggle, q/esc back
  - toggle: flips in place
  - submit row: submits
- confirming a text field or select moves the cursor to the next field
- q/esc while moving between fields cancels the form

The page owns the keyboard (activeWidgetKeyHandler) and forwards keys to
handleKey below, so they route through App like everywhere else. While a
text field is focused its own keydown handler takes over; App ignores keys
typed into inputs.

Rows look like { key, type, label, options?, placeholder?, suffix?, note? },
where key indexes into `values` and options are { value, label } (for
multi, values[key] is an array of booleans, one per option).

*/
    import { onMount, tick } from "svelte";

    export let rows;
    export let values;
    export let submitLabel = "Save";
    export let onSubmit;
    export let onCancel;
    // Row key to start typing into as soon as the form opens.
    export let startEditing = null;
    // False draws no cursor (the page hasn't handed the form the keyboard).
    export let active = true;

    const SUBMIT = "__submit";

    let cursorKey = startEditing;
    let activeKey = null; // text field being typed into
    let picker = null; // { key, index } for an open select or multi
    let snapshot;
    let inputs = {};
    let error = null;
    let saving = false;

    $: items = [...rows, { key: SUBMIT, type: "submit" }];
    $: if (!items.some((row) => row.key === cursorKey)) {
        cursorKey = items[0].key;
    }
    $: cursorIndex = items.findIndex((row) => row.key === cursorKey);

    onMount(() => {
        if (startEditing) {
            activate(items.find((row) => row.key === startEditing));
        }
    });

    const isTextual = (row) => ["text", "number", "textarea"].includes(row.type);

    function move(delta) {
        const next = Math.min(Math.max(cursorIndex + delta, 0), items.length - 1);
        cursorKey = items[next].key;
    }

    // Moves the cursor past `key`, after any rows a confirmed choice added
    // or removed (e.g. picking "Weekly" adds the weekday row) have rendered.
    async function advanceFrom(key) {
        await tick();
        const index = items.findIndex((row) => row.key === key);
        cursorKey = items[Math.min(index + 1, items.length - 1)].key;
    }

    async function activate(row) {
        if (!row) {
            return;
        }

        cursorKey = row.key;

        if (row.type === "submit") {
            submit();
        } else if (row.type === "toggle") {
            values[row.key] = !values[row.key];
        } else if (row.type === "select") {
            const index = row.options.findIndex((o) => o.value === values[row.key]);
            picker = { key: row.key, index: Math.max(index, 0) };
        } else if (row.type === "multi") {
            picker = { key: row.key, index: 0 };
        } else if (isTextual(row)) {
            activeKey = row.key;
            snapshot = values[row.key];
            await tick();
            inputs[row.key]?.focus();
            inputs[row.key]?.select();
        }
    }

    async function submit() {
        if (saving) {
            return;
        }

        error = null;
        saving = true;
        try {
            await onSubmit(values);
        } catch (e) {
            error = e instanceof Error ? e.message : String(e);
        } finally {
            saving = false;
        }
    }

    function handlePickerKey(event) {
        const row = items.find((r) => r.key === picker.key);
        const last = row.options.length - 1;
        const multi = row.type === "multi";
        // A multi's chips sit in a row, so h/l move along it too.
        const nextKeys = multi ? ["j", "ArrowDown", "l", "ArrowRight"] : ["j", "ArrowDown"];
        const prevKeys = multi ? ["k", "ArrowUp", "h", "ArrowLeft"] : ["k", "ArrowUp"];
        const key = event.key;

        if (nextKeys.includes(key)) {
            picker.index = Math.min(picker.index + 1, last);
        } else if (prevKeys.includes(key)) {
            picker.index = Math.max(picker.index - 1, 0);
        } else if (multi && (key === "Enter" || key === " ")) {
            values[row.key][picker.index] = !values[row.key][picker.index];
        } else if (key === "Enter") {
            values[row.key] = row.options[picker.index].value;
            picker = null;
            advanceFrom(row.key);
        } else if (key === "q" || key === "Escape") {
            picker = null;
        } else {
            return false;
        }

        event.preventDefault();
        return true;
    }

    // Called by the page for every key while the form is open. Returns
    // whether the form used it.
    export function handleKey(event) {
        if (activeKey) {
            return false;
        }
        if (picker) {
            return handlePickerKey(event);
        }

        switch (event.key) {
            case "j":
            case "ArrowDown":
                move(1);
                break;
            case "k":
            case "ArrowUp":
                move(-1);
                break;
            case "Enter":
                activate(items[cursorIndex]);
                break;
            case "q":
            case "Escape":
                onCancel();
                break;
            case "Tab":
                break;
            default:
                return false;
        }

        event.preventDefault();
        return true;
    }

    function onInputKeydown(event, row) {
        if (event.key === "Enter" && !(row.type === "textarea" && event.shiftKey)) {
            event.preventDefault();
            activeKey = null;
            event.target.blur();
            advanceFrom(row.key);
        } else if (event.key === "Escape") {
            event.preventDefault();
            values[row.key] = snapshot;
            activeKey = null;
            event.target.blur();
        } else if (event.key === "Tab") {
            event.preventDefault();
        }
    }

    // A mouse click into a field enters it the same way enter would.
    function onInputFocus(row) {
        if (activeKey !== row.key) {
            cursorKey = row.key;
            activeKey = row.key;
            snapshot = values[row.key];
        }
    }

    function onInputBlur(row) {
        if (activeKey === row.key) {
            activeKey = null;
        }
    }

    function optionLabel(row) {
        return row.options.find((o) => o.value === values[row.key])?.label ?? "";
    }

    $: pickerRow = picker && items.find((row) => row.key === picker.key);
</script>

<div class="key-form">
    <ul class="key-form-list">
        {#each items as row (row.key)}
            <li
                class="key-form-item"
                class:cursor={active && row.key === cursorKey}
                class:editing={row.key === activeKey || row.key === picker?.key}
            >
                {#if row.type === "submit"}
                    <span class="key-form-submit">{saving ? "Saving…" : submitLabel}</span>
                {:else}
                    <span class="key-form-label">{row.label}</span>
                    <div class="key-form-value">
                        {#if row.type === "textarea"}
                            <textarea
                                class="workouts-input key-form-textarea"
                                rows="3"
                                tabindex="-1"
                                placeholder={row.placeholder}
                                bind:this={inputs[row.key]}
                                bind:value={values[row.key]}
                                on:keydown={(e) => onInputKeydown(e, row)}
                                on:focus={() => onInputFocus(row)}
                                on:blur={() => onInputBlur(row)}
                            ></textarea>
                        {:else if isTextual(row)}
                            <input
                                class="workouts-input"
                                class:key-form-number={row.type === "number"}
                                inputmode={row.type === "number" ? "decimal" : "text"}
                                tabindex="-1"
                                placeholder={row.placeholder}
                                bind:this={inputs[row.key]}
                                bind:value={values[row.key]}
                                on:keydown={(e) => onInputKeydown(e, row)}
                                on:focus={() => onInputFocus(row)}
                                on:blur={() => onInputBlur(row)}
                            />
                        {:else if row.type === "select"}
                            <span class="key-form-choice">{optionLabel(row)} ▾</span>
                        {:else if row.type === "multi"}
                            <span class="key-form-chips">
                                {#each row.options as option, i}
                                    <span
                                        class="key-form-chip"
                                        class:on={values[row.key][i]}
                                        class:cursor={picker?.key === row.key && picker.index === i}
                                        >{option.label}</span
                                    >
                                {/each}
                            </span>
                        {:else if row.type === "toggle"}
                            <span class="key-form-choice">{values[row.key] ? "[x] yes" : "[ ] no"}</span>
                        {/if}

                        {#if row.suffix}
                            <span class="key-form-suffix">{row.suffix}</span>
                        {/if}
                        {#if row.note}
                            <span class="key-form-note">{row.note}</span>
                        {/if}
                    </div>

                    {#if row.type === "select" && picker?.key === row.key}
                        <ul class="key-form-options">
                            {#each row.options as option, i}
                                <li class="key-form-option" class:selected={picker.index === i}>
                                    {option.label}
                                </li>
                            {/each}
                        </ul>
                    {/if}
                {/if}
            </li>
        {/each}
    </ul>

    {#if error}
        <p class="todo-error">{error}</p>
    {/if}

    {#if active}
        <span class="key-form-hint">
            {#if activeKey}
                enter confirm • esc back{items[cursorIndex]?.type === "textarea"
                    ? " • shift+enter new line"
                    : ""}
            {:else if pickerRow?.type === "multi"}
                j/k move • enter toggle • q done
            {:else if pickerRow}
                j/k choose • enter select • q back
            {:else}
                j/k move • enter edit • q cancel
            {/if}
        </span>
    {/if}
</div>

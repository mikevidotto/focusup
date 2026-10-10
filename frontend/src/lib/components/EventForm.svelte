<script>
    /*

Create/edit form for a calendar event, shown in the Calendar page's right
panel. The keyboard behavior is KeyForm's (j/k fields, enter edit, q/esc
back); this only decides which fields show for the current choices.

Description, location and important only show when editing an existing
event, so creating stays quick.

*/
    import KeyForm from "./KeyForm.svelte";
    import { WEEKDAY_LABELS } from "../calendarGrid.js";
    import {
        END_MODES,
        REMINDER_OPTIONS,
        REPEAT_OPTIONS,
        REPEAT_UNITS,
        TIME_MODES,
        formToInput,
    } from "../eventForm.js";

    // Initial field values (see eventForm.js emptyForm / eventToForm). The
    // form edits its own copy; the parent remounts it to start over.
    export let initial;
    export let editing = false;
    export let onSave;
    export let onCancel;

    let form = structuredClone(initial);
    let keyForm;

    const WEEKDAY_OPTIONS = WEEKDAY_LABELS.map((label, i) => ({ value: i, label: label.slice(0, 2) }));

    export function handleKey(event) {
        return keyForm?.handleKey(event);
    }

    function buildRows(form, editing) {
        const rows = [
            { key: "title", type: "text", label: "Title", placeholder: "What's happening?" },
            { key: "date", type: "text", label: "Date", placeholder: "YYYY-MM-DD" },
            { key: "timeMode", type: "select", label: "Time", options: TIME_MODES },
        ];

        if (form.timeMode !== "allday") {
            const label = form.timeMode === "range" ? "Start" : "At";
            rows.push({ key: "startTime", type: "text", label, placeholder: "9am" });
        }
        if (form.timeMode === "range") {
            rows.push({ key: "endTime", type: "text", label: "End", placeholder: "10am" });
        }

        rows.push({ key: "repeat", type: "select", label: "Repeat", options: REPEAT_OPTIONS });

        if (form.repeat !== "none") {
            const unit = REPEAT_UNITS[form.repeat];
            rows.push({
                key: "interval",
                type: "number",
                label: "Every",
                suffix: Number(form.interval) === 1 ? unit : `${unit}s`,
            });

            if (form.repeat === "weekly") {
                rows.push({ key: "weekdays", type: "multi", label: "On", options: WEEKDAY_OPTIONS });
            }

            rows.push({ key: "endMode", type: "select", label: "Ends", options: END_MODES });

            if (form.endMode === "count") {
                rows.push({ key: "count", type: "number", label: "After", suffix: "times" });
            } else if (form.endMode === "until") {
                rows.push({ key: "untilDate", type: "text", label: "End date", placeholder: "YYYY-MM-DD" });
            }
        }

        rows.push({ key: "reminder", type: "select", label: "Reminder", options: REMINDER_OPTIONS });

        if (editing) {
            rows.push(
                { key: "location", type: "text", label: "Location" },
                { key: "description", type: "textarea", label: "Notes" },
                { key: "important", type: "toggle", label: "Important" },
            );
        }

        return rows;
    }

    $: rows = buildRows(form, editing);
</script>

<div class="event-form">
    <div class="event-form-heading">{editing ? "Edit event" : "New event"}</div>

    <KeyForm
        bind:this={keyForm}
        bind:values={form}
        {rows}
        submitLabel={editing ? "Save" : "Add event"}
        startEditing={editing ? null : "title"}
        onSubmit={(values) => onSave(formToInput(values))}
        {onCancel}
    />
</div>

# CLAUDE.md

Context for Claude Code sessions working on FocusUp. Keep this current — update it whenever a pattern, convention, or pending decision changes, since it's the first thing read each session.

## Project Overview

FocusUp is a personal desktop dashboard app for staying focused and organized, built as a **keyboard-driven** UI (vim-style h/j/k/l navigation, minimal mouse use) with a grid of independent "widgets" (To Do, Calendar, Habits, etc.) plus dedicated full pages for some features (e.g. Tasks has both a dashboard widget and its own page).

Stack:
- **Backend:** Go 1.25, exposed to the frontend via Wails v2 bindings (`app.go`)
- **Frontend:** Svelte 3 + Vite
- **Desktop shell:** Wails v2.16 (frameless window, native binary)
- **Persistence:** flat JSON files in the OS user-config directory (no database yet)

Currently being built: a **Calendar** widget/feature — recurring event scheduling (e.g. "garbage day every other Thursday"), configurable lead-time reminders (e.g. notify 1 day before), in-app popup notifications, plus a calendar widget for the dashboard grid.

## Project Structure

```
.
├── main.go                        # Wails entrypoint, window config, asset embedding
├── app.go                         # App struct — ALL Go methods exposed to frontend live here (thin wrappers only)
├── wails.json                     # Wails project config
├── internal/
│   ├── app/
│   │   └── info.go                # InfoService — trivial example of the internal/<feature> pattern
│   ├── tasks/                     # Reference implementation for any new backend feature
│   │   ├── task.go                #   struct + constants + pure helpers (e.g. normalizePriority)
│   │   ├── service.go             #   business logic, mutex-guarded in-memory state, calls store
│   │   └── store.go                #   load/save to a JSON file in os.UserConfigDir()
│   └── calendar/                  # Calendar feature — see "Calendar file split" below
│       ├── event.go               #   Event, Exception, Occurrence, OccurrenceView types
│       ├── recurrence.go          #   RecurrenceRule + Occurrences() expansion (pure, tested)
│       ├── reminder.go            #   Reminder, DueReminder + lead-time due calculation (pure, tested)
│       ├── service.go             #   Service struct, mutex + JSON store, wraps the pure functions above
│       └── store.go                #   load/save to calendar.json in os.UserConfigDir()
├── frontend/
│   ├── src/
│   │   ├── App.svelte             # Root: keyboard state machine, tab routing, layout shell
│   │   ├── lib/
│   │   │   ├── navigation.js       # `tabs` array — top-level tab bar definition
│   │   │   ├── widgets.js          # `widgets` array — dashboard grid definition (id/size/row/col/component)
│   │   │   ├── keyboardGrid.js     # Pure spatial nav helpers (moveSelection, firstWidget) over widgets[]
│   │   │   ├── taskDisplay.js      # Presentation helpers for the tasks feature (formatting, grouping)
│   │   │   ├── stores/keyboard.js  # `mode`, `selectedWidgetId`, `activeWidgetKeyHandler` stores — see below
│   │   │   └── components/
│   │   │       ├── Dashboard.svelte    # Renders widgets[] into the grid via WidgetSlot
│   │   │       ├── WidgetSlot.svelte   # Generic widget chrome (header/shortcut/selected state)
│   │   │       ├── TabBar.svelte, MenuBar.svelte
│   │   │       ├── TasksPage.svelte    # Reference: a full page (not just a widget) for a feature
│   │   │       └── widgets/
│   │   │           └── TodoWidget.svelte   # Reference: a dashboard-grid widget implementation
│   │   └── main.js
│   └── wailsjs/go/main/App.js     # AUTO-GENERATED bindings — regenerate via `wails dev`/`wails build`, never hand-edit
```

## Build & Dev Commands

```bash
wails dev                       # live-reload dev mode (backend + frontend together) — use this day-to-day
wails build                     # production build, outputs to build/bin
go build ./...                  # backend-only compile check
go vet ./...                    # backend static checks
gofmt -l .                      # list any files needing formatting (should be empty)
cd frontend && npm run dev      # frontend-only dev server
cd frontend && npm run build    # frontend-only production build (vite build)
```

Frontend bindings in `frontend/wailsjs/go/main/App.js` and `.d.ts` are regenerated automatically by `wails dev`/`wails build` whenever `app.go`'s method signatures change. If you add a new bound method and the frontend can't see it, re-run `wails dev`. (A headless/background session without a display can still regenerate bindings with `wails build -s`, which skips the frontend asset build but still runs bindings generation + a full Go compile.)

## Testing

**No test suite exists yet for the frontend; the backend now has one for Calendar.** The codebase is small enough that adding tests is cheap, and calendar logic (recurrence, date math, timezones) is exactly the kind of code that looks right and silently isn't.

- Go: `go test ./...`. `internal/calendar/recurrence_test.go` and `reminder_test.go` are table-driven tests covering weekly/biweekly/monthly recurrence, month-end dates (leap and non-leap years), `Count`-bounded recurrence, skip/reschedule exceptions (including a reschedule that moves an occurrence across the query-range boundary), and reminder lead-time due/not-due edges. Follow this pattern for any new recurrence/date logic.
- Svelte: no test runner configured yet. If widget logic gets complex enough to warrant it, `vitest` + `@testing-library/svelte` is the natural fit — not set up, don't assume it exists.

## Core Architectural Patterns (read before adding a feature)

**Backend — one package per feature, thin bindings in `app.go`:**
Follow `internal/tasks/` exactly:
- `<feature>.go` — struct definitions, constants, pure helper functions (no side effects)
- `service.go` — a `Service` struct holding state behind a `sync.Mutex`, with methods that mutate state and persist via the store
- `store.go` — `dataFilePath()` (under `os.UserConfigDir()/focusup/`), `load()`, `save()` (atomic write via temp file + rename)
- `app.go` only wires `App.<Method>()` → `a.<feature>.<Method>()`. Never put business logic directly in `app.go`.

**Calendar file split (decided, see `internal/calendar/`):**
- `event.go` — `Event`, `Exception`, `Occurrence`, `OccurrenceView` types (no `<feature>.go`/`task.go`-equivalent single file, since Calendar's domain types split naturally across recurrence/reminder concerns).
- `recurrence.go` — `RecurrenceRule` and the pure `Occurrences(event, rangeStart, rangeEnd)` expansion function.
- `reminder.go` — `Reminder`, `DueReminder`, and the pure `ReminderFireTime`/`IsReminderDue` functions.
- `service.go` / `store.go` — same shape as `internal/tasks/`, persisting to `calendar.json`.
- Bound `app.go` methods: `ListEvents`, `ListCalendarOccurrences` (date-ranged, recurrence-expanded — this is what the widget/month-page should call, never `ListEvents` + frontend-side date math), `AddEvent`, `AddReminder`, `DeleteEvent`, `GetDueReminders`.

**Frontend — the three-mode keyboard state machine:**
`stores/keyboard.js` defines `mode` as one of:
- `"tabs"` — top-level tab bar owns input (h/l cycle tabs, digit keys jump tabs)
- `"dashboard"` — grid navigation across widget slots (h/j/k/l move `selectedWidgetId` via `keyboardGrid.js`)
- `"widget"` — input is locked to whichever widget is focused; that widget owns `activeWidgetKeyHandler`

Any new widget must, like `TodoWidget.svelte`:
- Accept a `focused` prop from `Dashboard.svelte`
- Set `activeWidgetKeyHandler` to its own `handleKey` function only while `focused` is true, and clear it (`set(null)`) otherwise — this is a reactive `$:` block, not a one-time effect
- Never listen to `window` keydown directly — `App.svelte` is the single global listener and dispatches based on `mode`

**Frontend — widget registration:**
New dashboard widgets are added to the `widgets` array in `lib/widgets.js` with `id`, `title`, `shortcut`, `size` (`tall`/`medium`/etc., maps to a CSS grid span), and explicit `row`/`col` (logical grid position used by `keyboardGrid.js` for spatial nav — must match the actual CSS grid placement or navigation will feel wrong). There are currently two open slots (`widget-2`, `widget-3`) with `component: null` — Calendar is filling one of these (see Decisions Made).

**Full pages vs. widgets:**
Some features (Tasks) have both a compact dashboard widget AND a dedicated full page reached via the tab bar (`navigation.js` + routing in `App.svelte`). Decide up front whether Calendar needs both (a small "next 3 events" widget plus a full month-view page) or just one.

## Coding Conventions

- Go: `gofmt`-clean, explicit error returns (no panics in library code), mutex-guarded shared state matching `tasks.Service`'s pattern.
- Svelte: plain JS (no TypeScript in components — `vite-env.d.ts` exists but components are `.svelte` with plain `<script>`), reactive `$:` blocks over manual effects where possible, Wails-bound calls wrapped in try/catch with a local `error` variable (see `TodoWidget.svelte`).
- Indentation: 4 spaces in both Go and Svelte/JS (note: Go files currently have CRLF line endings — be aware when diffing/editing).
- No commit convention enforced yet.

## Things NOT To Do

- Don't hand-edit anything under `frontend/wailsjs/` — it's generated.
- Don't put business logic in `app.go` — it should only ever call into an `internal/<feature>` service.
- Don't have the frontend do date/recurrence math independently of the Go backend — keep recurrence expansion and reminder-time calculation server-side (Go) so behavior is consistent and testable. Concretely: the frontend calls `ListCalendarOccurrences(rangeStart, rangeEnd)`, never `ListEvents()` plus its own recurrence loop.
- Don't add a widget that listens to raw keyboard events outside the `activeWidgetKeyHandler` contract — it will break the global mode-based dispatch in `App.svelte`.

## Open Questions / Decisions Pending

- OS-level push notifications: confirm whether `go-toast` (already an indirect dependency) is already wired up anywhere or just transitively pulled in by Wails and unused.
- Two-way sync with an external calendar (Google/Apple) — in scope for v1 or later?
- Notes: does it need a dedicated full page, or is a widget-opens-to-modal pattern enough? Decide after Calendar/Habits ship and the widget-vs-page tradeoff is clearer in practice.

## Decisions Made

- **Calendar (v1):** both a dashboard widget (next 3-5 upcoming events) AND a full page (month grid view), matching the Tasks feature's widget+page pattern.
- **Calendar v1 excludes:** holiday-awareness/auto-shifting, weather-based reminders, location-based reminders, external calendar sync. Revisit as "someday" items once core scheduling + reminders work.
- **Calendar storage:** a flat `calendar.json` file under `os.UserConfigDir()/focusup/`, matching `tasks.json`'s pattern — not SQLite. Recurrence/exception data nests inside each `Event` (an `Event` owns its own `RecurrenceRule` and `[]Exception`), so it's still a flat list of events, not a relational schema. Revisit only if cross-event queries (e.g. "all exceptions in date range X across all events") become a real need — nothing in v1 requires that.
- **Next widgets after Calendar:** Habit Tracker, then Notes.

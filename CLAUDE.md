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
│   ├── calendar/                  # Calendar feature — see "Calendar file split" below
│   │   ├── event.go               #   Event, Exception, Occurrence, OccurrenceView types
│   │   ├── recurrence.go          #   RecurrenceRule + Occurrences() expansion (pure, tested)
│   │   ├── reminder.go            #   Reminder, DueReminder + lead-time due calculation (pure, tested)
│   │   ├── service.go             #   Service struct, mutex + JSON store, wraps the pure functions above
│   │   └── store.go                #   load/save to calendar.json in os.UserConfigDir()
│   ├── notify/
│   │   └── toast.go               # Push(title, body) — thin wrapper around go-toast (Windows OS toast)
│   └── notifier/
│       └── notifier.go            # Notifier.Run(ctx) — polls calendar.Service.DueReminders, fires each once
├── frontend/
│   ├── src/
│   │   ├── App.svelte             # Root: keyboard state machine, tab routing, layout shell
│   │   ├── lib/
│   │   │   ├── navigation.js       # `tabs` array — top-level tab bar definition
│   │   │   ├── widgets.js          # `widgets` array — dashboard grid definition (id/size/row/col/component)
│   │   │   ├── keyboardGrid.js     # Pure spatial nav helpers (moveSelection, firstWidget) over widgets[]
│   │   │   ├── calendarGrid.js     # Pure spatial nav helpers (buildMonthGrid, moveDayCursor) for the month page
│   │   │   ├── taskDisplay.js      # Presentation helpers for the tasks feature (formatting, grouping)
│   │   │   ├── calendarDisplay.js  # Presentation helpers for the calendar feature (formatting, #each keys)
│   │   │   ├── stores/keyboard.js  # `mode`, `selectedWidgetId`, `activeWidgetKeyHandler` stores — see below
│   │   │   └── components/
│   │   │       ├── Dashboard.svelte    # Renders widgets[] into the grid via WidgetSlot
│   │   │       ├── WidgetSlot.svelte   # Generic widget chrome (header/shortcut/selected state)
│   │   │       ├── TabBar.svelte, MenuBar.svelte
│   │   │       ├── TasksPage.svelte    # Reference: a full page (not just a widget) for a feature
│   │   │       ├── CalendarPage.svelte # Full page: month grid view, routed like TasksPage
│   │   │       ├── ReminderPopup.svelte # Global overlay: listens for the "calendar:reminder-due" event
│   │   │       └── widgets/
│   │   │           ├── TodoWidget.svelte   # Reference: a dashboard-grid widget implementation
│   │   │           └── CalendarWidget.svelte # Next 3-5 upcoming events/reminders
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

**Notifications (decided, see `internal/notify/`, `internal/notifier/`):**
- `internal/notify` is a thin wrapper around `go-toast` (`Push(title, body)`) — see the resolved "OS-level push notifications" question below for why it lives here instead of inline in `notifier`.
- `internal/notifier.Notifier` owns the actual polling: `Run(ctx)` calls `calendar.Service.DueReminders(now)` every 30s, dedups by `eventID|reminderID|occurrenceStart` (pruned after 24h past the occurrence, so a long-running app doesn't leak memory), and for every reminder that's newly due, pushes an OS toast **and** invokes a caller-supplied `DueHandler`.
- `app.go` wires the `DueHandler` in `NewApp()` to `runtime.EventsEmit(a.ctx, "calendar:reminder-due", reminder)`, and starts `notifier.Run(ctx)` as a goroutine in `startup()`. This is the one place `app.go` does more than a one-line method delegation — it's still just wiring (a single `EventsEmit` call in a closure), not business logic; the actual due/dedup/fire policy lives in `internal/notifier`.
- Frontend: `ReminderPopup.svelte` (mounted once, globally, in `App.svelte`) subscribes to the `"calendar:reminder-due"` event via `EventsOn` from `wailsjs/runtime/runtime.js` and renders a dismissible, auto-expiring popup stack. It does not touch `activeWidgetKeyHandler` or listen to `window` keydown — dismissal is click-only, so it can't conflict with the global keyboard dispatch.

**Frontend — the three-mode keyboard state machine:**
`stores/keyboard.js` defines `mode` as one of:
- `"tabs"` — top-level tab bar owns input (h/l cycle tabs, digit keys jump tabs)
- `"grid"` — the active tab's own 2D grid owns input. For the dashboard tab this is the widget grid (h/j/k/l move `selectedWidgetId` via `keyboardGrid.js`); a full page can opt into the same mode for its own grid (see `CalendarPage.svelte` below) — the mode name is intentionally page-agnostic, not `"dashboard"`
- `"widget"` — input is locked to whichever dashboard widget is focused; that widget owns `activeWidgetKeyHandler` until `q` returns to `"grid"`

Any new widget must, like `TodoWidget.svelte`:
- Accept a `focused` prop from `Dashboard.svelte`
- Set `activeWidgetKeyHandler` to its own `handleKey` function only while `focused` is true, and clear it (`set(null)`) otherwise — this is a reactive `$:` block, not a one-time effect
- Never listen to `window` keydown directly — `App.svelte` is the single global listener and dispatches based on `mode`

**Frontend — widget registration:**
New dashboard widgets are added to the `widgets` array in `lib/widgets.js` with `id`, `title`, `shortcut`, `size` (`tall`/`medium`/etc., maps to a CSS grid span), and explicit `row`/`col` (logical grid position used by `keyboardGrid.js` for spatial nav — must match the actual CSS grid placement or navigation will feel wrong). `widget-3` (`col: 8`) is the one remaining open slot with `component: null`; `todo` and `calendar` (both `size: "tall"`) fill the other two.

**Full pages vs. widgets:**
Some features (Tasks, now Calendar) have both a compact dashboard widget AND a dedicated full page reached via the tab bar (`navigation.js` + routing in `App.svelte`). A full page's key handler is set unconditionally in `onMount`/cleared in `onDestroy` (see `TasksPage.svelte`/`CalendarPage.svelte`) rather than gated by a `focused` prop like a dashboard widget — the page owns the whole view whenever its tab is active, there's no separate "focused" state to gate on.

A page with a genuine 2D grid can opt into the dashboard's own tabs↔grid scheme instead of owning its keys unconditionally like `TasksPage.svelte` does (1D lists don't need this — j/k is enough). `CalendarPage.svelte` is the current example: while `mode === "tabs"`, h/l cycle tabs as normal and `k` is a no-op (see the note on why it needs an explicit no-op, below) — pressing `j` (handled in `App.svelte`'s `handleTabsAndPageKey`, alongside the dashboard's own `"j"` case) sets `mode` to `"grid"`, at which point `App.svelte`'s `handleKeyboard` forwards h/j/k/l straight to the page's `activeWidgetKeyHandler` instead of `handleDashboardKey`. Moving `"up"` past the grid's top row (`calendarGrid.js`'s `moveDayCursor` returning `null`, mirroring `keyboardGrid.js`'s `moveSelection`) sets `mode` back to `"tabs"` — same exit shape as the dashboard grid.

One wrinkle full pages have that dashboard widgets don't: a dashboard widget only registers `activeWidgetKeyHandler` once actually focused (`mode === "widget"`), so stray keys before that are simply no-ops. A full page registers its handler unconditionally in `onMount` (see above), so without an explicit guard `k` would leak straight into `CalendarPage`'s handler even before `j` "enters" the grid — `handleTabsAndPageKey`'s `"k"` case exists specifically to prevent that leak for pages using this scheme.

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

- Two-way sync with an external calendar (Google/Apple) — in scope for v1 or later?
- Notes: does it need a dedicated full page, or is a widget-opens-to-modal pattern enough? Decide after Calendar/Habits ship and the widget-vs-page tradeoff is clearer in practice.

## Decisions Made

- **Calendar (v1):** both a dashboard widget AND a full page (month grid view), matching the Tasks feature's widget+page pattern. The widget is a pure *view* (no per-item selection/actions, unlike Tasks) with three modes — Today, This Week, Month — cycled with h/l while focused; it always shows the current today/week/month, it doesn't support navigating to other weeks/months (that's what the full page is for). Weeks start **Monday**, end **Sunday**, everywhere (widget and full page's month grid) — a deliberate, non-default choice, see `calendarGrid.js`.
- **Event.Important:** a plain `bool` on `Event`/`OccurrenceView` for emphasis in compact views (e.g. a red dot vs. a neutral dot on the widget's week/month grid — a birthday vs. a routine garbage-day reminder). As of this writing there's no add/edit-event UI at all (`AddEvent` is Go-binding-only, unreached from any form), so nothing can actually set `Important` to `true` yet outside a hand-edited `calendar.json` — worth flagging for whoever builds that form next.
- **Calendar v1 excludes:** holiday-awareness/auto-shifting, weather-based reminders, location-based reminders, external calendar sync. Revisit as "someday" items once core scheduling + reminders work.
- **Calendar storage:** a flat `calendar.json` file under `os.UserConfigDir()/focusup/`, matching `tasks.json`'s pattern — not SQLite. Recurrence/exception data nests inside each `Event` (an `Event` owns its own `RecurrenceRule` and `[]Exception`), so it's still a flat list of events, not a relational schema. Revisit only if cross-event queries (e.g. "all exceptions in date range X across all events") become a real need — nothing in v1 requires that.
- **Next widgets after Calendar:** Habit Tracker, then Notes.
- **OS-level push notifications:** `go-toast` was an indirect dependency (pulled in transitively by Wails, not imported anywhere) before this work — confirmed by `grep -rn "toast" --include="*.go" .` turning up nothing in our own code. It's now wired up directly in `internal/notify` and used by `internal/notifier`; `go.mod` reflects it as a direct dependency.

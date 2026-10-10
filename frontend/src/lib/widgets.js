import TodoWidget from "./components/widgets/TodoWidget.svelte";
import CalendarWidget from "./components/widgets/CalendarWidget.svelte";
import JobsWidget from "./components/widgets/JobsWidget.svelte";
import HabitsWidget from "./components/widgets/HabitsWidget.svelte";
import HealthWidget from "./components/widgets/HealthWidget.svelte";

// row/col are logical grid positions used for spatial keyboard navigation.
// col matches the actual CSS grid-column start on the 12-column grid
// (tall/medium = span 4, wide = span 6) so navigation lines up with what's
// on screen. Tall and wide widgets both span two grid rows, so each logical
// row here is two CSS rows.
export const widgets = [
    {
        id: "todo",
        title: "To Do List",
        shortcut: "1",
        size: "tall",
        row: 0,
        col: 0,
        component: TodoWidget,
        customHeader: true
    },
    {
        id: "calendar",
        title: "Calendar",
        shortcut: "2",
        size: "tall",
        row: 0,
        col: 4,
        component: CalendarWidget,
        customHeader: true
    },
    {
        id: "jobs",
        title: "Jobs",
        shortcut: "3",
        size: "tall",
        row: 0,
        col: 8,
        component: JobsWidget,
        customHeader: true
    },
    {
        id: "habits",
        title: "Habits",
        shortcut: "4",
        size: "wide",
        row: 1,
        col: 0,
        component: HabitsWidget,
        customHeader: true
    },
    {
        id: "health",
        title: "Health",
        shortcut: "5",
        size: "wide",
        row: 1,
        col: 6,
        component: HealthWidget,
        customHeader: true
    }
];

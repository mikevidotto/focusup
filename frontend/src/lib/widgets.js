import TodoWidget from "./components/widgets/TodoWidget.svelte";

// row/col are logical grid positions used for spatial keyboard navigation.
// col matches the actual CSS grid-column start on the 12-column grid
// (tall/medium = span 4) so navigation lines up with what's on screen.
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
        id: "widget-2",
        title: "Widget Slot",
        shortcut: "2",
        size: "medium",
        row: 0,
        col: 4,
        component: null
    },
    {
        id: "widget-3",
        title: "Widget Slot",
        shortcut: "3",
        size: "medium",
        row: 0,
        col: 8,
        component: null
    }
];

import TodoWidget from "./components/widgets/TodoWidget.svelte";

// row/col are logical grid positions used for spatial keyboard navigation.
// col matches the actual CSS grid-column start on the 12-column grid
// (large = span 6, medium = span 3) so navigation lines up with what's on screen.
export const widgets = [
    {
        id: "todo",
        title: "To Do List",
        shortcut: "1",
        size: "large",
        row: 0,
        col: 0,
        component: TodoWidget
    },
    {
        id: "widget-2",
        title: "Widget Slot",
        shortcut: "2",
        size: "medium",
        row: 0,
        col: 6,
        component: null
    },
    {
        id: "widget-3",
        title: "Widget Slot",
        shortcut: "3",
        size: "medium",
        row: 0,
        col: 9,
        component: null
    }
];

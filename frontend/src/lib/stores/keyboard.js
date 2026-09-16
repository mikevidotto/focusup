import { writable } from "svelte/store";

// "tabs"      - top-level tab bar navigation (h/l cycle tabs, digits jump tabs)
// "dashboard" - grid navigation across widget slots (h/j/k/l move selection)
// "widget"    - keyboard control locked to the selected widget's own handler
export const mode = writable("tabs");

export const selectedWidgetId = writable(null);

// Set by whichever widget is currently focused (mode === "widget").
// Called with the keydown event for every key not already handled globally (e.g. "q").
export const activeWidgetKeyHandler = writable(null);

import { writable } from "svelte/store";

// "tabs"   - top-level tab bar navigation (h/l cycle tabs, digits jump tabs)
// "grid"   - the active tab's own 2D grid owns input: the dashboard's widget
//            grid (h/j/k/l move selectedWidgetId), or a full page's own grid
//            that's opted into this same scheme (e.g. Calendar's month grid,
//            entered via "j" from "tabs", exited via "up" at the top edge).
// "widget" - keyboard control locked to the dashboard's selected widget; that
//            widget owns activeWidgetKeyHandler until "q" returns to "grid"
export const mode = writable("tabs");

export const selectedWidgetId = writable(null);

// Set by whichever widget is currently focused (mode === "widget").
// Called with the keydown event for every key not already handled globally (e.g. "q").
export const activeWidgetKeyHandler = writable(null);

import { writable, get } from "svelte/store";
import { GetSettings, SetTheme } from "../../../wailsjs/go/main/App.js";

// The saved theme lives in the backend's settings.json. localStorage is only
// a fast cache so the first paint already uses the right theme instead of
// flashing dark before GetSettings resolves.
const CACHE_KEY = "focusup:theme";

function readCache() {
    try {
        return localStorage.getItem(CACHE_KEY);
    } catch {
        return null;
    }
}

function writeCache(value) {
    try {
        localStorage.setItem(CACHE_KEY, value);
    } catch {
        // cache is optional
    }
}

function apply(value) {
    document.documentElement.dataset.theme = value;
    writeCache(value);
}

export const theme = writable(readCache() === "light" ? "light" : "dark");

apply(get(theme));

export async function initTheme() {
    try {
        const settings = await GetSettings();
        theme.set(settings.theme);
        apply(settings.theme);
    } catch (e) {
        console.error("failed to load theme", e);
    }
}

export async function toggleTheme() {
    const next = get(theme) === "dark" ? "light" : "dark";

    theme.set(next);
    apply(next);

    try {
        await SetTheme(next);
    } catch (e) {
        console.error("failed to save theme", e);
    }
}

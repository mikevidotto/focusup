// Pure helpers for the Journal page. Each day gets a few prompts picked
// deterministically from PROMPTS (so reopening the tab shows the same ones),
// and an entry stores the prompt text it was answered against, so editing
// this pool never rewrites past entries.

import { toDateKey } from "./habitDisplay.js";

export const PROMPTS_PER_DAY = 3;

const PROMPTS = [
    // gratitude
    "What are three things you're grateful for today?",
    "Who made your day a little better recently?",
    "What's something small you'd miss if it were gone?",
    "What went right today that you didn't plan for?",
    "What part of your routine are you thankful for?",

    // focus & priorities
    "What is the one thing that would make today a win?",
    "What are you avoiding, and why?",
    "What deserves less of your attention right now?",
    "If you could only finish one task this week, which would it be?",
    "What's taking up mental space that you can let go of?",
    "Where did your time actually go today?",
    "What would you do today if you weren't afraid of doing it badly?",

    // reflection
    "What did you learn today?",
    "What's a decision you're glad you made recently?",
    "What would you do differently if you could redo today?",
    "What surprised you this week?",
    "What's a recent win you haven't given yourself credit for?",
    "When did you feel most like yourself today?",
    "What pattern keeps showing up in your days lately?",

    // emotions
    "How are you feeling right now, honestly?",
    "What drained your energy today, and what gave it back?",
    "What's worrying you, and how much of it is in your control?",
    "What would you tell a friend who felt the way you do today?",
    "What made you smile or laugh today?",
    "What do you need more of right now?",
    "What's one thing you can forgive yourself for?",

    // growth & values
    "What kind of person are you trying to become?",
    "Which habit is helping you most, and which is holding you back?",
    "What would future you thank you for doing today?",
    "What's a skill you're building, and how did you practice it?",
    "Where did you step outside your comfort zone recently?",
    "What does a good life look like to you this year?",
    "What are you proud of that nobody else saw?",
    "What's one belief about yourself worth questioning?",

    // tomorrow & intention
    "What's your intention for tomorrow?",
    "What's one thing you can do tomorrow to make it easier?",
    "What are you looking forward to?",
    "What would make this week feel meaningful?",
    "What's the first small step on something you've been putting off?",
    "How do you want to feel at the end of tomorrow?",
];

// A small string hash (FNV-1a) so the same date always seeds the same picks.
function hash(text) {
    let h = 0x811c9dc5;

    for (let i = 0; i < text.length; i++) {
        h ^= text.charCodeAt(i);
        h = Math.imul(h, 0x01000193);
    }

    return h >>> 0;
}

// Picks n distinct prompts for a "YYYY-MM-DD" day. Walks the pool with a
// stride coprime to its length, so every pick is distinct and the starting
// point and spacing both vary day to day.
export function promptsForDate(dateKey, n = PROMPTS_PER_DAY) {
    const seed = hash(dateKey);
    const start = seed % PROMPTS.length;
    let stride = 1 + ((seed >>> 8) % (PROMPTS.length - 1));

    while (gcd(stride, PROMPTS.length) !== 1) {
        stride++;
    }

    return Array.from(
        { length: Math.min(n, PROMPTS.length) },
        (_, i) => PROMPTS[(start + i * stride) % PROMPTS.length],
    );
}

// Returns the prompt that should replace current[index]: the next one in the
// pool (after the one being replaced) that isn't already on the page.
export function swapPrompt(current, index) {
    const from = PROMPTS.indexOf(current[index]);

    for (let step = 1; step <= PROMPTS.length; step++) {
        const candidate = PROMPTS[(from + step + PROMPTS.length) % PROMPTS.length];

        if (!current.includes(candidate)) {
            return candidate;
        }
    }

    return current[index];
}

function gcd(a, b) {
    return b === 0 ? a : gcd(b, a % b);
}

function countWords(text) {
    return (text ?? "").split(/\s+/).filter(Boolean).length;
}

export function wordCount(entry) {
    if (!entry) {
        return 0;
    }

    const answers = (entry.prompts ?? []).reduce(
        (sum, p) => sum + countWords(p.answer),
        0,
    );

    return answers + countWords(entry.body);
}

// Consecutive days with an entry, ending today — or yesterday, so the streak
// isn't shown as broken before you've had a chance to write today.
export function currentStreak(entries, today) {
    const dates = new Set(entries.map((e) => e.date));
    const day = new Date(today);

    if (!dates.has(toDateKey(day))) {
        day.setDate(day.getDate() - 1);
    }

    let streak = 0;

    while (dates.has(toDateKey(day))) {
        streak++;
        day.setDate(day.getDate() - 1);
    }

    return streak;
}

// The first non-empty line written in an entry, for the recent-entries list.
export function entryPreview(entry) {
    const texts = [...(entry.prompts ?? []).map((p) => p.answer), entry.body];
    const first = texts.find((t) => t && t.trim());

    return first ? first.trim().split("\n")[0] : "";
}

// "Thursday, October 8"
export function formatEntryDate(date) {
    return date.toLocaleDateString("en-US", {
        weekday: "long",
        month: "long",
        day: "numeric",
    });
}

// Parses a "YYYY-MM-DD" key as a local date (new Date(key) would be UTC).
export function fromDateKey(key) {
    const [y, m, d] = key.split("-").map(Number);

    return new Date(y, m - 1, d);
}

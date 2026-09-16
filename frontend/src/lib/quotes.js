const QUOTES = [
    "Consistency builds strength",
    "Small steps, every day",
    "Progress over perfection",
    "Discipline is choosing what you want most",
    "Focus on what matters",
    "Done is better than perfect",
    "Show up, especially when it's hard"
];

export function quoteForToday() {
    const dayIndex = Math.floor(Date.now() / 86400000);
    return QUOTES[dayIndex % QUOTES.length];
}

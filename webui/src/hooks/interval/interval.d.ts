/**
 * Declarative `setInterval` hook: calls `callback` every `delay` milliseconds,
 * always using the most recent callback. A `null` delay stops the timer.
 *
 * @param callback function to call on each tick.
 * @param delay interval in milliseconds, or `null` to pause.
 */
declare const useInterval: (callback: () => void, delay: number | null) => void
export default useInterval

// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

/**
 * A filter pattern together with its matching options.
 */
export interface FilterPattern {
    /** substring or regular expression to match. */
    pattern: string
    /** match case-sensitive instead of case-insensitive. */
    isCaseSensitive: boolean
    /** interpret the pattern as a regular expression instead of a substring. */
    isRegexp: boolean
}

/**
 * Returns a predicate function that tests strings against the specified filter
 * pattern: either as a (case-insensitive) substring match, or as a regular
 * expression match. An invalid regular expression never matches.
 *
 * @param fp filter pattern including its matching options.
 */
export const getFilterFn = (fp: FilterPattern) => {
    if (!fp.isRegexp) {
        if (fp.isCaseSensitive) {
            return (s: string) => s.includes(fp.pattern)
        }
        const filter = fp.pattern.toLocaleLowerCase()
        return (s: string) => s.toLocaleLowerCase().includes(filter)
    }
    try {
        const re = new RegExp(fp.pattern, fp.isCaseSensitive ? "" : "i")
        return (s: string) => re.test(s)
    } catch {
        return () => false
    }
}
